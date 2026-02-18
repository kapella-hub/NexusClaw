package app

import (
	"crypto/sha256"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kapella-hub/NexusClaw/internal/nodes"
	"github.com/kapella-hub/NexusClaw/internal/pass"
	"github.com/kapella-hub/NexusClaw/internal/platform/config"
	mw "github.com/kapella-hub/NexusClaw/internal/platform/middleware"
	"github.com/kapella-hub/NexusClaw/internal/platform/respond"
	"github.com/kapella-hub/NexusClaw/internal/sentry"
	"golang.org/x/oauth2"
)

// New creates the HTTP handler with all routes and middleware wired up.
func New(cfg *config.Config, logger *slog.Logger, pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RealIP)
	r.Use(mw.RequestID)
	r.Use(mw.Logging)
	r.Use(mw.NewRateLimit(10, 10))
	r.Use(chimw.Recoverer)

	// Auth middleware constructor
	tokenSecret := []byte(cfg.Auth.TokenSecret)
	authMW := mw.Auth(tokenSecret)

	// Vault key: SHA-256 of the token secret to produce a 32-byte AES key.
	vaultKeyHash := sha256.Sum256(tokenSecret)
	vaultKey := vaultKeyHash[:]

	// Health endpoints
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			respond.Error(w, http.StatusServiceUnavailable, "database not ready")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})

	// -- Pass module --
	passRepo := pass.NewPgRepository(pool)
	userRepo := pass.NewPgUserRepository(pool)
	passRelay := pass.NewRelay(passRepo, vaultKey)
	passSvc := pass.NewService(passRepo, userRepo, passRelay, tokenSecret, cfg.Auth.TokenExpiry, vaultKey)
	passHandler := &pass.Handler{Service: passSvc, AuthMW: authMW}

	// -- Nodes module --
	nodesRepo := nodes.NewPgRepository(pool)
	var containerMgr nodes.ContainerManager
	if cm, err := nodes.NewContainerManager(); err != nil {
		logger.Warn("docker container manager unavailable, server lifecycle disabled", "error", err)
	} else {
		containerMgr = cm
	}
	nodesSvc := nodes.NewService(nodesRepo, containerMgr)
	nodesRegistry := nodes.NewRegistry(nodesRepo)
	nodesHandler := &nodes.Handler{Service: nodesSvc, Registry: nodesRegistry, AuthMW: authMW}

	// -- Sentry module --
	sentryRepo := sentry.NewPgRepository(pool)
	sentrySvc := sentry.NewService(sentryRepo)
	sentryHandler := &sentry.Handler{Service: sentrySvc, AuthMW: authMW}

	// -- OAuth handler (optional, from config) --
	var oauthHandler *nodes.OAuthHandler
	if len(cfg.OAuth) > 0 {
		providers := make(map[string]*oauth2.Config, len(cfg.OAuth))
		for name, p := range cfg.OAuth {
			providers[name] = &oauth2.Config{
				ClientID:     p.ClientID,
				ClientSecret: p.ClientSecret,
				Endpoint: oauth2.Endpoint{
					AuthURL:  p.AuthURL,
					TokenURL: p.TokenURL,
				},
				RedirectURL: p.RedirectURL,
				Scopes:      p.Scopes,
			}
		}
		oauthHandler = &nodes.OAuthHandler{
			Providers: providers,
			Repo:      nodesRepo,
			VaultKey:  vaultKey,
			AuthMW:    authMW,
		}
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/pass", passHandler.Routes())
		r.Mount("/nodes", nodesHandler.Routes())
		r.Mount("/sentry", sentryHandler.Routes())
		if oauthHandler != nil {
			r.Mount("/nodes/oauth", oauthHandler.Routes())
		}
	})

	return r
}
