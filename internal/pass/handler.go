package pass

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kapella-hub/NexusClaw/internal/platform/middleware"
	"github.com/kapella-hub/NexusClaw/internal/platform/respond"
)

// Handler exposes HTTP endpoints for the AuthBridge module.
type Handler struct {
	Service Service
	AuthMW  func(http.Handler) http.Handler
}

// Routes returns a chi.Router with all AuthBridge routes mounted.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// Public routes (no auth required).
	r.Post("/register", h.Register)
	r.Post("/sessions", h.CreateSession)

	// Protected routes (auth required).
	r.Group(func(r chi.Router) {
		r.Use(h.AuthMW)
		r.Delete("/sessions/{id}", h.DeleteSession)
		r.Get("/vault", h.ListVault)
		r.Post("/vault", h.CreateVault)
		r.Delete("/vault/{id}", h.DeleteVault)
		r.Post("/relay/{provider}", h.RelayAuth)
	})

	return r
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !respond.Decode(w, r, &req) {
		return
	}

	user, err := h.Service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			respond.Error(w, http.StatusConflict, "email already registered")
			return
		}
		respond.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.JSON(w, http.StatusCreated, registerResponse{
		ID:    user.ID,
		Email: user.Email,
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !respond.Decode(w, r, &req) {
		return
	}

	session, err := h.Service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			respond.Error(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		respond.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.JSON(w, http.StatusOK, session)
}

func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid session id")
		return
	}

	if err := h.Service.Logout(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "session not found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListVault(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(middleware.GetUserID(r.Context()))
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "invalid user identity")
		return
	}

	entries, err := h.Service.ListCredentials(r.Context(), userID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.JSON(w, http.StatusOK, entries)
}

func (h *Handler) CreateVault(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(middleware.GetUserID(r.Context()))
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "invalid user identity")
		return
	}

	var cred Credential
	if !respond.Decode(w, r, &cred) {
		return
	}

	entry, err := h.Service.StoreCredential(r.Context(), userID, &cred)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.JSON(w, http.StatusCreated, entry)
}

func (h *Handler) DeleteVault(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid entry id")
		return
	}

	if err := h.Service.RemoveCredential(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "vault entry not found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RelayAuth(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(middleware.GetUserID(r.Context()))
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "invalid user identity")
		return
	}

	provider := chi.URLParam(r, "provider")
	if provider == "" {
		respond.Error(w, http.StatusBadRequest, "missing provider")
		return
	}

	cred, err := h.Service.Relay(r.Context(), userID, provider)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "no credentials for provider")
			return
		}
		respond.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{
		"provider":     cred.Provider,
		"access_token": cred.AccessToken,
	})
}
