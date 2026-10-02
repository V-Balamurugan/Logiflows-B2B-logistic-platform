package httpapi

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type DependencyChecker interface {
	CheckPostgres(ctx context.Context) error
	CheckRedis(ctx context.Context) error
	CheckAI(ctx context.Context) error
}

type HealthHandler struct {
	startTime time.Time
	version   string
	db        *sql.DB
	redis     *redis.Client
	aiURL     string
}

func NewHealthHandler(db *sql.DB, redisClient *redis.Client, aiURL string, version string) *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
		version:   version,
		db:        db,
		redis:     redisClient,
		aiURL:     aiURL,
	}
}

type LivenessResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	UptimeSec int64  `json:"uptime_seconds"`
	Timestamp string `json:"timestamp"`
}

type ReadinessResponse struct {
	Status       string            `json:"status"`
	Dependencies map[string]string `json:"dependencies"`
	Timestamp    string            `json:"timestamp"`
}

func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	resp := LivenessResponse{
		Status:    "OK",
		Version:   h.version,
		UptimeSec: int64(time.Since(h.startTime).Seconds()),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	RespondJSON(w, r, http.StatusOK, resp, nil)
}

func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	deps := make(map[string]string)
	isReady := true

	// Check Postgres if initialized
	if h.db != nil {
		if err := h.db.PingContext(ctx); err != nil {
			deps["postgres"] = "DOWN: " + err.Error()
			isReady = false
		} else {
			deps["postgres"] = "UP"
		}
	} else {
		deps["postgres"] = "UNINITIALIZED"
	}

	// Check Redis if initialized
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			deps["redis"] = "DOWN: " + err.Error()
			// Redis failure is non-fatal for critical database operations (graceful degradation)
		} else {
			deps["redis"] = "UP"
		}
	} else {
		deps["redis"] = "UNINITIALIZED"
	}

	deps["ai_service"] = "CONFIGURED: " + h.aiURL

	status := "READY"
	httpStatus := http.StatusOK
	if !isReady {
		status = "DEGRADED"
		httpStatus = http.StatusServiceUnavailable
	}

	resp := ReadinessResponse{
		Status:       status,
		Dependencies: deps,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}

	RespondJSON(w, r, httpStatus, resp, nil)
}
