package beranda

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"simrs-backend/internal/modules/bpjs"
)

type StatusKoneksiBPJS struct {
	Dikonfigurasi bool  `json:"dikonfigurasi"`
	Terhubung     bool  `json:"terhubung"`
	LatensiMS     int64 `json:"latensi_ms"`
}

type PemeriksaBPJS struct {
	url           string
	consumerID    string
	secretKey     string
	userKey       string
	dikonfigurasi bool
	klienHTTP     *http.Client
}

func NewPemeriksaBPJS(url string, consumerID string, secretKey string, userKey string) *PemeriksaBPJS {
	return &PemeriksaBPJS{
		url:        strings.TrimRight(strings.TrimSpace(url), "/"),
		consumerID: strings.TrimSpace(consumerID),
		secretKey:  strings.TrimSpace(secretKey),
		userKey:    strings.TrimSpace(userKey),
		dikonfigurasi: strings.TrimSpace(url) != "" &&
			strings.TrimSpace(consumerID) != "" &&
			strings.TrimSpace(secretKey) != "" &&
			strings.TrimSpace(userKey) != "",
		// Samakan dengan timeout layanan Data Klaim. Respons VClaim pada jaringan
		// rumah sakit kadang lebih dari empat detik meskipun akhirnya berhasil.
		klienHTTP: &http.Client{Timeout: 30 * time.Second},
	}
}

// Periksa memakai endpoint monitoring klaim yang sama dengan modul BPJS agar
// gateway menerima metode dan header resmi. Hasilnya tidak disimpan dan tidak
// mengubah data peserta maupun transaksi BPJS.
func (p *PemeriksaBPJS) Periksa(ctx context.Context) StatusKoneksiBPJS {
	status := StatusKoneksiBPJS{Dikonfigurasi: p != nil && p.dikonfigurasi}
	if p == nil || !p.dikonfigurasi {
		return status
	}

	mulai := time.Now()
	timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	header, err := bpjs.BuatHeader(p.consumerID, p.secretKey, p.userKey, timestamp)
	if err != nil {
		return status
	}

	alamat := p.url +
		"/Monitoring/Klaim/Tanggal/" + url.PathEscape(time.Now().Format("2006-01-02")) +
		"/JnsPelayanan/1/Status/3"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, alamat, nil)
	if err != nil {
		return status
	}
	request.Header = header

	response, err := p.klienHTTP.Do(request)
	status.LatensiMS = time.Since(mulai).Milliseconds()
	if err != nil {
		return status
	}
	defer response.Body.Close()

	status.Terhubung = response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices
	return status
}
