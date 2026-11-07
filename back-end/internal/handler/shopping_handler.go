package handler

import (
	"errors"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/putita-lek/goodbuy-backend/internal/dto"
	"github.com/putita-lek/goodbuy-backend/internal/service"
)

type ShoppingHandler struct{ svc *service.ShoppingService }

func NewShoppingHandler(svc *service.ShoppingService) *ShoppingHandler {
	return &ShoppingHandler{svc: svc}
}

func (h *ShoppingHandler) Register(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) })
	app.Post("/submit-preference", h.SubmitPreference)
	app.Get("/shopping-item/option-result", h.GetOptionResult)
	app.Post("/shopping-item/submit-option", h.SubmitOption)
	app.Get("/shopping-item/summary", h.GetSummary) // also serves /summary/ (strict routing is off)
}

// statusFor maps service errors to HTTP status codes and a client-safe message.
func statusFor(err error) (int, string) {
	switch {
	case errors.Is(err, service.ErrValidation):
		return fiber.StatusBadRequest, err.Error()
	case errors.Is(err, service.ErrSessionNotFound):
		return fiber.StatusNotFound, err.Error()
	default:
		log.Printf("internal error: %v", err)
		return fiber.StatusInternalServerError, "internal server error"
	}
}

func fail(c *fiber.Ctx, err error) error {
	status, msg := statusFor(err)
	return c.Status(status).JSON(dto.ErrorResponse{Error: msg})
}

func sessionIDFromQuery(c *fiber.Ctx) (uint64, error) {
	raw := c.Query("session_id")
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("session_id query parameter must be a positive integer")
	}
	return id, nil
}

// POST /submit-preference
func (h *ShoppingHandler) SubmitPreference(c *fiber.Ctx) error {
	var req dto.SubmitPreferenceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Error: "invalid JSON body: " + err.Error()})
	}
	id, err := h.svc.SubmitPreference(req)
	if err != nil {
		return fail(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.SubmitPreferenceResponse{SessionID: id})
}

// GET /shopping-item/option-result?session_id=
func (h *ShoppingHandler) GetOptionResult(c *fiber.Ctx) error {
	id, err := sessionIDFromQuery(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Error: err.Error()})
	}
	resp, err := h.svc.GetOptionResult(id)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(resp)
}

// POST /shopping-item/submit-option
func (h *ShoppingHandler) SubmitOption(c *fiber.Ctx) error {
	var req dto.SubmitOptionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.SubmitOptionResponse{Error: "invalid JSON body: " + err.Error()})
	}
	if err := h.svc.SubmitOption(req); err != nil {
		status, msg := statusFor(err)
		return c.Status(status).JSON(dto.SubmitOptionResponse{IsSuccess: false, Error: msg})
	}
	return c.JSON(dto.SubmitOptionResponse{IsSuccess: true})
}

// GET /shopping-item/summary?session_id=
func (h *ShoppingHandler) GetSummary(c *fiber.Ctx) error {
	id, err := sessionIDFromQuery(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Error: err.Error()})
	}
	resp, err := h.svc.GetSummary(id)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(resp)
}
