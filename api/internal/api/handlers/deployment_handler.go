package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"kari/api/internal/core/services"
	"kari/api/internal/api/middleware"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096, // Larger buffer for log-heavy builds
	CheckOrigin: func(r *http.Request) bool {
		// 🛡️ Zero-Trust: In production, strictly match the SvelteKit URL
		// For now, we'll allow cross-origin but enforce auth via context
		return true 
	},
}

type DeploymentHandler struct {
	appService *services.ApplicationService
}

func NewDeploymentHandler(svc *services.ApplicationService) *DeploymentHandler {
	return &DeploymentHandler{appService: svc}
}

// StreamLogs handles the WebSocket upgrade and pipes the service channel to the client
func (h *DeploymentHandler) StreamLogs(w http.ResponseWriter, r *http.Request) {
	// 1. Extract IDs and Context
	appIDStr := chi.URLParam(r, "id")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		http.Error(w, "Invalid Application ID", http.StatusBadRequest)
		return
	}

	// 🛡️ SLA: Extract UserID from the RBAC Middleware context
	userID, ok := r.Context().Value(middleware.UserKey).(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return // Upgrade handles its own error response
	}
	defer conn.Close()

	// 3. Trigger Service Orchestration
	// We pass the request context so that if the WS disconnects, the gRPC call cancels.
	logChan, err := h.appService.Deploy(r.Context(), appID, userID)
	if err != nil {
		_ = conn.WriteJSON(map[string]string{"error": err.Error()})
		return
	}

	// 🛡️ 4. The Pipe Loop (Memory-Safe)
	// We use a "Ping-Pong" heartbeat to detect stale connections.
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case logLine, ok := <-logChan:
			if !ok {
				// Service closed the channel (Build Finished)
				_ = conn.WriteMessage(websocket.TextMessage, []byte("🏗️ BUILD_COMPLETE"))
				return
			}

			// Write the log line to xterm.js
			err := conn.WriteMessage(websocket.TextMessage, []byte(logLine))
			if err != nil {
				return // Client disconnected
			}

		case <-ticker.C:
			// Send heartbeat to keep the connection alive through Nginx proxy
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-r.Context().Done():
			// 🛡️ Zero-Waste: If the user leaves, the gRPC context is cancelled automatically.
			return
		}
	}
}
