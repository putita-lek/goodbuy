package service

import (
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

func sampleProducts() []models.Product {
	near := models.Store{Name: "Near", Latitude: 55.950, Longitude: -3.190}
	far := models.Store{Name: "Far", Latitude: 55.990, Longitude: -3.190, IsLocalBusiness: true}
	tooFar := models.Store{Name: "TooFar", Latitude: 56.500, Longitude: -3.190}
	return []models.Product{
		{ID: 1, Name: "Milk 1L carton", Keywords: "milk carton", PriceGBP: 1.50, Store: near},
		{ID: 2, Name: "Milk 1L carton", Keywords: "milk carton", PriceGBP: 0.90, Store: far},
		{ID: 3, Name: "Milk 1L carton", Keywords: "milk carton", PriceGBP: 0.10, Store: tooFar},
		{ID: 4, Name: "Jasmine rice", Keywords: "rice", PriceGBP: 1.00, Store: near},
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
	got := RankCandidates(dto.ShoppingItem{ItemName: "a carton of milk", ItemQuantity: 2}, sampleProducts(), pref)
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
	got := RankCandidates(dto.ShoppingItem{ItemName: "milk", ItemQuantity: 1}, sampleProducts(), pref)
	if len(got) == 0 || got[0].Product.ID != 1 {
		t.Fatalf("expected closest (id 1) first, got %+v", got)
	}
}

func TestRankCandidates_WeakMatchExcluded(t *testing.T) {
	products := []models.Product{{ID: 9, Name: "Gala apple", Keywords: "red apple fruit", PriceGBP: 0.5,
		Store: models.Store{Latitude: 55.95, Longitude: -3.19}}}
	if got := RankCandidates(dto.ShoppingItem{ItemName: "dragon fruit"}, products, basePref()); len(got) != 0 {
		t.Fatalf("dragon fruit should not match apples, got %+v", got)
	}
}

func TestRankCandidates_PriceBeatsSustainability(t *testing.T) {
	store := models.Store{Latitude: 55.95, Longitude: -3.19}
	local := models.Store{Latitude: 55.95, Longitude: -3.19, IsLocalBusiness: true}
	products := []models.Product{
		{ID: 1, Name: "Organic rice flour", PriceGBP: 2.95, IsOrganic: true, IsRecyclable: true, Store: local},
		{ID: 2, Name: "Rice flour", PriceGBP: 1.80, Store: store},
	}
	pref := basePref()
	pref.PreferencePriority = dto.PreferencePriority{BestPrice: 1, ClosestLocation: 2}
	pref.PreferenceSustainability = dto.PreferenceSustainability{LocalBusiness: 1, Recyclable: 2, Organic: 3}
	got := RankCandidates(dto.ShoppingItem{ItemName: "rice flour"}, products, pref)
	if len(got) != 2 || got[0].Product.ID != 2 {
		t.Fatalf("expected cheaper product first when best_price is rank 1, got %+v", got)
	}
}

func TestRankCandidates_Budget(t *testing.T) {
	pref := basePref()
	pref.BudgetGBP = 2
	got := RankCandidates(dto.ShoppingItem{ItemName: "milk", ItemQuantity: 2}, sampleProducts(), pref)
	if len(got) != 1 || got[0].Product.ID != 2 {
		t.Fatalf("expected only id 2 (0.90 x 2 within £2), got %+v", got)
	}
}
