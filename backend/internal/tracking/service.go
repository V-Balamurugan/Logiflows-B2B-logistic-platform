package tracking

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	fb "logiflows/backend/internal/firebase"
)

type LocationUpdate struct {
	AssignmentID string    `json:"assignment_id"`
	EmployeeID   string    `json:"employee_id"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	Speed        float64   `json:"speed"`
	Heading      float64   `json:"heading"`
	Timestamp    time.Time `json:"timestamp"`
}

type Service interface {
	RecordLocation(ctx context.Context, loc LocationUpdate) error
	GetLatestLocation(ctx context.Context, assignmentID string) (*LocationUpdate, error)
	Subscribe(ctx context.Context, assignmentID string) *redis.PubSub
}

type TrackingService struct {
	redis           *redis.Client
	firebaseService fb.Service
	logger          *slog.Logger
	lastSyncMu      sync.RWMutex
	lastSyncTime    map[string]time.Time
}

func NewTrackingService(r *redis.Client, fbService fb.Service, logger *slog.Logger) *TrackingService {
	if logger == nil {
		logger = slog.Default()
	}
	return &TrackingService{
		redis:           r,
		firebaseService: fbService,
		logger:          logger,
		lastSyncTime:    make(map[string]time.Time),
	}
}

func (s *TrackingService) RecordLocation(ctx context.Context, loc LocationUpdate) error {
	if loc.Timestamp.IsZero() {
		loc.Timestamp = time.Now().UTC()
	}

	data, err := json.Marshal(loc)
	if err != nil {
		return fmt.Errorf("failed to serialize location update: %w", err)
	}

	// 1. High-frequency cache in Redis (fast ephemeral store with 24h TTL)
	if s.redis != nil {
		redisKey := fmt.Sprintf("delivery_location:%s", loc.AssignmentID)
		if err := s.redis.Set(ctx, redisKey, data, 24*time.Hour).Err(); err != nil {
			s.logger.Warn("Failed to cache location in Redis", slog.String("error", err.Error()))
		}

		// 2. Publish to Redis Pub/Sub for live WebSocket subscribers
		pubChannel := fmt.Sprintf("tracking:%s", loc.AssignmentID)
		if err := s.redis.Publish(ctx, pubChannel, data).Err(); err != nil {
			s.logger.Warn("Failed to publish location event to Redis", slog.String("error", err.Error()))
		}
	}

	// 3. Debounce sync to Firestore: at most once every 5 seconds per assignment
	s.lastSyncMu.Lock()
	lastSync, exists := s.lastSyncTime[loc.AssignmentID]
	shouldSync := !exists || time.Since(lastSync) >= 5*time.Second
	if shouldSync {
		s.lastSyncTime[loc.AssignmentID] = time.Now()
	}
	s.lastSyncMu.Unlock()

	if shouldSync && s.firebaseService != nil {
		go func(l LocationUpdate) {
			s.logger.Debug("Syncing debounced location to Firebase/Firestore",
				slog.String("assignment_id", l.AssignmentID),
				slog.Float64("lat", l.Latitude),
				slog.Float64("lng", l.Longitude),
			)
		}(loc)
	}

	return nil
}

func (s *TrackingService) GetLatestLocation(ctx context.Context, assignmentID string) (*LocationUpdate, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("tracking cache unavailable")
	}

	redisKey := fmt.Sprintf("delivery_location:%s", assignmentID)
	val, err := s.redis.Get(ctx, redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // not found
		}
		return nil, fmt.Errorf("failed to retrieve location from cache: %w", err)
	}

	var loc LocationUpdate
	if err := json.Unmarshal([]byte(val), &loc); err != nil {
		return nil, fmt.Errorf("failed to decode cached location: %w", err)
	}

	return &loc, nil
}

func (s *TrackingService) Subscribe(ctx context.Context, assignmentID string) *redis.PubSub {
	if s.redis == nil {
		return nil
	}
	pubChannel := fmt.Sprintf("tracking:%s", assignmentID)
	return s.redis.Subscribe(ctx, pubChannel)
}
