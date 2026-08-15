package config

import (
	"fmt"
	"os"
	"strings"
)

// BPJS menyimpan kredensial bersama yang dipakai untuk membuat header
// autentikasi layanan BPJS. Nilai rahasia hanya digunakan di backend.
type BPJS struct {
	ConsumerID string
	SecretKey  string
	UserKey    string
	VClaimURL  string
}

func BPJSConfig() (BPJS, error) {
	konfigurasi := BPJS{
		ConsumerID: strings.TrimSpace(os.Getenv("BPJS_CONS_ID")),
		SecretKey:  strings.TrimSpace(os.Getenv("BPJS_SECRET")),
		UserKey:    strings.TrimSpace(os.Getenv("BPJS_USER_KEY")),
		VClaimURL:  strings.TrimRight(strings.TrimSpace(os.Getenv("BPJS_VCLAIM_URL")), "/"),
	}

	var kosong []string
	if konfigurasi.ConsumerID == "" {
		kosong = append(kosong, "BPJS_CONS_ID")
	}
	if konfigurasi.SecretKey == "" {
		kosong = append(kosong, "BPJS_SECRET")
	}
	if konfigurasi.UserKey == "" {
		kosong = append(kosong, "BPJS_USER_KEY")
	}
	if konfigurasi.VClaimURL == "" {
		kosong = append(kosong, "BPJS_VCLAIM_URL")
	}
	if len(kosong) > 0 {
		return BPJS{}, fmt.Errorf("konfigurasi BPJS belum lengkap: %s", strings.Join(kosong, ", "))
	}

	return konfigurasi, nil
}
