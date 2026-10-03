package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	NATSURL      string
	HTTPAddr     string
	ServiceToken string
	MaxInflight  int
	MaxBodyBytes int64
	ReqTimeout   time.Duration
}

func Load() (Config, error) {
	token := os.Getenv("GATEWAY_SERVICE_TOKEN")
	if token == "" {
		return Config{}, errors.New("GATEWAY_SERVICE_TOKEN is required")
	}

	maxInflight := envInt("GATEWAY_MAX_INFLIGHT", 64)
	maxBody := int64(envInt("GATEWAY_MAX_BODY_KIB", 256)) << 10

	return Config{
		NATSURL:      env("NATS_URL", "nats://127.0.0.1:4222"),
		HTTPAddr:     env("HTTP_ADDR", ":8080"),
		ServiceToken: token,
		MaxInflight:  maxInflight,
		MaxBodyBytes: maxBody,
		ReqTimeout:   time.Duration(envInt("GATEWAY_TIMEOUT_MS", 2000)) * time.Millisecond,
	}, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
