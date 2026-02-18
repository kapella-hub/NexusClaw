package sentry

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	mw "github.com/kapella-hub/NexusClaw/internal/platform/middleware"
	"github.com/kapella-hub/NexusClaw/internal/platform/respond"
)

// Handler exposes HTTP endpoints for the Firewall module.
type Handler struct {
	Service Service
	AuthMW  func(http.Handler) http.Handler
}

// Routes returns a chi.Router with all Firewall routes mounted.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	if h.AuthMW != nil {
		r.Use(h.AuthMW)
	}

	r.Get("/audit", h.ListAudit)
	r.Get("/rules", h.ListRules)
	r.Post("/rules", h.CreateRule)
	r.Put("/rules/{id}", h.UpdateRule)
	r.Delete("/rules/{id}", h.DeleteRule)
	r.Get("/budget", h.GetBudget)
	r.Put("/budget", h.UpdateBudget)

	return r
}

func (h *Handler) ListAudit(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(mw.GetUserID(r.Context()))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	entries, err := h.Service.ListAuditEntries(r.Context(), userID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to list audit entries")
		return
	}

	respond.JSON(w, http.StatusOK, entries)
}

func (h *Handler) ListRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.Service.ListRules(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to list rules")
		return
	}

	respond.JSON(w, http.StatusOK, rules)
}

func (h *Handler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var rule Rule
	if !respond.Decode(w, r, &rule) {
		return
	}

	if err := h.Service.CreateRule(r.Context(), &rule); err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to create rule")
		return
	}

	respond.JSON(w, http.StatusCreated, rule)
}

func (h *Handler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid rule id")
		return
	}

	var rule Rule
	if !respond.Decode(w, r, &rule) {
		return
	}
	rule.ID = id

	if err := h.Service.UpdateRule(r.Context(), &rule); err != nil {
		if errors.Is(err, ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "not found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "failed to update rule")
		return
	}

	respond.JSON(w, http.StatusOK, rule)
}

func (h *Handler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid rule id")
		return
	}

	if err := h.Service.DeleteRule(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "not found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "failed to delete rule")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetBudget(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(mw.GetUserID(r.Context()))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	budget, err := h.Service.GetBudget(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respond.JSON(w, http.StatusOK, map[string]int64{"max_tokens": 0, "used_tokens": 0})
			return
		}
		respond.Error(w, http.StatusInternalServerError, "failed to get budget")
		return
	}

	respond.JSON(w, http.StatusOK, budget)
}

func (h *Handler) UpdateBudget(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(mw.GetUserID(r.Context()))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var budget BudgetCap
	if !respond.Decode(w, r, &budget) {
		return
	}
	budget.UserID = userID

	if budget.ID == uuid.Nil {
		budget.ID = uuid.New()
	}

	if err := h.Service.UpdateBudget(r.Context(), &budget); err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to update budget")
		return
	}

	respond.JSON(w, http.StatusOK, budget)
}
