package whatsapp_gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrBelumDikonfigurasi = errors.New("WhatsApp Gateway belum dikonfigurasi")

type ErrorGateway struct {
	Status int
	Pesan  string
}

func (e *ErrorGateway) Error() string { return e.Pesan }

type Layanan struct {
	url   string
	key   string
	aktif bool
	klien *http.Client
}

func NewLayanan(alamat, key string, timeout time.Duration) *Layanan {
	alamat = strings.TrimRight(strings.TrimSpace(alamat), "/")
	key = strings.TrimSpace(key)
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Layanan{url: alamat, key: key, aktif: alamat != "" && key != "", klien: &http.Client{Timeout: timeout}}
}

func (l *Layanan) StatusKonfigurasi() map[string]any {
	return map[string]any{"dikonfigurasi": l != nil && l.aktif, "api_key_aman_di_backend": true}
}

func (l *Layanan) DaftarPerangkat(ctx context.Context) (any, error) {
	return l.json(ctx, http.MethodGet, "/api/v1/devices", nil, nil)
}

func (l *Layanan) BuatPerangkat(ctx context.Context, nama string) (any, error) {
	return l.json(ctx, http.MethodPost, "/api/v1/devices", map[string]any{"name": nama}, nil)
}

func (l *Layanan) DetailPerangkat(ctx context.Context, id string) (any, error) {
	return l.json(ctx, http.MethodGet, "/api/v1/devices/"+url.PathEscape(id), nil, nil)
}

func (l *Layanan) HubungkanPerangkat(ctx context.Context, id string) (any, error) {
	return l.json(ctx, http.MethodPost, "/api/v1/devices/"+url.PathEscape(id)+"/connect", nil, nil)
}

func (l *Layanan) KodePasangan(ctx context.Context, id, nomor string) (any, error) {
	return l.json(ctx, http.MethodPost, "/api/v1/devices/"+url.PathEscape(id)+"/pair-code", map[string]any{"phone": nomor}, nil)
}

func (l *Layanan) QR(ctx context.Context, id string) ([]byte, string, error) {
	response, err := l.request(ctx, http.MethodGet, "/api/v1/devices/"+url.PathEscape(id)+"/qr.png", nil, map[string]string{"Accept": "image/png"})
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return nil, "", fmt.Errorf("QR WhatsApp tidak dapat dibaca")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, "", bacaErrorGateway(response.StatusCode, body)
	}
	return body, response.Header.Get("X-QR-Expires-At"), nil
}

func (l *Layanan) DaftarPesan(ctx context.Context, deviceID, status string, limit int) (any, error) {
	if limit < 1 || limit > 100 {
		limit = 25
	}
	query := url.Values{"limit": {fmt.Sprint(limit)}, "offset": {"0"}}
	if strings.TrimSpace(deviceID) != "" {
		query.Set("device_id", strings.TrimSpace(deviceID))
	}
	if strings.TrimSpace(status) != "" {
		query.Set("status", strings.TrimSpace(status))
	}
	return l.json(ctx, http.MethodGet, "/api/v1/messages?"+query.Encode(), nil, nil)
}

type PesanTeks struct {
	DeviceID     string
	Tujuan       string
	NamaPenerima string
	ExternalID   string
	Isi          string
	Idempotency  string
}

func (l *Layanan) KirimPesanTeks(ctx context.Context, pesan PesanTeks) (any, error) {
	body := map[string]any{
		"device_id": pesan.DeviceID,
		"to":        pesan.Tujuan,
		"type":      "text",
		"text":      map[string]any{"body": pesan.Isi},
	}
	if pesan.NamaPenerima != "" {
		body["recipient_name"] = pesan.NamaPenerima
	}
	if pesan.ExternalID != "" {
		body["external_id"] = pesan.ExternalID
	}
	hasil, err := l.json(ctx, http.MethodPost, "/api/v1/messages", body, map[string]string{"Idempotency-Key": pesan.Idempotency})
	if err != nil {
		return nil, err
	}
	return map[string]any{"hasil": hasil, "idempotency_key": pesan.Idempotency}, nil
}

func (l *Layanan) json(ctx context.Context, method, path string, body any, headers map[string]string) (any, error) {
	response, err := l.request(ctx, method, path, body, headers)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("respons WhatsApp Gateway tidak dapat dibaca")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, bacaErrorGateway(response.StatusCode, raw)
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return map[string]any{}, nil
	}
	var envelope map[string]any
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("format respons WhatsApp Gateway tidak valid")
	}
	if data, ada := envelope["data"]; ada {
		return data, nil
	}
	return envelope, nil
}

func (l *Layanan) request(ctx context.Context, method, path string, body any, headers map[string]string) (*http.Response, error) {
	if l == nil || !l.aktif {
		return nil, ErrBelumDikonfigurasi
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	request, err := http.NewRequestWithContext(ctx, method, l.url+path, reader)
	if err != nil {
		return nil, fmt.Errorf("request WhatsApp Gateway tidak dapat dibuat")
	}
	request.Header.Set("Authorization", "Bearer "+l.key)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "SIRAPI-WHATSAPP-GATEWAY/1.0")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response, err := l.klien.Do(request)
	if err != nil {
		return nil, fmt.Errorf("WhatsApp Gateway tidak dapat dihubungi: %w", err)
	}
	return response, nil
}

func bacaErrorGateway(status int, raw []byte) error {
	pesan := "WhatsApp Gateway menolak request"
	var payload map[string]any
	if json.Unmarshal(raw, &payload) == nil {
		if errorBody, ok := payload["error"].(map[string]any); ok {
			if nilai, ok := errorBody["message"].(string); ok && strings.TrimSpace(nilai) != "" {
				pesan = nilai
			}
		} else if nilai, ok := payload["message"].(string); ok && strings.TrimSpace(nilai) != "" {
			pesan = nilai
		}
	}
	if strings.TrimSpace(pesan) == "" {
		pesan = http.StatusText(status)
	}
	return &ErrorGateway{Status: status, Pesan: pesan}
}

func NormalisasiNomor(nomor string) string {
	nomor = strings.NewReplacer(" ", "", "-", "", "+", "", "(", "", ")", "").Replace(strings.TrimSpace(nomor))
	if strings.HasPrefix(nomor, "0") {
		nomor = "62" + strings.TrimPrefix(nomor, "0")
	}
	return nomor
}
