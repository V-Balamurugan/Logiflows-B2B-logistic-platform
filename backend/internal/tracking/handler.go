package tracking

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Configured via CORS middleware at outer router layer
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Handler struct {
	service Service
	logger  *slog.Logger
}

func NewHandler(svc Service, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		service: svc,
		logger:  logger,
	}
}

func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	var payload LocationUpdate
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":{"message":"Invalid JSON payload"}}`, http.StatusBadRequest)
		return
	}

	if payload.AssignmentID == "" {
		http.Error(w, `{"error":{"message":"assignment_id is required"}}`, http.StatusBadRequest)
		return
	}

	if err := h.service.RecordLocation(r.Context(), payload); err != nil {
		h.logger.Error("Failed to record location update", slog.String("error", err.Error()))
		http.Error(w, `{"error":{"message":"Failed to record location"}}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "SUCCESS",
		"message": "Location updated and broadcasted",
		"data":    payload,
	})
}

func (h *Handler) GetLocation(w http.ResponseWriter, r *http.Request) {
	assignmentID := chi.URLParam(r, "assignment_id")
	if assignmentID == "" {
		http.Error(w, `{"error":{"message":"Missing assignment_id parameter"}}`, http.StatusBadRequest)
		return
	}

	loc, err := h.service.GetLatestLocation(r.Context(), assignmentID)
	if err != nil {
		http.Error(w, `{"error":{"message":"Failed to retrieve location"}}`, http.StatusInternalServerError)
		return
	}

	if loc == nil {
		http.Error(w, `{"error":{"message":"No tracking telemetry available for this assignment"}}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "OK",
		"data":   loc,
	})
}

func (h *Handler) StreamWebSocket(w http.ResponseWriter, r *http.Request) {
	assignmentID := chi.URLParam(r, "assignment_id")
	if assignmentID == "" {
		http.Error(w, "assignment_id required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Warn("WebSocket upgrade failed", slog.String("error", err.Error()))
		return
	}
	defer conn.Close()

	// Initial push of latest cached coordinate if exists
	if initialLoc, _ := h.service.GetLatestLocation(r.Context(), assignmentID); initialLoc != nil {
		_ = conn.WriteJSON(initialLoc)
	}

	pubsub := h.service.Subscribe(r.Context(), assignmentID)
	if pubsub == nil {
		_ = conn.WriteJSON(map[string]string{"error": "pubsub broker unavailable"})
		return
	}
	defer pubsub.Close()

	ch := pubsub.Channel()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var loc LocationUpdate
			if err := json.Unmarshal([]byte(msg.Payload), &loc); err == nil {
				if err := conn.WriteJSON(loc); err != nil {
					return
				}
			}
		case <-ticker.C:
			// Ping heartbeat to maintain connection
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}
