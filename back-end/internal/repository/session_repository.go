package repository

import (
	"errors"

	"github.com/putita-lek/goodbuy-backend/internal/models"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type SessionRepository struct{ db *gorm.DB }

func NewSessionRepository(db *gorm.DB) *SessionRepository { return &SessionRepository{db: db} }

func (r *SessionRepository) Exists(id uint64) (bool, error) {
	var count int64
	err := r.db.Model(&models.ShoppingSession{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// Create stores the session with its items and options (GORM wraps it in a transaction).
func (r *SessionRepository) Create(s *models.ShoppingSession) error {
	return r.db.Create(s).Error
}

func (r *SessionRepository) Get(id uint64) (*models.ShoppingSession, error) {
	var s models.ShoppingSession
	err := r.db.
		Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("result_id") }).
		Preload("Items.Options", func(db *gorm.DB) *gorm.DB { return db.Order("option_id") }).
		First(&s, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &s, err
}

// SetSelections saves the chosen option for each session item, keyed by item primary key.
func (r *SessionRepository) SetSelections(selections map[uint]int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for itemID, optionID := range selections {
			if err := tx.Model(&models.SessionItem{}).Where("id = ?", itemID).
				Update("selected_option_id", optionID).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
