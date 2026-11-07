package seed

import (
	"log"

	"github.com/putita-lek/goodbuy-backend/internal/models"
	"gorm.io/gorm"
)

type product struct {
	name, keywords string
	price          float64
	recyclable     bool
	organic        bool
}

type store struct {
	name     string
	lat, lng float64
	local    bool
	products []product
}

// Sample catalogue around central Edinburgh (roughly 55.95, -3.19). Prices are illustrative.
var catalogue = []store{
	{"Sainsbury's Cameron Toll", 55.9311, -3.1621, false, []product{
		{"Diet Coke 330ml can", "diet coke small can soda", 1.39, true, false},
		{"Diet Coke 2L bottle", "diet coke large bottle soda", 2.25, true, false},
		{"Apple (Red Delicious) 500g", "red apple fruit", 1.59, false, false},
		{"Semi-skimmed milk 1L carton", "milk carton dairy", 1.15, true, false},
		{"Rice flour 500g pack", "rice flour pack baking gluten free", 1.80, false, false},
		{"Jasmine rice 1kg", "jasmine rice kg thai", 2.40, false, false},
	}},
	{"Sainsbury's Local Princes Street", 55.9520, -3.1960, false, []product{
		{"Diet Coke 330ml can", "diet coke small can soda", 1.50, true, false},
		{"Semi-skimmed milk 1L carton", "milk carton dairy", 1.25, true, false},
		{"Gala apple", "red apple fruit", 0.45, false, false},
	}},
	{"Lidl Leith", 55.9707, -3.1700, false, []product{
		{"Diet Coke 330ml can", "diet coke small can soda", 1.35, true, false},
		{"Gala apple 6 pack", "red apple fruit", 1.57, false, false},
		{"Whole milk 1L carton", "milk carton dairy", 0.99, true, false},
		{"Jasmine rice 1kg", "jasmine rice kg thai", 1.99, false, false},
		{"Rice flour 1kg pack", "rice flour pack baking", 2.10, false, false},
	}},
	{"Edinburgh Sunday Farmers' Market", 55.9466, -3.2045, true, []product{
		{"Red apple (organic, per kg)", "red apple fruit organic", 2.00, false, true},
		{"Organic whole milk 1L glass bottle", "milk bottle dairy organic", 1.90, true, true},
	}},
	{"Real Foods Broughton Street", 55.9580, -3.1880, true, []product{
		{"Organic jasmine rice 1kg", "jasmine rice kg organic", 3.75, true, true},
		{"Organic rice flour 500g pack", "rice flour pack organic gluten free", 2.95, true, true},
		{"Organic oat milk 1L carton", "milk carton oat plant", 1.85, true, true},
	}},
	{"Tesco Express Morningside", 55.9290, -3.2100, false, []product{
		{"Diet Coke 330ml can", "diet coke small can soda", 1.45, true, false},
		{"Semi-skimmed milk 2L carton", "milk carton dairy", 1.65, true, false},
		{"Jasmine rice 500g", "jasmine rice thai", 1.40, false, false},
	}},
}

// Run inserts the sample catalogue when the stores table is empty.
func Run(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.Store{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		n := 0
		for _, s := range catalogue {
			st := models.Store{Name: s.name, Latitude: s.lat, Longitude: s.lng, IsLocalBusiness: s.local}
			for _, p := range s.products {
				st.Products = append(st.Products, models.Product{
					Name: p.name, Keywords: p.keywords, PriceGBP: p.price,
					IsRecyclable: p.recyclable, IsOrganic: p.organic,
				})
				n++
			}
			if err := tx.Create(&st).Error; err != nil {
				return err
			}
		}
		log.Printf("seeded %d stores and %d products", len(catalogue), n)
		return nil
	})
}
