package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv            string
	AppPort           int
	AppName           string
	Debug             bool
	DatabaseURL       string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	RedisURL          string
	AIServiceURL      string
	JWTSecret         string
	JWTIssuer         string
	JWTAccessExpiry   time.Duration
	AllowedOrigins         []string
	FirebaseProjectID      string
	FirebaseClientEmail    string
	FirebasePrivateKey     string
	FirebaseStorageBucket  string
	FirebaseCredentialsFile string
	RoutingAPIKey          string
}

func LoadConfig() (*Config, error) {
	appEnv := getEnv("APP_ENV", "development")
	appPort, err := strconv.Atoi(getEnv("APP_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid APP_PORT: %w", err)
	}

	debug := getEnv("DEBUG", "false") == "true"
	dbURL := getEnv("DATABASE_URL", "postgres://logiflows:logiflows_secret@localhost:5432/logiflows?sslmode=disable")
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379/0")
	aiURL := getEnv("AI_SERVICE_URL", "http://localhost:8000")
	jwtSecret := getEnv("JWT_SECRET", "default_dev_jwt_secret_must_be_overridden_in_prod")
	jwtIssuer := getEnv("JWT_ISSUER", "logiflows-auth-service")

	originsStr := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:5174,http://localhost:3000,http://127.0.0.1:5173,http://127.0.0.1:5174")
	var origins []string
	for _, o := range strings.Split(originsStr, ",") {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	return &Config{
		AppEnv:            appEnv,
		AppPort:           appPort,
		AppName:           getEnv("APP_NAME", "LogiFlows"),
		Debug:             debug,
		DatabaseURL:       dbURL,
		DBMaxOpenConns:    25,
		DBMaxIdleConns:    10,
		DBConnMaxLifetime: 5 * time.Minute,
		RedisURL:          redisURL,
		AIServiceURL:      aiURL,
		JWTSecret:         jwtSecret,
		JWTIssuer:         jwtIssuer,
		JWTAccessExpiry:        15 * time.Minute,
		AllowedOrigins:         origins,
		FirebaseProjectID:      getEnv("FIREBASE_PROJECT_ID", ""),
		FirebaseClientEmail:    getEnv("FIREBASE_CLIENT_EMAIL", ""),
		FirebasePrivateKey:     getEnv("FIREBASE_PRIVATE_KEY", ""),
		FirebaseStorageBucket:  getEnv("FIREBASE_STORAGE_BUCKET", ""),
		FirebaseCredentialsFile: getEnv("GOOGLE_APPLICATION_CREDENTIALS", getEnv("FIREBASE_CREDENTIALS_FILE", "serviceAccountKey.json")),
		RoutingAPIKey:          strings.Trim(strings.TrimSpace(getEnv("ORS_API_KEY", getEnv("ROUTING_API_KEY", getEnv("OPENROUTESERVICE_API_KEY", "")))), "\"'"),
	}, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
