package service

import (
	"math"
	"sort"

	"github.com/putita-lek/goodbuy-backend/internal/dto"
	"github.com/putita-lek/goodbuy-backend/internal/models"
)

const (
	// MaxOptionsPerItem is how many suggestions are returned for each shopping list item.
	MaxOptionsPerItem = 3
	// MaxStoresPerProduct is how many of a retailer's stores (the nearest ones in range)
	// a product can be suggested at.
	MaxStoresPerProduct = 3
	// MinMatchScore: a product must contain more than this share of the item's words to count as a match,
	// so one generic shared word ("fruit" in "dragon fruit") is not enough.
	MinMatchScore = 0.5

	relevanceWeight     = 1.0 // how closely the product name matches the item text
	priorityGroupWeight = 1.0 // best price / closest location
	// Sustainability is secondary to the priorities: its maximum (0.4) stays below
	// the top priority's full weight, so it breaks ties rather than overriding them.
	sustainabilityWeight = 0.2 // local business / recyclable / organic
)

// Candidate is a product that matched a shopping list item, at one store of its
// retailer, with its computed score. Product has Price loaded; Store has Retailer.
type Candidate struct {
	Product       models.Product
	Store         models.Store
	Match         float64
	DistanceMiles float64
	Score         float64
}

// rankWeight turns a 1-based rank (1 = most important) into a weight in (0, 1].
// A rank of 0 or less, or above n, means "not important" and gives 0.
func rankWeight(rank, n int) float64 {
	if rank <= 0 || rank > n {
		return 0
	}
	return float64(n-rank+1) / float64(n)
}

// HaversineMiles returns the great-circle distance between two points in miles.
func HaversineMiles(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusMiles = 3958.8
	toRad := func(d float64) float64 { return d * math.Pi / 180 }
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * earthRadiusMiles * math.Asin(math.Min(1, math.Sqrt(a)))
}

// normalizeLowerIsBetter maps v into 0..1 where the minimum gets 1 and the maximum 0.
func normalizeLowerIsBetter(v, min, max float64) float64 {
	if max-min < 1e-9 {
		return 1
	}
	return (max - v) / (max - min)
}

func boolScore(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// RankCandidates matches products to one shopping list item, pairs each with up to
// MaxStoresPerProduct of its retailer's nearest stores in range, drops those over
// budget, scores them against the user's preference and returns the best ones.
func RankCandidates(item dto.ShoppingItem, products []models.Product, stores []models.Store, pref dto.SubmitPreferenceRequest) []Candidate {
	queryTokens := Tokenize(item.ItemName)
	qty := item.ItemQuantity
	if qty <= 0 {
		qty = 1
	}

	// Stores in range per retailer, nearest first.
	type near struct {
		store models.Store
		dist  float64
	}
	byRetailer := map[uint][]near{}
	for _, st := range stores {
		dist := HaversineMiles(pref.UserLocation.Latitude, pref.UserLocation.Longitude, st.Latitude, st.Longitude)
		if pref.MaximumDistanceMiles > 0 && dist > pref.MaximumDistanceMiles {
			continue
		}
		byRetailer[st.RetailerID] = append(byRetailer[st.RetailerID], near{st, dist})
	}
	for id := range byRetailer {
		ns := byRetailer[id]
		sort.SliceStable(ns, func(i, j int) bool { return ns[i].dist < ns[j].dist })
		if len(ns) > MaxStoresPerProduct {
			byRetailer[id] = ns[:MaxStoresPerProduct]
		}
	}

	var cands []Candidate
	for _, p := range products {
		match := MatchScore(queryTokens, p.Name, p.Keywords)
		if match <= MinMatchScore {
			continue
		}
		if pref.BudgetGBP > 0 && p.Price.PriceGBP*qty > pref.BudgetGBP {
			continue
		}
		for _, n := range byRetailer[p.RetailerID] {
			cands = append(cands, Candidate{Product: p, Store: n.store, Match: match, DistanceMiles: n.dist})
		}
	}
	if len(cands) == 0 {
		return nil
	}

	minPrice, maxPrice := math.Inf(1), math.Inf(-1)
	minDist, maxDist := math.Inf(1), math.Inf(-1)
	for _, c := range cands {
		minPrice, maxPrice = math.Min(minPrice, c.Product.Price.PriceGBP), math.Max(maxPrice, c.Product.Price.PriceGBP)
		minDist, maxDist = math.Min(minDist, c.DistanceMiles), math.Max(maxDist, c.DistanceMiles)
	}

	pp, ps := pref.PreferencePriority, pref.PreferenceSustainability
	wPrice, wDist := rankWeight(pp.BestPrice, 2), rankWeight(pp.ClosestLocation, 2)
	wLocal, wRecycle, wOrganic := rankWeight(ps.LocalBusiness, 3), rankWeight(ps.Recyclable, 3), rankWeight(ps.Organic, 3)

	for i := range cands {
		c := &cands[i]
		priority := wPrice*normalizeLowerIsBetter(c.Product.Price.PriceGBP, minPrice, maxPrice) +
			wDist*normalizeLowerIsBetter(c.DistanceMiles, minDist, maxDist)
		sustain := wLocal*boolScore(c.Store.Retailer.IsLocalBusiness) +
			wRecycle*boolScore(c.Product.IsRecyclable) +
			wOrganic*boolScore(c.Product.IsOrganic)
		c.Score = relevanceWeight*c.Match + priorityGroupWeight*priority + sustainabilityWeight*sustain
	}

	sort.SliceStable(cands, func(i, j int) bool {
		if cands[i].Score != cands[j].Score {
			return cands[i].Score > cands[j].Score
		}
		return cands[i].Product.Price.PriceGBP < cands[j].Product.Price.PriceGBP
	})
	if len(cands) > MaxOptionsPerItem {
		cands = cands[:MaxOptionsPerItem]
	}
	return cands
}
