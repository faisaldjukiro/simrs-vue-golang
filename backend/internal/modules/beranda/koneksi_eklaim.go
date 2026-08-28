package beranda

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type StatusKoneksiEKlaim struct {
	Dikonfigurasi bool  `json:"dikonfigurasi"`
	Terhubung     bool  `json:"terhubung"`
	LatensiMS     int64 `json:"latensi_ms"`
}

type PemeriksaEKlaim struct {
	url           string
	dikonfigurasi bool
	klienHTTP     *http.Client
}

func NewPemeriksaEKlaim(url string, key string) *PemeriksaEKlaim {
	return &PemeriksaEKlaim{
		url:           strings.TrimSpace(url),
		dikonfigurasi: strings.TrimSpace(url) != "" && strings.TrimSpace(key) != "",
		klienHTTP: &http.Client{
			Timeout: 4 * time.Second,
		},
	}
}

// Periksa hanya mengecek apakah endpoint E-Klaim dapat dijangkau. Request HEAD
// tidak membuat atau mengubah data klaim pasien.
func (p *PemeriksaEKlaim) Periksa(ctx context.Context) StatusKoneksiEKlaim {
	status := StatusKoneksiEKlaim{Dikonfigurasi: p != nil && p.dikonfigurasi}
	if p == nil || !p.dikonfigurasi {
		return status
	}

	mulai := time.Now()
	request, err := http.NewRequestWithContext(ctx, http.MethodHead, p.url, nil)
	if err != nil {
		return status
	}

	response, err := p.klienHTTP.Do(request)
	status.LatensiMS = time.Since(mulai).Milliseconds()
	if err != nil {
		return status
	}
	defer response.Body.Close()

	// Respons 4xx tetap membuktikan endpoint dapat dijangkau; sebagian server
	// E-Klaim memang menolak metode HEAD karena operasi resminya memakai POST.
	status.Terhubung = response.StatusCode < http.StatusInternalServerError
	return status
}
