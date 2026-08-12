package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func ApplicationPort() (int, error) {
	raw := valueOrDefault("APP_PORT", "8080")
	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("APP_PORT harus berupa port yang valid")
	}
	return port, nil
}

func AuthTokenTTL() (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv("AUTH_TOKEN_TTL"))
	if raw == "" {
		raw = "8h"
	}
	ttl, err := time.ParseDuration(raw)
	if err != nil || ttl <= 0 {
		return 0, fmt.Errorf("AUTH_TOKEN_TTL harus berupa durasi positif, contoh 8h")
	}
	return ttl, nil
}
