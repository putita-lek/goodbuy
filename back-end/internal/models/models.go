package models

import "time"

// Store is a physical shop where products are sold.
type Store struct {
	ID              uint    `gorm:"primaryKey"`
	Name            string  `gorm:"size:255;not null"`
	Latitude        float64 `gorm:"not null"`
	Longitude       float64 `gorm:"not null"`
	IsLocalBusiness bool    `gorm:"not null;default:false"`
	Products        []Product
}

// Product is a single product item sold at one store at one price.
type Product struct {
	ID           uint    `gorm:"primaryKey"`
	Name         string  `gorm:"size:255;not null;index"`
	Keywords     string  `gorm:"size:512"` // extra search terms, space separated
	PriceGBP     float64 `gorm:"not null"`
	StoreID      uint    `gorm:"not null;index"`
	Store        Store
	IsRecyclable bool `gorm:"not null;default:false"`
	IsOrganic    bool `gorm:"not null;default:false"`
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
