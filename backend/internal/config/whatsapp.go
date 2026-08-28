package config

import (
	"os"
	"strings"
	"time"
)

// WhatsApp menyimpan konfigurasi WhatsApp Gateway. API key hanya digunakan
// oleh backend dan tidak pernah dikirim ke frontend.
type WhatsApp struct {
	URL     string
	Key     string
	Timeout time.Duration
}

func WhatsAppConfig() WhatsApp {
	return WhatsApp{
		URL:     strings.TrimRight(strings.TrimSpace(os.Getenv("URL_WHATSAPP")), "/"),
		Key:     strings.TrimSpace(os.Getenv("KEY_WHATSAPP")),
		Timeout: 30 * time.Second,
	}
}

func (c WhatsApp) Aktif() bool {
	return c.URL != "" && c.Key != ""
}
