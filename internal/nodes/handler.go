package nodes

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	mw "github.com/kapella-hub/NexusClaw/internal/platform/middleware"
	"github.com/kapella-hub/NexusClaw/internal/platform/respond"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handler exposes HTTP endpoints for the Managed MCP module.
type Handler struct {
	Service     Service
	Registry    Registry
	AuthMW      func(http.Handler) http.Handler
	RateLimiter *RateLimiter
}

// Routes returns a chi.Router with all MCP server routes mounted.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	if h.AuthMW != nil {
		r.Use(h.AuthMW)
	}

	r.Get("/", h.ListServers)
	r.Post("/", h.RegisterServer)
	r.Get("/discover", h.DiscoverServers)
	r.Get("/{id}", h.GetServer)
	r.Delete("/{id}", h.RemoveServer)
	r.Post("/{id}/start", h.StartServer)
	r.Post("/{id}/stop", h.StopServer)
	r.Get("/{id}/ws", h.ConnectWebSocket)

	return r
}

func (h *Handler) ListServers(w http.ResponseWriter, r *http.Request) {
	ownerID, err := uuid.Parse(mw.GetUserID(r.Context()))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	servers, err := h.Service.ListServers(r.Context(), ownerID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to list servers")
		return
	}

	respond.JSON(w, http.StatusOK, servers)
}

func (h *Handler) DiscoverServers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	servers, err := h.Registry.Discover(r.Context(), query)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to discover servers")
		return
	}

	respond.JSON(w, http.StatusOK, servers)
}

func (h *Handler) RegisterServer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string         `json:"name"`
		Image  string         `json:"image"`
		Config map[string]any `json:"config"`
	}
	if !respond.Decode(w, r, &req) {
		return
	}

	ownerID, err := uuid.Parse(mw.GetUserID(r.Context()))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	server := &MCPServer{
		OwnerID: ownerID,
		Name:    req.Name,
		Image:   req.Image,
		Config:  req.Config,
	}

	if err := h.Service.RegisterServer(r.Context(), server); err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to register server")
		return
	}

	respond.JSON(w, http.StatusCreated, server)
}

func (h *Handler) GetServer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid server id")
		return
	}

	server, err := h.Service.GetServer(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, server)
}

func (h *Handler) RemoveServer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid server id")
		return
	}

	if err := h.Service.RemoveServer(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) StartServer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid server id")
		return
	}

	if err := h.Service.StartServer(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func (h *Handler) StopServer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid server id")
		return
	}

	if err := h.Service.StopServer(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (h *Handler) ConnectWebSocket(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid server id")
		return
	}

	// Validate server is running.
	if err := h.Service.ConnectWebSocket(r.Context(), id, w, r); err != nil {
		handleServiceError(w, err)
		return
	}

	// Upgrade client connection to WebSocket.
	clientConn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer clientConn.Close()

	// Get server to find container port for backend connection.
	server, err := h.Service.GetServer(r.Context(), id)
	if err != nil {
		clientConn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "server not found"))
		return
	}

	// Determine backend WebSocket URL from server config.
	backendPort := "8080"
	if p, ok := server.Config["ws_port"].(string); ok {
		backendPort = p
	}
	backendURL := "ws://localhost:" + backendPort

	// Connect to backend MCP server.
	backendConn, _, err := websocket.DefaultDialer.Dial(backendURL, nil)
	if err != nil {
		clientConn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "backend connection failed"))
		return
	}
	defer backendConn.Close()

	// Bidirectional proxy.
	done := make(chan struct{})

	// Client → Backend
	go func() {
		defer close(done)
		for {
			mt, msg, err := clientConn.ReadMessage()
			if err != nil {
				return
			}

			if h.RateLimiter != nil && !h.RateLimiter.Allow(server.ID) {
				clientConn.WriteMessage(mt, []byte(`{"jsonrpc":"2.0","error":{"code":-32005,"message":"Rate limit exceeded"},"id":null}`))
				continue
			}

			if err := backendConn.WriteMessage(mt, msg); err != nil {
				return
			}
		}
	}()

	// Backend → Client
	for {
		mt, msg, err := backendConn.ReadMessage()
		if err != nil {
			break
		}
		if err := clientConn.WriteMessage(mt, msg); err != nil {
			break
		}
	}

	<-done
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		respond.Error(w, http.StatusNotFound, "not found")
	case errors.Is(err, ErrContainerNotAvailable):
		respond.Error(w, http.StatusServiceUnavailable, "container runtime not available")
	case errors.Is(err, ErrNotImplemented):
		respond.Error(w, http.StatusNotImplemented, "not implemented")
	default:
		respond.Error(w, http.StatusInternalServerError, "internal server error")
	}
}
