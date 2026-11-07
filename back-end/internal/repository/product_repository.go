package repository

import (
	"github.com/putita-lek/goodbuy-backend/internal/models"
	"gorm.io/gorm"
)

type ProductRepository struct{ db *gorm.DB }

func NewProductRepository(db *gorm.DB) *ProductRepository { return &ProductRepository{db: db} }

// FindByAnyToken returns products whose name or keywords contain at least one token.
func (r *ProductRepository) FindByAnyToken(tokens []string) ([]models.Product, error) {
	if len(tokens) == 0 {
		return nil, nil
	}
	cond := r.db.Where("1 = 0")
	for _, t := range tokens {
		like := "%" + t + "%"
		cond = cond.Or("LOWER(name) LIKE ? OR LOWER(keywords) LIKE ?", like, like)
	}
	var products []models.Product
	err := r.db.Preload("Store").Where(cond).Find(&products).Error
	return products, err
}
