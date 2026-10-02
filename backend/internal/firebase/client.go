package firebase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	cloudstorage "cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"firebase.google.com/go/v4/storage"
	"google.golang.org/api/option"
	"logiflows/backend/internal/config"
)

type Service interface {
	IsAvailable() bool
	SendToDevice(ctx context.Context, token, title, body string, data map[string]string) (string, error)
	SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) (string, error)
	SendMulticast(ctx context.Context, tokens []string, title, body string, data map[string]string) (*messaging.BatchResponse, error)
	GetSignedURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error)
}

type Client struct {
	app       *firebase.App
	messaging *messaging.Client
	storage   *storage.Client
	bucket    string
	logger    *slog.Logger
	enabled   bool
}

func NewFirebaseService(ctx context.Context, cfg *config.Config, logger *slog.Logger) (Service, error) {
	if logger == nil {
		logger = slog.Default()
	}

	opt, hasCreds := resolveClientOption(cfg, logger)
	if !hasCreds {
		logger.Warn("Firebase credentials not configured or file not found; running in simulated mock mode")
		return &Client{
			logger:  logger,
			enabled: false,
		}, nil
	}

	fbConfig := &firebase.Config{
		ProjectID:     cfg.FirebaseProjectID,
		StorageBucket: cfg.FirebaseStorageBucket,
	}

	app, err := firebase.NewApp(ctx, fbConfig, opt)
	if err != nil {
		logger.Warn("Firebase initialization failed; falling back to simulated mode", slog.String("error", err.Error()))
		return &Client{
			logger:  logger,
			enabled: false,
		}, nil
	}

	var msgClient *messaging.Client
	msgClient, err = app.Messaging(ctx)
	if err != nil {
		logger.Warn("Failed to initialize Firebase Messaging client", slog.String("error", err.Error()))
	}

	var storageClient *storage.Client
	storageClient, err = app.Storage(ctx)
	if err != nil {
		logger.Warn("Failed to initialize Firebase Storage client", slog.String("error", err.Error()))
	}

	logger.Info("Firebase Admin SDK initialized successfully",
		slog.String("project_id", cfg.FirebaseProjectID),
		slog.String("bucket", cfg.FirebaseStorageBucket),
	)

	return &Client{
		app:       app,
		messaging: msgClient,
		storage:   storageClient,
		bucket:    cfg.FirebaseStorageBucket,
		logger:    logger,
		enabled:   true,
	}, nil
}

func (c *Client) IsAvailable() bool {
	return c != nil && c.enabled
}

func (c *Client) SendToDevice(ctx context.Context, token, title, body string, data map[string]string) (string, error) {
	if !c.IsAvailable() || c.messaging == nil {
		c.logger.Info("Simulated Firebase FCM SendToDevice",
			slog.String("device_token_preview", previewToken(token)),
			slog.String("title", title),
			slog.String("body", body),
		)
		return fmt.Sprintf("mock-fcm-msg-%d", time.Now().UnixNano()), nil
	}

	msg := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	msgID, err := c.messaging.Send(ctx, msg)
	if err != nil {
		c.logger.Error("Failed to send FCM push notification", slog.String("error", err.Error()))
		return "", fmt.Errorf("fcm send error: %w", err)
	}

	c.logger.Info("FCM push notification dispatched", slog.String("message_id", msgID))
	return msgID, nil
}

func (c *Client) SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) (string, error) {
	if !c.IsAvailable() || c.messaging == nil {
		c.logger.Info("Simulated Firebase FCM SendToTopic",
			slog.String("topic", topic),
			slog.String("title", title),
			slog.String("body", body),
		)
		return fmt.Sprintf("mock-topic-msg-%d", time.Now().UnixNano()), nil
	}

	msg := &messaging.Message{
		Topic: topic,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	msgID, err := c.messaging.Send(ctx, msg)
	if err != nil {
		c.logger.Error("Failed to send FCM topic message", slog.String("error", err.Error()))
		return "", fmt.Errorf("fcm topic error: %w", err)
	}

	c.logger.Info("FCM topic notification dispatched", slog.String("message_id", msgID), slog.String("topic", topic))
	return msgID, nil
}

func (c *Client) SendMulticast(ctx context.Context, tokens []string, title, body string, data map[string]string) (*messaging.BatchResponse, error) {
	if !c.IsAvailable() || c.messaging == nil {
		c.logger.Info("Simulated Firebase FCM SendMulticast",
			slog.Int("recipient_count", len(tokens)),
			slog.String("title", title),
		)
		return &messaging.BatchResponse{
			SuccessCount: len(tokens),
			FailureCount: 0,
		}, nil
	}

	multicastMsg := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	resp, err := c.messaging.SendEachForMulticast(ctx, multicastMsg)
	if err != nil {
		c.logger.Error("Failed to send multicast FCM notification", slog.String("error", err.Error()))
		return nil, fmt.Errorf("fcm multicast error: %w", err)
	}

	c.logger.Info("FCM multicast dispatched",
		slog.Int("success", resp.SuccessCount),
		slog.Int("failure", resp.FailureCount),
	)
	return resp, nil
}

func (c *Client) GetSignedURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error) {
	if !c.IsAvailable() || c.storage == nil {
		mockURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s?simulated_token=valid", c.bucket, objectPath)
		return mockURL, nil
	}

	bucketHandle, err := c.storage.DefaultBucket()
	if err != nil {
		return "", fmt.Errorf("failed to get default bucket: %w", err)
	}

	opts := &cloudstorage.SignedURLOptions{
		Scheme:  cloudstorage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(expiry),
	}

	url, err := bucketHandle.SignedURL(objectPath, opts)
	if err != nil {
		c.logger.Error("Failed to generate signed URL", slog.String("error", err.Error()))
		return "", fmt.Errorf("storage signed url error: %w", err)
	}

	return url, nil
}

func resolveClientOption(cfg *config.Config, logger *slog.Logger) (option.ClientOption, bool) {
	// 1. Explicit service account JSON file
	if cfg.FirebaseCredentialsFile != "" {
		if _, err := os.Stat(cfg.FirebaseCredentialsFile); err == nil {
			logger.Info("Using Firebase credentials file", slog.String("path", cfg.FirebaseCredentialsFile))
			return option.WithCredentialsFile(cfg.FirebaseCredentialsFile), true
		}
	}

	// 2. Direct JSON payload in environment
	if jsonStr := os.Getenv("FIREBASE_CREDENTIALS_JSON"); strings.TrimSpace(jsonStr) != "" {
		logger.Info("Using FIREBASE_CREDENTIALS_JSON from environment")
		return option.WithCredentialsJSON([]byte(jsonStr)), true
	}

	// 3. Reconstruct JSON from individual .env parameters
	if cfg.FirebaseClientEmail != "" && cfg.FirebasePrivateKey != "" && cfg.FirebaseProjectID != "" {
		privKey := strings.ReplaceAll(cfg.FirebasePrivateKey, "\\n", "\n")
		saMap := map[string]string{
			"type":                        "service_account",
			"project_id":                  cfg.FirebaseProjectID,
			"client_email":                cfg.FirebaseClientEmail,
			"private_key":                 privKey,
			"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
			"token_uri":                   "https://oauth2.googleapis.com/token",
			"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		}
		data, err := json.Marshal(saMap)
		if err == nil {
			logger.Info("Reconstructed Firebase credentials from .env parameters",
				slog.String("client_email", cfg.FirebaseClientEmail),
				slog.String("project_id", cfg.FirebaseProjectID),
			)
			return option.WithCredentialsJSON(data), true
		}
	}

	// 4. Default serviceAccountKey.json in working dir
	if _, err := os.Stat("serviceAccountKey.json"); err == nil {
		logger.Info("Found serviceAccountKey.json in working directory")
		return option.WithCredentialsFile("serviceAccountKey.json"), true
	}

	return nil, false
}

func previewToken(t string) string {
	if len(t) <= 10 {
		return t
	}
	return t[:6] + "..." + t[len(t)-4:]
}
