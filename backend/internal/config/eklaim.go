package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type EKlaim struct {
	BaseURL   string
	WSURL     string
	Key       string
	KodeTarif string
	CoderNIK  string
	Timeout   time.Duration
	Retry     int
}

func EKlaimConfig() (EKlaim, error) {
	timeout, err := eklaimTimeout()
	if err != nil {
		return EKlaim{}, err
	}

	retry, err := strconv.Atoi(valueOrDefault("EKLAIM_RETRY", valueOrDefault("EKLAM_RETRY", "3")))
	if err != nil || retry < 0 {
		return EKlaim{}, fmt.Errorf("EKLAIM_RETRY harus berupa angka positif")
	}

	return EKlaim{
		BaseURL:   valueOrDefault("EKLAIM_BASE_URL", valueOrDefault("EKLAM_BASE_URL", "https://api-eklaim.kemkes.go.id")),
		WSURL:     strings.TrimSpace(os.Getenv("INACBG_URL_WS")),
		Key:       firstFilled("INACBG_KEY", "INACBGKEY", "EKLAIM_INACBG_KEY"),
		KodeTarif: valueOrDefault("INACBG_KELAS", "BP"),
		CoderNIK:  firstFilled("INACBG_CODER_NIK", "CODER_NIK"),
		Timeout:   timeout,
		Retry:     retry,
	}, nil
}

func (c EKlaim) WSEnabled() bool {
	return strings.TrimSpace(c.WSURL) != "" && strings.TrimSpace(c.Key) != ""
}

func eklaimTimeout() (time.Duration, error) {
	raw := valueOrDefault("EKLAIM_TIMEOUT", valueOrDefault("EKLAM_TIMEOUT", "60"))
	if durasi, err := time.ParseDuration(raw); err == nil && durasi > 0 {
		return durasi, nil
	}

	detik, err := strconv.Atoi(raw)
	if err != nil || detik <= 0 {
		return 0, fmt.Errorf("EKLAIM_TIMEOUT harus berupa durasi positif, contoh 60 atau 60s")
	}
	return time.Duration(detik) * time.Second, nil
}

func firstFilled(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}
