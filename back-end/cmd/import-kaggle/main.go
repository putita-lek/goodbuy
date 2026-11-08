// Command import-kaggle loads supermarket prices from the Kaggle dataset
// "Time Series UK Supermarket Data" (All_Data_<Chain>.csv files).
//
// Each file belongs to one retailer. Each row becomes a product of that retailer
// and its chain-wide price in product_prices (the dataset has no per-store prices).
//
// Re-running the command replaces the retailer's previously imported products
// and prices; hand-entered products (source "") and other retailers are left
// alone. Run the server once first (or the seed) so the retailers exist.
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
	"gorm.io/gorm/clause"
)

const source = "kaggle"

// chain links one CSV file to the retailer (by exact name) it is imported into.
// Sainsbury's Local is deliberately not listed: it gets no imported products.
type chain struct {
	file     string
	retailer string
}

var chains = []chain{
	{"All_Data_Sains.csv", "Sainsbury's"},
	{"All_Data_Tesco.csv", "Tesco"},
	{"All_Data_Aldi.csv", "Aldi"},
	{"All_Data_ASDA.csv", "ASDA"},
	{"All_Data_Morrisons.csv", "Morrisons"},
}

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
	batch := flag.Int("batch", 1000, "rows per INSERT")
	dryRun := flag.Bool("dry-run", false, "read and summarise the files without writing to the database")
	flag.Parse()

	if *date != "" {
		if _, err := time.Parse("20060102", *date); err != nil {
			log.Fatalf("-date must be YYYYMMDD: %v", err)
		}
	}

	var db *gorm.DB
	if !*dryRun {
		var err error
		if db, err = database.Connect(config.Load().DSN()); err != nil {
			log.Fatal(err)
		}
		if err := database.Migrate(db); err != nil {
			log.Fatalf("migrate: %v", err)
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

		var ret models.Retailer
		err = db.Where("name = ?", c.retailer).First(&ret).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("%s: no retailer named %q, nothing imported (seed retailers first)", c.file, c.retailer)
			continue
		}
		if err != nil {
			log.Fatalf("load retailer %q: %v", c.retailer, err)
		}

		n, err := importRetailer(db, ret, rows, *batch)
		if err != nil {
			log.Fatalf("import into %q: %v", c.retailer, err)
		}
		log.Printf("  -> %s: %d products with prices", c.retailer, n)
		total += n
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

		name := truncate(strings.Join(strings.Fields(rec[col["names"]]), " "), 255)
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

// importRetailer replaces the retailer's imported products and their prices in one
// transaction, and returns the number of products written.
func importRetailer(db *gorm.DB, ret models.Retailer, rows map[string]row, batch int) (int, error) {
	keys := make([]string, 0, len(rows))
	for k := range rows {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	products := make([]models.Product, 0, len(rows))
	for _, k := range keys {
		r := rows[k]
		products = append(products, models.Product{
			RetailerID: ret.ID,
			Name:       r.name,
			Keywords:   truncate(r.keywords, 512),
			IsOrganic:  strings.Contains(strings.ToLower(r.name), "organic"),
			IsOwnBrand: r.ownBrand,
			Source:     source,
		})
	}

	n := 0
	err := db.Transaction(func(tx *gorm.DB) error {
		imported := tx.Model(&models.Product{}).Select("id").Where("retailer_id = ? AND source = ?", ret.ID, source)
		if err := tx.Where("product_id IN (?)", imported).Delete(&models.ProductPrice{}).Error; err != nil {
			return err
		}
		if err := tx.Where("retailer_id = ? AND source = ?", ret.ID, source).Delete(&models.Product{}).Error; err != nil {
			return err
		}

		// Names unique to Go can still clash under MySQL's case- and accent-insensitive
		// collation (e.g. "Café" and "Cafe"), or with a hand-entered product: INSERT IGNORE
		// skips those. (OnConflict{DoNothing} becomes "UPDATE id=id", which MySQL rejects.)
		if err := tx.Omit(clause.Associations).Clauses(clause.Insert{Modifier: "IGNORE"}).
			CreateInBatches(products, batch).Error; err != nil {
			return err
		}

		// Read the ids back by name rather than trusting ids filled in by a batch insert.
		var saved []models.Product
		if err := tx.Select("id", "name").Where("retailer_id = ? AND source = ?", ret.ID, source).
			Find(&saved).Error; err != nil {
			return err
		}

		prices := make([]models.ProductPrice, 0, len(saved))
		for _, p := range saved {
			r, ok := rows[strings.ToLower(p.Name)]
			if !ok {
				continue
			}
			day := r.date
			prices = append(prices, models.ProductPrice{
				ProductID: p.ID, PriceGBP: r.price,
				UnitPriceGBP: r.unitPrice, Unit: r.unit, PriceDate: &day,
			})
		}
		if err := tx.CreateInBatches(prices, batch).Error; err != nil {
			return err
		}
		n = len(prices)
		return nil
	})
	return n, err
}

// truncate shortens s to at most n characters without splitting a UTF-8 character.
func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}
