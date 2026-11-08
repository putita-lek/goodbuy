package service

import (
	"fmt"
	"testing"

	"github.com/putita-lek/goodbuy-backend/internal/dto"
	"github.com/putita-lek/goodbuy-backend/internal/models"
)

func TestTokenize(t *testing.T) {
	got := Tokenize("A carton of Milk, apples")
	want := []string{"carton", "milk", "apple"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestMatchScore(t *testing.T) {
	q := Tokenize("diet coke small size")
	if s := MatchScore(q, "Diet Coke 330ml can", "diet coke small can"); s < 0.74 {
		t.Fatalf("expected strong match, got %v", s)
	}
	if s := MatchScore(q, "Jasmine rice 1kg", "rice"); s != 0 {
		t.Fatalf("expected no match, got %v", s)
	}
}

func TestHaversineMiles(t *testing.T) {
	// Edinburgh Waverley to Glasgow Central is about 41-42 miles.
	d := HaversineMiles(55.9521, -3.1890, 55.8590, -4.2583)
	if d < 40 || d > 43 {
		t.Fatalf("unexpected distance %v", d)
	}
}

// product builds a priced product of retailer rid.
func product(id, rid uint, name, keywords string, price float64) models.Product {
	return models.Product{ID: id, RetailerID: rid, Name: name, Keywords: keywords,
		Price: models.ProductPrice{ProductID: id, PriceGBP: price}}
}

func store(id, rid uint, branch string, lat, lng float64, local bool) models.Store {
	return models.Store{ID: id, RetailerID: rid, BranchName: branch, Latitude: lat, Longitude: lng,
		Retailer: models.Retailer{ID: rid, Name: fmt.Sprintf("R%d", rid), IsLocalBusiness: local}}
}

// sample: retailer 1 is near, retailer 2 (local) is further, retailer 3 is out of range.
func sampleStores() []models.Store {
	return []models.Store{
		store(1, 1, "Near", 55.950, -3.190, false),
		store(2, 2, "Far", 55.990, -3.190, true),
		store(3, 3, "TooFar", 56.500, -3.190, false),
	}
}

func sampleProducts() []models.Product {
	return []models.Product{
		product(1, 1, "Milk 1L carton", "milk carton", 1.50),
		product(2, 2, "Milk 1L carton", "milk carton", 0.90),
		product(3, 3, "Milk 1L carton", "milk carton", 0.10),
		product(4, 1, "Jasmine rice", "rice", 1.00),
	}
}

func basePref() dto.SubmitPreferenceRequest {
	return dto.SubmitPreferenceRequest{
		UserLocation:         dto.Location{Latitude: 55.950, Longitude: -3.190},
		MaximumDistanceMiles: 5,
		BudgetGBP:            25,
	}
}

func TestRankCandidates_PriceFirst(t *testing.T) {
	pref := basePref()
	pref.PreferencePriority = dto.PreferencePriority{BestPrice: 1, ClosestLocation: 2}
	got := RankCandidates(dto.ShoppingItem{ItemName: "a carton of milk", ItemQuantity: 2}, sampleProducts(), sampleStores(), pref)
	if len(got) != 2 {
		t.Fatalf("expected 2 candidates within distance, got %d", len(got))
	}
	if got[0].Product.ID != 2 {
		t.Fatalf("expected cheapest (id 2) first, got %d", got[0].Product.ID)
	}
}

func TestRankCandidates_ClosestFirst(t *testing.T) {
	pref := basePref()
	pref.PreferencePriority = dto.PreferencePriority{BestPrice: 2, ClosestLocation: 1}
	got := RankCandidates(dto.ShoppingItem{ItemName: "milk", ItemQuantity: 1}, sampleProducts(), sampleStores(), pref)
	if len(got) == 0 || got[0].Product.ID != 1 {
		t.Fatalf("expected closest (id 1) first, got %+v", got)
	}
}

func TestRankCandidates_WeakMatchExcluded(t *testing.T) {
	products := []models.Product{product(9, 1, "Gala apple", "red apple fruit", 0.5)}
	if got := RankCandidates(dto.ShoppingItem{ItemName: "dragon fruit"}, products, sampleStores(), basePref()); len(got) != 0 {
		t.Fatalf("dragon fruit should not match apples, got %+v", got)
	}
}

func TestRankCandidates_PriceBeatsSustainability(t *testing.T) {
	stores := []models.Store{store(1, 1, "", 55.95, -3.19, true), store(2, 2, "", 55.95, -3.19, false)}
	organic := product(1, 1, "Organic rice flour", "", 2.95)
	organic.IsOrganic, organic.IsRecyclable = true, true
	products := []models.Product{organic, product(2, 2, "Rice flour", "", 1.80)}
	pref := basePref()
	pref.PreferencePriority = dto.PreferencePriority{BestPrice: 1, ClosestLocation: 2}
	pref.PreferenceSustainability = dto.PreferenceSustainability{LocalBusiness: 1, Recyclable: 2, Organic: 3}
	got := RankCandidates(dto.ShoppingItem{ItemName: "rice flour"}, products, stores, pref)
	if len(got) != 2 || got[0].Product.ID != 2 {
		t.Fatalf("expected cheaper product first when best_price is rank 1, got %+v", got)
	}
}

func TestRankCandidates_Budget(t *testing.T) {
	pref := basePref()
	pref.BudgetGBP = 2
	got := RankCandidates(dto.ShoppingItem{ItemName: "milk", ItemQuantity: 2}, sampleProducts(), sampleStores(), pref)
	if len(got) != 1 || got[0].Product.ID != 2 {
		t.Fatalf("expected only id 2 (0.90 x 2 within £2), got %+v", got)
	}
}

func TestRankCandidates_NearestStoresPerProduct(t *testing.T) {
	// Five stores of one retailer: the product is offered at the 3 nearest in range only.
	stores := []models.Store{
		store(1, 1, "A", 55.950, -3.190, false),
		store(2, 1, "B", 55.955, -3.190, false),
		store(3, 1, "C", 55.960, -3.190, false),
		store(4, 1, "D", 55.965, -3.190, false),
		store(5, 1, "TooFar", 56.500, -3.190, false),
	}
	pref := basePref()
	pref.PreferencePriority = dto.PreferencePriority{ClosestLocation: 1}
	got := RankCandidates(dto.ShoppingItem{ItemName: "milk"}, []models.Product{product(7, 1, "Milk 1L carton", "milk", 1.10)}, stores, pref)
	if len(got) != MaxStoresPerProduct {
		t.Fatalf("expected %d candidates, got %+v", MaxStoresPerProduct, got)
	}
	for i, want := range []string{"A", "B", "C"} {
		if got[i].Store.BranchName != want {
			t.Fatalf("candidate %d: expected store %s, got %s", i, want, got[i].Store.BranchName)
		}
	}
	if name := got[0].Store.DisplayName(); name != "R1 A" {
		t.Fatalf("unexpected display name %q", name)
	}
}
