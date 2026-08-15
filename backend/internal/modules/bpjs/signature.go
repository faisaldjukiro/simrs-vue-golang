package bpjs

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var ErrKredensialTidakLengkap = errors.New("kredensial BPJS belum lengkap")

type Konfigurasi struct {
	ConsumerID string
	SecretKey  string
	UserKey    string
}

// BuatHeader menyusun header autentikasi standar yang digunakan bersama oleh
// seluruh request servis BPJS.
func BuatHeader(consumerID, secretKey, userKey, timestamp string) (http.Header, error) {
	signature, err := BuatSignature(consumerID, secretKey, timestamp)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(userKey) == "" {
		return nil, ErrKredensialTidakLengkap
	}

	header := make(http.Header)
	header.Set("X-cons-id", strings.TrimSpace(consumerID))
	header.Set("X-timestamp", strings.TrimSpace(timestamp))
	header.Set("X-signature", signature)
	header.Set("user_key", strings.TrimSpace(userKey))
	header.Set("Content-Type", "application/json")
	header.Set("Accept", "application/json")
	header.Set("User-Agent", "SIMRS-BRIDGING-BPJS/1.0")
	return header, nil
}

type HasilSignature struct {
	Timestamp  string `json:"timestamp"`
	Signature  string `json:"signature"`
	Algoritma  string `json:"algoritma"`
	FormatData string `json:"format_data"`
}

type Layanan struct {
	konfigurasi   Konfigurasi
	waktuSekarang func() time.Time
}

func NewLayanan(konfigurasi Konfigurasi) *Layanan {
	return &Layanan{
		konfigurasi:   konfigurasi,
		waktuSekarang: time.Now,
	}
}

// BuatSignature membuat pasangan X-timestamp dan X-signature untuk header
// permintaan BPJS.
func (l *Layanan) BuatSignature() (HasilSignature, error) {
	waktu := l.waktuSekarang().UTC()
	timestamp := strconv.FormatInt(waktu.Unix(), 10)
	signature, err := BuatSignature(l.konfigurasi.ConsumerID, l.konfigurasi.SecretKey, timestamp)
	if err != nil {
		return HasilSignature{}, err
	}

	return HasilSignature{
		Timestamp:  timestamp,
		Signature:  signature,
		Algoritma:  "HMAC-SHA256 + Base64",
		FormatData: "ConsumerID&Timestamp",
	}, nil
}

// BuatSignature membuat X-signature BPJS dari Consumer ID, Secret Key, dan
// X-timestamp. Fungsi ini dipakai bersama oleh seluruh service BPJS.
func BuatSignature(consumerID, secretKey, timestamp string) (string, error) {
	consumerID = strings.TrimSpace(consumerID)
	secretKey = strings.TrimSpace(secretKey)
	timestamp = strings.TrimSpace(timestamp)
	if consumerID == "" || secretKey == "" || timestamp == "" {
		return "", ErrKredensialTidakLengkap
	}

	mac := hmac.New(sha256.New, []byte(secretKey))
	_, _ = mac.Write([]byte(consumerID + "&" + timestamp))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}
