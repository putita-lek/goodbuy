package service

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/putita-lek/goodbuy-backend/internal/dto"
	"github.com/putita-lek/goodbuy-backend/internal/models"
	"github.com/putita-lek/goodbuy-backend/internal/repository"
)

var (
	ErrValidation      = errors.New("validation error")
	ErrSessionNotFound = errors.New("session not found")
)

type ProductFinder interface {
	FindByAnyToken(tokens []string) ([]models.Product, error)
}

type SessionStore interface {
	Exists(id uint64) (bool, error)
	Create(s *models.ShoppingSession) error
	Get(id uint64) (*models.ShoppingSession, error)
	SetSelections(selections map[uint]int) error
}

type ShoppingService struct {
	products ProductFinder
	sessions SessionStore
}

func NewShoppingService(p ProductFinder, s SessionStore) *ShoppingService {
	return &ShoppingService{products: p, sessions: s}
}

func validationErr(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrValidation, fmt.Sprintf(format, args...))
}

func validatePreference(req dto.SubmitPreferenceRequest) error {
	if len(req.ShoppingItemList) == 0 {
		return validationErr("shopping_item_list must not be empty")
	}
	if len(req.ShoppingItemList) > 100 {
		return validationErr("shopping_item_list may contain at most 100 items")
	}
	for i, it := range req.ShoppingItemList {
		if strings.TrimSpace(it.ItemName) == "" {
			return validationErr("shopping_item_list[%d].item_name is required", i)
		}
		if it.ItemQuantity < 0 {
			return validationErr("shopping_item_list[%d].item_quantity must not be negative", i)
		}
	}
	if req.MaximumDistanceMiles < 0 {
		return validationErr("maximum_distance_from_user_location_miles must not be negative")
	}
	if req.BudgetGBP < 0 {
		return validationErr("budget_gbp must not be negative")
	}
	return nil
}

// SubmitPreference computes suggestions for every shopping list item and stores them in a new session.
func (s *ShoppingService) SubmitPreference(req dto.SubmitPreferenceRequest) (uint64, error) {
	if err := validatePreference(req); err != nil {
		return 0, err
	}
	prefJSON, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}
	id, err := s.newSessionID()
	if err != nil {
		return 0, err
	}

	session := &models.ShoppingSession{ID: id, PreferenceJSON: string(prefJSON)}
	for i, item := range req.ShoppingItemList {
		products, err := s.products.FindByAnyToken(Tokenize(item.ItemName))
		if err != nil {
			return 0, err
		}
		qty := item.ItemQuantity
		if qty == 0 {
			qty = 1
		}
		si := models.SessionItem{ResultID: i + 1, ItemName: strings.TrimSpace(item.ItemName), Quantity: qty}
		for j, c := range RankCandidates(item, products, req) {
			si.Options = append(si.Options, models.SessionItemOption{
				OptionID:        j + 1,
				ProductID:       c.Product.ID,
				OptionName:      c.Product.Name,
				PriceGBP:        c.Product.PriceGBP,
				StoreName:       c.Product.Store.Name,
				Latitude:        c.Product.Store.Latitude,
				Longitude:       c.Product.Store.Longitude,
				DistanceMiles:   round2(c.DistanceMiles),
				IsRecyclable:    c.Product.IsRecyclable,
				IsLocalBusiness: c.Product.Store.IsLocalBusiness,
				IsOrganic:       c.Product.IsOrganic,
				Score:           c.Score,
			})
		}
		session.Items = append(session.Items, si)
	}
	if err := s.sessions.Create(session); err != nil {
		return 0, err
	}
	return id, nil
}

// newSessionID returns a random, unused 6 to 9 digit id that is hard to guess.
func (s *ShoppingService) newSessionID() (uint64, error) {
	for i := 0; i < 10; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(999_900_000))
		if err != nil {
			return 0, err
		}
		id := n.Uint64() + 100_000
		exists, err := s.sessions.Exists(id)
		if err != nil {
			return 0, err
		}
		if !exists {
			return id, nil
		}
	}
	return 0, errors.New("could not allocate a session id")
}

func (s *ShoppingService) getSession(id uint64) (*models.ShoppingSession, error) {
	if id == 0 {
		return nil, validationErr("session_id is required")
	}
	session, err := s.sessions.Get(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrSessionNotFound
	}
	return session, err
}

func toOption(o models.SessionItemOption) dto.Option {
	return dto.Option{
		OptionID:   o.OptionID,
		OptionName: o.OptionName,
		PriceGBP:   o.PriceGBP,
		Location: dto.OptionLocation{
			Name:      o.StoreName,
			Latitude:  o.Latitude,
			Longitude: o.Longitude,
		},
		DistanceMiles:   o.DistanceMiles,
		IsRecyclable:    o.IsRecyclable,
		IsLocalBusiness: o.IsLocalBusiness,
		IsOrganic:       o.IsOrganic,
	}
}

func (s *ShoppingService) GetOptionResult(sessionID uint64) (*dto.OptionResultResponse, error) {
	session, err := s.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	resp := &dto.OptionResultResponse{OptionResult: []dto.ItemOptionResult{}}
	for _, it := range session.Items {
		r := dto.ItemOptionResult{ResultID: it.ResultID, ItemName: it.ItemName, ItemQuantity: it.Quantity, Options: []dto.Option{}}
		for _, o := range it.Options {
			r.Options = append(r.Options, toOption(o))
		}
		resp.OptionResult = append(resp.OptionResult, r)
	}
	return resp, nil
}

func (s *ShoppingService) SubmitOption(req dto.SubmitOptionRequest) error {
	if len(req.SubmittedOption) == 0 {
		return validationErr("submitted_option must not be empty")
	}
	session, err := s.getSession(req.SessionID)
	if err != nil {
		return err
	}
	byResult := map[int]models.SessionItem{}
	for _, it := range session.Items {
		byResult[it.ResultID] = it
	}
	selections := map[uint]int{}
	for _, so := range req.SubmittedOption {
		item, ok := byResult[so.ResultID]
		if !ok {
			return validationErr("result_id %d does not exist in this session", so.ResultID)
		}
		found := false
		for _, o := range item.Options {
			if o.OptionID == so.SelectedOptionID {
				found = true
				break
			}
		}
		if !found {
			return validationErr("option_id %d does not exist for result_id %d", so.SelectedOptionID, so.ResultID)
		}
		selections[item.ID] = so.SelectedOptionID
	}
	return s.sessions.SetSelections(selections)
}

func (s *ShoppingService) GetSummary(sessionID uint64) (*dto.SummaryResponse, error) {
	session, err := s.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	var pref dto.SubmitPreferenceRequest
	_ = json.Unmarshal([]byte(session.PreferenceJSON), &pref)

	resp := &dto.SummaryResponse{Summary: []dto.SummaryItem{}, BudgetGBP: pref.BudgetGBP}
	for _, it := range session.Items {
		si := dto.SummaryItem{ResultID: it.ResultID, ItemName: it.ItemName, ItemQuantity: it.Quantity}
		if it.SelectedOptionID != nil {
			for _, o := range it.Options {
				if o.OptionID == *it.SelectedOptionID {
					opt := toOption(o)
					si.SelectedOption = &opt
					resp.TotalPriceGBP += o.PriceGBP * it.Quantity
					break
				}
			}
		}
		resp.Summary = append(resp.Summary, si)
	}
	resp.TotalPriceGBP = round2(resp.TotalPriceGBP)
	resp.WithinBudget = pref.BudgetGBP <= 0 || resp.TotalPriceGBP <= pref.BudgetGBP
	return resp, nil
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }
