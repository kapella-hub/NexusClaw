package nodes

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kapella-hub/NexusClaw/internal/platform/crypto"
	"github.com/kapella-hub/NexusClaw/internal/platform/respond"
	"golang.org/x/oauth2"
)

// OAuthHandler manages OAuth flows for MCP server integrations.
type OAuthHandler struct {
	Providers map[string]*oauth2.Config
	Repo      Repository
	VaultKey  []byte
	AuthMW    func(http.Handler) http.Handler
}

// Routes returns a chi.Router with OAuth-related routes.
func (h *OAuthHandler) Routes() chi.Router {
	r := chi.NewRouter()

	if h.AuthMW != nil {
		r.Use(h.AuthMW)
	}

	r.Get("/initiate/{provider}", h.InitiateFlow)
	r.Get("/callback/{provider}", h.HandleCallback)

	return r
}

func (h *OAuthHandler) InitiateFlow(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")

	oauthCfg, ok := h.Providers[providerName]
	if !ok {
		respond.Error(w, http.StatusNotFound, "unknown oauth provider")
		return
	}

	state, err := randomState()
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to generate state")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, oauthCfg.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")

	oauthCfg, ok := h.Providers[providerName]
	if !ok {
		respond.Error(w, http.StatusNotFound, "unknown oauth provider")
		return
	}

	cookie, err := r.Cookie("oauth_state")
	if err != nil {
		respond.Error(w, http.StatusForbidden, "missing state cookie")
		return
	}

	if r.URL.Query().Get("state") != cookie.Value {
		respond.Error(w, http.StatusForbidden, "state mismatch")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		respond.Error(w, http.StatusBadRequest, "missing authorization code")
		return
	}

	token, err := oauthCfg.Exchange(r.Context(), code)
	if err != nil {
		respond.Error(w, http.StatusBadGateway, "token exchange failed")
		return
	}

	encToken, err := crypto.Seal([]byte(token.AccessToken), h.VaultKey)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to encrypt token")
		return
	}

	grant := &OAuthGrant{
		ID:             uuid.New(),
		ServerID:       uuid.Nil,
		Provider:       providerName,
		AccessTokenEnc: encToken,
		CreatedAt:      time.Now(),
	}
	if token.Expiry.IsZero() {
		grant.ExpiresAt = nil
	} else {
		grant.ExpiresAt = &token.Expiry
	}

	if err := h.Repo.CreateOAuthGrant(r.Context(), grant); err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to store grant")
		return
	}

	// Clear the state cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	})

	respond.JSON(w, http.StatusOK, map[string]string{
		"status":   "connected",
		"provider": providerName,
	})
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
