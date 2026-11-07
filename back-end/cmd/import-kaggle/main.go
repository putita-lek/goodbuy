// Command import-kaggle loads supermarket prices from the Kaggle dataset
// "Time Series UK Supermarket Data" (All_Data_<Chain>.csv files) into the
// products table.
//
// The dataset has prices per chain, not per store, so each chain's latest
// prices are copied to every existing store whose name starts with that
// chain's name (for example "Sainsbury's Cameron Toll" gets the Sainsbury's
// prices). Re-running the command replaces previously imported rows; seed
// products and other stores are left alone.
package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/putita-lek/goodbuy-backend/internal/config"
	"github.com/putita-lek/goodbuy-backend/internal/database"
	"github.com/putita-lek/goodbuy-backend/internal/models"
	"gorm.io/gorm"
)

const source = "kaggle"

// chain describes one CSV file and the store-name prefixes it applies to.
type chain struct {
	file     string
	prefixes []string
}

var chains = []chain{
	{"All_Data_Sains.csv", []string{"Sainsbury's", "Sainsburys"}},
	{"All_Data_Tesco.csv", []string{"Tesco"}},
	{"All_Data_Aldi.csv", []string{"Aldi"}},
	{"All_Data_ASDA.csv", []string{"ASDA"}},
	{"All_Data_Morrisons.csv", []string{"Morrisons"}},
}

// Store names containing these words are convenience stores, which usually
// charge more than the chain's main (online) prices.
var convenienceWords = []string{"local", "express", "metro"}

type row struct {
	name, unit, keywords string
	price                float64
	unitPrice            *float64
	ownBrand             bool
	date                 time.Time
}

func main() {
	dir := flag.String("dir", "../kaggle", "folder containing the All_Data_*.csv files")
	date := flag.String("date", "", "snapshot day to import as YYYYMMDD (default: latest day in each file)")
	markup := flag.Float64("convenience-markup", 0, "percent added to prices for Local/Express/Metro stores, e.g. 8")
	batch := flag.Int("batch", 1000, "rows per INSERT")
	dryRun := flag.Bool("dry-run", false, "read and summarise the files without writing to the database")
	flag.Parse()

	if *date != "" {
		if _, err := time.Parse("20060102", *date); err != nil {
			log.Fatalf("-date must be YYYYMMDD: %v", err)
		}
	}

	var db *gorm.DB
	var stores []models.Store
	if !*dryRun {
		var err error
		if db, err = database.Connect(config.Load().DSN()); err != nil {
			log.Fatal(err)
		}
		if err := database.Migrate(db); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		if err := db.Find(&stores).Error; err != nil {
			log.Fatalf("load stores: %v", err)
		}
	}

	total := 0
	for _, c := range chains {
		path := filepath.Join(*dir, c.file)
		rows, day, skipped, err := readLatest(path, *date)
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("%s: not found, skipping", path)
			continue
		}
		if err != nil {
			log.Fatalf("%s: %v", path, err)
		}
		log.Printf("%s: %d products on %s (%d bad rows skipped)", c.file, len(rows), day, skipped)
		if *dryRun {
			continue
		}

		targets := matchStores(stores, c.prefixes)
		if len(targets) == 0 {
			log.Printf("%s: no store named %q..., nothing imported (add a store for this chain first)",
				c.file, c.prefixes[0])
			continue
		}
		for _, st := range targets {
			factor := 1.0
			if isConvenience(st.Name) {
				factor += *markup / 100
			}
			if err := replaceProducts(db, st, rows, factor, *batch); err != nil {
				log.Fatalf("import into %q: %v", st.Name, err)
			}
			log.Printf("  -> %s: %d products", st.Name, len(rows))
			total += len(rows)
		}
	}
	if !*dryRun {
		log.Printf("done: %d products imported", total)
	}
}

// readLatest returns one row per product name for the requested day, or for
// the latest day in the file when want is empty.
func readLatest(path, want string) (map[string]row, string, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", 0, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.ReuseRecord = true
	r.FieldsPerRecord = -1 // a malformed row is skipped below rather than stopping the import
	header, err := r.Read()
	if err != nil {
		return nil, "", 0, fmt.Errorf("read header: %w", err)
	}
	col := map[string]int{}
	for i, h := range header {
		col[strings.TrimPrefix(strings.TrimSpace(h), string(rune(0xFEFF)))] = i // drop a UTF-8 byte order mark
	}
	need := []string{"prices_(£)", "prices_unit_(£)", "unit", "names", "date", "category", "own_brand"}
	for _, n := range need {
		if _, ok := col[n]; !ok {
			return nil, "", 0, fmt.Errorf("missing column %q", n)
		}
	}

	day := want
	rows := map[string]row{}
	skipped := 0
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(rec) < len(header) {
			skipped++
			continue
		}
		d := strings.TrimSpace(rec[col["date"]])
		if want != "" && d != want {
			continue
		}
		if want == "" {
			if d < day {
				continue
			}
			if d > day { // a newer day: start again from this day
				day = d
				rows = map[string]row{}
			}
		}

		name := strings.Join(strings.Fields(rec[col["names"]]), " ")
		price, err := strconv.ParseFloat(strings.TrimSpace(rec[col["prices_(£)"]]), 64)
		if name == "" || err != nil || price <= 0 {
			skipped++
			continue
		}
		category := strings.NewReplacer("_", " ", "-", " ").Replace(strings.TrimSpace(rec[col["category"]]))

		key := strings.ToLower(name)
		if existing, ok := rows[key]; ok {
			// The same product can be listed under several categories: keep
			// the first price and add the extra category as a keyword.
			if category != "" && !strings.Contains(existing.keywords, category) {
				existing.keywords = strings.TrimSpace(existing.keywords + " " + category)
				rows[key] = existing
			}
			continue
		}

		var unitPrice *float64
		if up, err := strconv.ParseFloat(strings.TrimSpace(rec[col["prices_unit_(£)"]]), 64); err == nil && up > 0 {
			unitPrice = &up
		}
		parsedDay, err := time.Parse("20060102", d)
		if err != nil {
			skipped++
			continue
		}
		rows[key] = row{
			name:      name,
			price:     price,
			unitPrice: unitPrice,
			unit:      truncate(strings.ToLower(strings.TrimSpace(rec[col["unit"]])), 16),
			keywords:  category,
			ownBrand:  strings.EqualFold(strings.TrimSpace(rec[col["own_brand"]]), "true"),
			date:      parsedDay,
		}
	}
	if len(rows) == 0 {
		return nil, "", skipped, fmt.Errorf("no rows found for date %q", want)
	}
	return rows, day, skipped, nil
}

func matchStores(stores []models.Store, prefixes []string) []models.Store {
	var out []models.Store
	for _, s := range stores {
		for _, p := range prefixes {
			if strings.HasPrefix(strings.ToLower(s.Name), strings.ToLower(p)) {
				out = append(out, s)
				break
			}
		}
	}
	return out
}

func isConvenience(storeName string) bool {
	for _, w := range strings.Fields(strings.ToLower(storeName)) {
		for _, c := range convenienceWords {
			if w == c {
				return true
			}
		}
	}
	return false
}

// replaceProducts deletes the store's previously imported products and
// inserts the new snapshot in one transaction.
func replaceProducts(db *gorm.DB, st models.Store, rows map[string]row, factor float64, batch int) error {
	keys := make([]string, 0, len(rows))
	for k := range rows {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	products := make([]models.Product, 0, len(rows))
	for _, k := range keys {
		r := rows[k]
		day := r.date
		p := models.Product{
			Name:       truncate(r.name, 255),
			Keywords:   truncate(r.keywords, 512),
			PriceGBP:   round2(r.price * factor),
			StoreID:    st.ID,
			IsOrganic:  strings.Contains(strings.ToLower(r.name), "organic"),
			IsOwnBrand: r.ownBrand,
			Unit:       r.unit,
			PriceDate:  &day,
			Source:     source,
		}
		if r.unitPrice != nil {
			up := round2(*r.unitPrice * factor)
			p.UnitPriceGBP = &up
		}
		products = append(products, p)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("store_id = ? AND source = ?", st.ID, source).Delete(&models.Product{}).Error; err != nil {
			return err
		}
		return tx.Omit("Store").CreateInBatches(products, batch).Error
	})
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}

// truncate shortens s to at most n characters without splitting a UTF-8 character.
func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}
