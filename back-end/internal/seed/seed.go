package seed

import (
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/putita-lek/goodbuy-backend/internal/models"
	"gorm.io/gorm"
)

// store is one branch. branch is shown after the retailer name, e.g. "Tesco" + "Express Nicolson Street".
type store struct {
	branch, format    string
	address, postcode string
	lat, lng          float64
}

type retailer struct {
	name   string
	local  bool
	stores []store
}

// Retailers and their branches in the City of Edinburgh council area.
//
// Source: OpenStreetMap (© OpenStreetMap contributors, ODbL), queried via the
// Overpass API in November 2026 for shops tagged with each chain's brand.
// Branches with no address in OSM were reverse-geocoded with Nominatim.
// Branch names are the street name unless the store is widely known by
// another name (e.g. Cameron Toll). Check before relying on this list:
// stores open and close, and OSM can be incomplete.
//
// Products and prices are not seeded; they are loaded separately.
var retailers = []retailer{
	{
		name: "Sainsbury's", local: false,
		stores: []store{
			{"Cameron Toll", "supermarket", "6 Lady Road", "EH16 5PB", 55.92698, -3.1633},
			{"Craigleith", "supermarket", "185 Craigleith Road", "EH4 2EB", 55.95746, -3.24083},
			{"Inglis Green Road", "supermarket", "22 Inglis Green Road", "EH14 2ER", 55.92409, -3.25083},
			{"Meadowbank", "supermarket", "11 Moray Park", "EH7 5TS", 55.95878, -3.16436},
			{"Westfield Road", "supermarket", "39 Westfield Road", "EH11 2QW", 55.93763, -3.23938},
		},
	},
	{
		name: "Sainsbury's Local", local: false,
		stores: []store{
			{"Bernard Street", "local", "41-45 Bernard Street", "EH6 6SL", 55.97592, -3.16889},
			{"Biggar Road", "local", "44 Biggar Road", "EH10 7BJ", 55.89723, -3.20409},
			{"Bruntsfield Place", "local", "155-159 Bruntsfield Place", "EH10 4DG", 55.93593, -3.209},
			{"Causewayside", "local", "80 Causewayside", "EH9 1PY", 55.938, -3.18141},
			{"Craighall Road", "local", "83 Craighall Road", "EH6 4RR", 55.97678, -3.19648},
			{"Deanhaugh Street", "local", "35 Deanhaugh Street", "EH4 1LR", 55.95863, -3.20947},
			{"Dundee Street", "local", "81-85 Dundee Street", "EH11 1AW", 55.94062, -3.2169},
			{"Earl Grey Street", "local", "6-14 Earl Grey Street", "EH3 9BN", 55.94392, -3.20424},
			{"Howe Street", "local", "28 Howe Street", "EH3 6TG", 55.95643, -3.20264},
			{"Jubilee Road", "local", "Jubilee Road", "EH12 9DN", 55.94845, -3.36282},
			{"Lanark Road West", "local", "57 Lanark Road West", "EH14 5JX", 55.89918, -3.29756},
			{"Leith Walk", "local", "Leith Walk", "EH7 4PA", 55.96272, -3.17936},
			{"Marchmont Road", "local", "102,104 Marchmont Road", "EH9 1BG", 55.93613, -3.19466},
			{"Morningside Road", "local", "358-364 Morningside Road", "EH10 4QN", 55.92668, -3.20942},
			{"Morrison Street", "local", "3 Morrison Street", "EH3 8BH", 55.94566, -3.20637},
			{"Nicolson Street", "local", "52 Nicolson Street", "EH8 9DT", 55.94568, -3.1845},
			{"Portobello High Street", "local", "222-224-226-228 Portobello High Street", "EH15 2AU", 55.95226, -3.11194},
			{"Princes Street", "local", "18-19 Princes Street", "EH2 2AN", 55.95323, -3.19072},
			{"Shandwick Place", "local", "32-36 Shandwick Place", "EH2 4RT", 55.94973, -3.20957},
			{"Simpson Loan", "local", "29-30 Simpson Loan", "EH3 9GG", 55.94387, -3.19187},
			{"South Bridge", "local", "79 South Bridge", "EH1 1HN", 55.9486, -3.18699},
			{"South Clerk Street", "local", "31-37 South Clerk Street", "EH8 9NZ", 55.94082, -3.18058},
			{"Waverley Steps", "local", "Waverley Steps", "EH1 1BQ", 55.95258, -3.19024},
			{"West Port", "local", "102 West Port", "EH3 9DN", 55.94592, -3.20106},
		},
	},
	{
		name: "Lidl", local: false,
		stores: []store{
			{"Dalry Road", "supermarket", "201-207 Dalry Road", "EH11 2EF", 55.94141, -3.22194},
			{"Easter Road", "supermarket", "234-248 Easter Road", "EH6 8LE", 55.96317, -3.17027},
			{"Glasgow Road", "supermarket", "1-17 Glasgow Road", "EH12 8HW", 55.94216, -3.29416},
			{"Lawrie Reilly Place", "supermarket", "Lawrie Reilly Place", "EH7 5FA", 55.95813, -3.16608},
			{"Logie Green Road", "supermarket", "11 Logie Green Road", "EH7 4EY", 55.96398, -3.19533},
			{"Murrayburn Gate", "supermarket", "Murrayburn Gate", "EH14 2RR", 55.91522, -3.28423},
			{"Newkirkgate", "supermarket", "35 Newkirkgate", "EH6 6AA", 55.97142, -3.1722},
			{"Nicolson Street", "supermarket", "56-58 Nicolson Street", "EH8 9DT", 55.94558, -3.18437},
			{"Niddrie Mains Road", "supermarket", "201 Niddrie Mains Road", "EH16 4PA", 55.93314, -3.13289},
			{"Seafield Road", "supermarket", "64 Seafield Road", "EH6 7LW", 55.96811, -3.13997},
			{"Slateford Road", "supermarket", "155 Slateford Road", "EH14 1NZ", 55.9312, -3.23538},
			{"South Groathill Avenue", "supermarket", "52 South Groathill Avenue", "EH4 2LN", 55.95871, -3.24331},
			{"West Granton Road", "supermarket", "162 West Granton Road", "EH5 1PE", 55.97901, -3.23424},
		},
	},
	{
		name: "Edinburgh Farmers' Market", local: true,
		stores: []store{
			{"Castle Terrace", "market", "Castle Terrace", "", 55.94776, -3.20348}, // Saturdays 09:00-14:00
		},
	},
	{
		name: "Real Foods", local: true,
		stores: []store{
			{"Broughton Street", "shop", "37 Broughton Street", "EH1 3JU", 55.95728, -3.18813},
			{"Tollcross", "shop", "8 Brougham Street", "EH3 9JH", 55.9436, -3.20268},
		},
	},
	{
		name: "Tesco", local: false,
		stores: []store{
			{"Broughton Road", "supermarket", "7 Broughton Road", "EH7 4EW", 55.9633, -3.1963},
			{"Colinton Mains Drive", "supermarket", "10-12 Colinton Mains Drive", "EH13 9AH", 55.91149, -3.23873},
			{"Cramond Road South", "supermarket", "36 Cramond Road South", "EH4 6AA", 55.96616, -3.27415},
			{"Cultins Road", "supermarket", "Cultins Road", "EH11 4DG", 55.9247, -3.30549},
			{"Duke Street", "supermarket", "76 Duke Street", "EH6 8HL", 55.96978, -3.17035},
			{"Express 2A Colinton Road", "express", "2A Colinton Road", "EH10 5DN", 55.93418, -3.21073},
			{"Express 62B Colinton Road", "express", "62B Colinton Road", "EH14 1AT", 55.92684, -3.23264},
			{"Express Brunswick Place", "express", "4-6 Brunswick Place", "EH7 5HW", 55.9613, -3.18035},
			{"Express Bughtlin Market", "express", "1 Bughtlin Market", "EH12 8XP", 55.95186, -3.30455},
			{"Express Causewayside", "express", "158 Causewayside", "EH9 1PR", 55.93602, -3.18028},
			{"Express Dundas Street", "express", "30 Dundas Street", "EH3 6JN", 55.9573, -3.19963},
			{"Express Earl Grey Street", "express", "32 Earl Grey Street", "EH3 9BN", 55.94429, -3.20478},
			{"Express Fountainbridge", "express", "135-137 Fountainbridge", "EH3 9QG", 55.94306, -3.20924},
			{"Express Gracemount Drive", "express", "2 Gracemount Drive", "EH16 6RS", 55.90246, -3.15492},
			{"Express Great Junction Street", "express", "110 Great Junction Street", "EH6 5LH", 55.97326, -3.1758},
			{"Express Haymarket Square", "express", "1 Haymarket Square", "EH3 8RY", 55.94598, -3.21572},
			{"Express Haymarket Terrace", "express", "76-82 Haymarket Terrace", "EH12 5LQ", 55.94605, -3.22245},
			{"Express Holyrood Road", "express", "59 Holyrood Road", "EH8 8AU", 55.95032, -3.17827},
			{"Express Leith Walk", "express", "110-120 Leith Walk", "EH6 5DT", 55.96865, -3.17379},
			{"Express Nicolson Street", "express", "94 Nicolson Street", "EH8 9EW", 55.94478, -3.18355},
			{"Express Niddrie Mains Road", "express", "82 Niddrie Mains Road", "EH16 4DT", 55.9327, -3.13688},
			{"Express Ocean Drive", "express", "Ocean Drive", "EH6 6JJ", 55.9813, -3.17668},
			{"Express Picardy Place", "express", "8 Picardy Place", "EH1 3JT", 55.95699, -3.18723},
			{"Express Princes Street", "express", "141 Princes Street", "EH2 4BL", 55.95046, -3.20695},
			{"Express Queensferry Road", "express", "503 Queensferry Road", "EH4 7QD", 55.96012, -3.29539},
			{"Express Russell Road", "express", "1B Russell Road", "EH12 5NG", 55.94532, -3.23341},
			{"Express South Bridge", "express", "60-62 South Bridge", "EH1 1LS", 55.94758, -3.18602},
			{"Express Stenhouse Place East", "express", "74/1 Stenhouse Place East", "EH11 3DQ", 55.93324, -3.26074},
			{"Express Willowbrae Road", "express", "226 Willowbrae Road", "EH8 7NG", 55.94721, -3.13748},
			{"Extra Meadow Place Road", "extra", "30 Meadow Place Road", "EH12 7UQ", 55.94054, -3.29516},
			{"Ferrymuir", "supermarket", "5 Ferrymuir", "EH30 9QZ", 55.98336, -3.40146},
		},
	},
	{
		name: "Aldi", local: false,
		stores: []store{
			{"Cameron Toll", "supermarket", "6 Lady Road", "EH16 5PB", 55.92639, -3.16411},
			{"Chesser Avenue", "supermarket", "Chesser Avenue", "EH14 1UD", 55.92905, -3.24769},
			{"Dock Street", "supermarket", "6 Dock Street", "EH6 6EY", 55.9767, -3.17305},
			{"Gilmerton Road", "supermarket", "376 Gilmerton Road", "EH17 7QS", 55.91278, -3.14665},
			{"Gorgie Park Road", "supermarket", "100 Gorgie Park Road", "EH11 2QL", 55.93641, -3.2384},
			{"Oxgangs Road North", "supermarket", "345 Oxgangs Road North", "EH13 9LY", 55.90482, -3.2225},
			{"Portobello High Street", "supermarket", "3-15 Portobello High Street", "EH15 1DW", 55.95574, -3.12026},
			{"Wester Hailes Centre", "supermarket", "Wester Hailes Centre", "EH11 4DF", 55.92594, -3.30602},
		},
	},
	{
		name: "ASDA", local: false,
		stores: []store{
			{"New Mart Road", "supermarket", "3 New Mart Road", "EH14 1RJ", 55.92599, -3.24655},
			{"Sandpiper Drive", "supermarket", "2 Sandpiper Drive", "EH6 6QJ", 55.98156, -3.18967},
			{"The Jewel", "supermarket", "100 The Jewel", "EH15 3AR", 55.9387, -3.10311},
		},
	},
	{
		name: "Morrisons", local: false,
		stores: []store{
			{"Daily Craigentinny Avenue", "daily", "2C,2D,2E Craigentinny Avenue", "EH7 6PX", 55.95635, -3.13812},
			{"Daily Ferry Road", "daily", "663-667 Ferry Road", "EH4 2TX", 55.96727, -3.24549},
			{"Daily Gilmerton Road", "daily", "Gilmerton Road", "EH17 7JQ", 55.90983, -3.14251},
			{"Daily Macmillan Square", "daily", "8 Macmillan Square", "EH4 4AB", 55.97232, -3.25312},
			{"Daily Restalrig Road", "daily", "59-63 Restalrig Road", "EH6 8BE", 55.96817, -3.15831},
			{"Daily South Gyle Broadway", "daily", "South Gyle Broadway", "EH12 9JU", 55.93647, -3.31361},
			{"Daily Waterfront Broadway", "daily", "3 Waterfront Broadway", "EH5 1SA", 55.97766, -3.24099},
			{"Ferry Road", "supermarket", "102 Pilton Drive", "EH5 2XS", 55.97013, -3.232},
			{"Gilmerton Road", "supermarket", "406 Gilmerton Road", "EH17 7JQ", 55.90901, -3.14267},
			{"Gyle", "supermarket", "47 Gyle Avenue", "EH12 9JU", 55.9393, -3.31659},
			{"New Swanston", "supermarket", "30 New Swanston", "EH10 7JA", 55.90165, -3.22226},
			{"Piersfield Terrace", "supermarket", "4 Piersfield Terrace", "EH8 7EF", 55.95429, -3.14179},
			{"Waterfront Broadway", "supermarket", "3 Waterfront Broadway", "EH5 1SA", 55.97824, -3.24332},
		},
	},
}

// mockProduct is a made-up product with a made-up price, for retailers that the
// Kaggle dataset does not cover. unitPrice is per unit (kg, l or unit).
type mockProduct struct {
	name, keywords   string
	price, unitPrice float64
	unit             string
	recyclable       bool // packaging can be recycled; false for loose produce
	organic          bool
}

const mockSource = "mock"

// mockProducts are illustrative only: names, prices and attributes are invented,
// not taken from the shops. Keyed by retailer name.
var mockProducts = map[string][]mockProduct{
	"Real Foods": {
		{"Organic Jasmine Rice 1kg", "rice jasmine thai grain wholefood", 3.75, 3.75, "kg", true, true},
		{"Organic Brown Rice Flour 500g", "rice flour baking gluten free", 2.95, 5.90, "kg", true, true},
		{"Organic Oat Milk 1L", "oat milk plant dairy free carton", 1.85, 1.85, "l", true, true},
		{"Organic Semi Skimmed Milk 1L", "milk semi skimmed dairy carton", 1.65, 1.65, "l", true, true},
		{"Organic Gala Apples Loose", "apple fruit gala red", 3.20, 3.20, "kg", false, true},
		{"Organic Porridge Oats 1kg", "oats porridge oat cereal breakfast", 2.60, 2.60, "kg", true, true},
		{"Organic Red Lentils 500g", "lentil red pulse dal wholefood", 2.20, 4.40, "kg", true, true},
		{"Organic Extra Virgin Olive Oil 500ml", "olive oil cooking glass bottle", 7.50, 15.00, "l", true, true},
		{"Organic Smooth Peanut Butter 340g", "peanut butter spread nut glass jar", 3.95, 11.62, "kg", true, true},
		{"Organic Free Range Eggs 6 Pack", "egg free range box", 3.40, 0.57, "unit", true, true},
	},
	"Edinburgh Farmers' Market": {
		{"Free Range Eggs 12 Pack", "egg free range box", 4.00, 0.33, "unit", true, false},
		{"Organic Whole Milk 1L Glass Bottle", "milk whole dairy glass bottle", 1.90, 1.90, "l", true, true},
		{"Sourdough Loaf 800g", "bread sourdough loaf bakery", 4.50, 5.63, "kg", true, false},
		{"Scottish Heather Honey 340g", "honey heather scottish glass jar", 7.50, 22.06, "kg", true, false},
		{"Red Apples Loose", "apple fruit red", 2.80, 2.80, "kg", false, false},
		{"Scottish Potatoes 2.5kg", "potato vegetable scottish paper sack", 2.50, 1.00, "kg", true, false},
		{"Organic Carrots Loose", "carrot vegetable", 1.80, 1.80, "kg", false, true},
		{"Aberdeen Angus Beef Mince 500g", "beef mince meat angus", 6.50, 13.00, "kg", false, false},
		{"Scottish Smoked Salmon 200g", "salmon smoked fish scottish", 7.00, 35.00, "kg", false, false},
		{"Mature Scottish Cheddar 250g", "cheese cheddar mature dairy", 4.20, 16.80, "kg", false, false},
	},
}

// Run inserts the retailers and stores when the retailers table is empty, then
// adds the mock products to any listed retailer that has no products yet.
func Run(db *gorm.DB) error {
	if err := seedRetailers(db); err != nil {
		return err
	}
	return seedMockProducts(db)
}

func seedRetailers(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.Retailer{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		nStores := 0
		for _, r := range retailers {
			ret := models.Retailer{Name: r.name, IsLocalBusiness: r.local}
			for _, s := range r.stores {
				ret.Stores = append(ret.Stores, models.Store{
					BranchName: s.branch, Format: s.format,
					Address: s.address, Postcode: s.postcode,
					Latitude: s.lat, Longitude: s.lng,
				})
			}
			if err := tx.Create(&ret).Error; err != nil {
				return err
			}
			nStores += len(ret.Stores)
		}
		log.Printf("seeded %d retailers and %d stores", len(retailers), nStores)
		return nil
	})
}

func seedMockProducts(db *gorm.DB) error {
	names := make([]string, 0, len(mockProducts))
	for name := range mockProducts {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		var ret models.Retailer
		err := db.Where("name = ?", name).First(&ret).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		var count int64
		if err := db.Model(&models.Product{}).Where("retailer_id = ?", ret.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		var products []models.Product
		for _, m := range mockProducts[name] {
			unitPrice := m.unitPrice
			products = append(products, models.Product{
				RetailerID: ret.ID, Name: m.name, Keywords: m.keywords,
				IsRecyclable: m.recyclable, IsOrganic: m.organic, Source: mockSource,
				Price: models.ProductPrice{PriceGBP: m.price, UnitPriceGBP: &unitPrice, Unit: m.unit},
			})
		}
		// Creating the products also creates their Price rows (has-one association).
		if err := db.Create(&products).Error; err != nil {
			return fmt.Errorf("seed mock products for %q: %w", name, err)
		}
		log.Printf("seeded %d mock products for %s", len(products), name)
	}
	return nil
}
