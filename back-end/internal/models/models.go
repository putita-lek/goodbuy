package models

import (
	"strings"
	"time"
)

// Retailer is a brand or chain (e.g. "Sainsbury's"), or a single independent business.
type Retailer struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:255;not null;uniqueIndex"`
	// IsLocalBusiness describes the business as a whole, so it lives here rather than on each branch.
	IsLocalBusiness bool `gorm:"not null;default:false"`
	Stores          []Store
	Products        []Product
}

// Store is one physical branch of a retailer.
type Store struct {
	ID         uint `gorm:"primaryKey"`
	RetailerID uint `gorm:"not null;index"`
	Retailer   Retailer
	BranchName string  `gorm:"size:255;not null"` // e.g. "Cameron Toll"
	Format     string  `gorm:"size:50"`           // e.g. "superstore", "local", "express"
	Address    string  `gorm:"size:512"`
	Postcode   string  `gorm:"size:16;index"`
	Latitude   float64 `gorm:"not null;index:idx_store_geo"`
	Longitude  float64 `gorm:"not null;index:idx_store_geo"`
}

// DisplayName is the retailer name followed by the branch name, e.g. "Sainsbury's Cameron Toll".
func (s Store) DisplayName() string {
	return strings.TrimSpace(s.Retailer.Name + " " + s.BranchName)
}

// Product is a catalogue item a retailer sells. It is sold at all of the retailer's
// stores at one chain-wide price, kept in ProductPrice.
type Product struct {
	ID           uint `gorm:"primaryKey"`
	RetailerID   uint `gorm:"not null;uniqueIndex:idx_retailer_product"`
	Retailer     Retailer
	Name         string       `gorm:"size:255;not null;index;uniqueIndex:idx_retailer_product"`
	Keywords     string       `gorm:"size:512"` // extra search terms, space separated
	IsRecyclable bool         `gorm:"not null;default:false"`
	IsOrganic    bool         `gorm:"not null;default:false"`
	Price        ProductPrice `gorm:"constraint:OnDelete:CASCADE"`

	// Optional fields filled by the Kaggle import (cmd/import-kaggle).
	IsOwnBrand bool   `gorm:"not null;default:false"`
	Source     string `gorm:"size:32;index"` // "" for hand-entered data, "kaggle" for imported rows
}

// ProductPrice is the current price of a product, the same at every store of its retailer.
type ProductPrice struct {
	ProductID    uint       `gorm:"primaryKey;autoIncrement:false"`
	PriceGBP     float64    `gorm:"not null"`
	UnitPriceGBP *float64   // price per Unit, nil when unknown
	Unit         string     `gorm:"size:16"` // e.g. kg, l, unit
	PriceDate    *time.Time // day the price was recorded, nil when unknown
}

// ShoppingSession holds one submitted preference and its suggested options.
// It is not a user profile: it only lives as long as one shopping flow.
type ShoppingSession struct {
	ID             uint64 `gorm:"primaryKey;autoIncrement:false"`
	PreferenceJSON string `gorm:"type:json;not null"`
	CreatedAt      time.Time
	Items          []SessionItem `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE"`
}

// SessionItem is one entry of the user's shopping list.
type SessionItem struct {
	ID               uint   `gorm:"primaryKey"`
	SessionID        uint64 `gorm:"not null;uniqueIndex:idx_session_result"`
	ResultID         int    `gorm:"not null;uniqueIndex:idx_session_result"`
	ItemName         string `gorm:"size:255;not null"`
	Quantity         float64
	SelectedOptionID *int
	Options          []SessionItemOption `gorm:"constraint:OnDelete:CASCADE"`
}

// SessionItemOption is a snapshot of a suggested product, so results stay
// stable even if prices in the catalogue change later.
type SessionItemOption struct {
	ID              uint `gorm:"primaryKey"`
	SessionItemID   uint `gorm:"not null;uniqueIndex:idx_item_option"`
	OptionID        int  `gorm:"not null;uniqueIndex:idx_item_option"`
	ProductID       uint
	StoreID         uint
	OptionName      string `gorm:"size:255;not null"`
	PriceGBP        float64
	StoreName       string `gorm:"size:255"`
	Latitude        float64
	Longitude       float64
	DistanceMiles   float64
	IsRecyclable    bool
	IsLocalBusiness bool
	IsOrganic       bool
	Score           float64
}
