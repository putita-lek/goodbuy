package repository

import (
	"math"

	"github.com/putita-lek/goodbuy-backend/internal/models"
	"gorm.io/gorm"
)

type ProductRepository struct{ db *gorm.DB }

func NewProductRepository(db *gorm.DB) *ProductRepository { return &ProductRepository{db: db} }

// Area is a circle around a point. A RadiusMiles of 0 or less means "anywhere".
type Area struct {
	Latitude, Longitude float64
	RadiusMiles         float64
}

const milesPerDegreeLat = 69.0

// FindStores returns stores that may lie inside area, with Retailer loaded.
//
// It filters with a latitude/longitude bounding box (cheap, uses idx_store_geo).
// The box is slightly larger than the circle, so the caller still applies the
// exact distance check.
func (r *ProductRepository) FindStores(area Area) ([]models.Store, error) {
	q := r.db.Preload("Retailer")
	if area.RadiusMiles > 0 {
		dLat := area.RadiusMiles / milesPerDegreeLat
		// Longitude degrees shrink towards the poles; clamp to avoid dividing by ~0.
		cosLat := math.Max(math.Cos(area.Latitude*math.Pi/180), 0.01)
		dLng := area.RadiusMiles / (milesPerDegreeLat * cosLat)
		q = q.Where("latitude BETWEEN ? AND ?", area.Latitude-dLat, area.Latitude+dLat).
			Where("longitude BETWEEN ? AND ?", area.Longitude-dLng, area.Longitude+dLng)
	}
	var stores []models.Store
	err := q.Find(&stores).Error
	return stores, err
}

// FindProducts returns priced products of the given retailers whose name or
// keywords contain at least one token, with Price loaded.
func (r *ProductRepository) FindProducts(tokens []string, retailerIDs []uint) ([]models.Product, error) {
	if len(tokens) == 0 || len(retailerIDs) == 0 {
		return nil, nil
	}
	match := r.db.Where("1 = 0")
	for _, t := range tokens {
		like := "%" + t + "%"
		match = match.Or("LOWER(products.name) LIKE ? OR LOWER(products.keywords) LIKE ?", like, like)
	}
	var products []models.Product
	err := r.db.InnerJoins("Price"). // products without a price are skipped
						Where("products.retailer_id IN ?", retailerIDs).
						Where(match).
						Find(&products).Error
	return products, err
}
