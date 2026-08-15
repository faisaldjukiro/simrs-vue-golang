package data_klaim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"simrs-backend/internal/modules/bpjs"
)

var (
	ErrInputTidakValid     = errors.New("parameter monitoring klaim tidak valid")
	ErrVClaimTidakTersedia = errors.New("servis VClaim tidak dapat dihubungi")
	ErrResponsTidakValid   = errors.New("respons VClaim tidak valid")
)

type Konfigurasi struct {
	ConsumerID string
	SecretKey  string
	UserKey    string
	BaseURL    string
}

type Input struct {
	TanggalPulang  string
	TanggalMulai   string
	TanggalSelesai string
	JenisPelayanan string
	StatusKlaim    string
}

type MetaData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Hasil struct {
	MetaData         MetaData       `json:"metaData"`
	Response         any            `json:"response,omitempty"`
	Periode          *Periode       `json:"periode,omitempty"`
	TanggalTanpaData []string       `json:"tanggal_tanpa_data,omitempty"`
	TanggalGagal     []TanggalGagal `json:"tanggal_gagal,omitempty"`
}

type Periode struct {
	TanggalMulai           string `json:"tanggal_mulai"`
	TanggalSelesai         string `json:"tanggal_selesai"`
	JumlahHari             int    `json:"jumlah_hari"`
	JumlahTanggalBerhasil  int    `json:"jumlah_tanggal_berhasil"`
	JumlahTanggalTanpaData int    `json:"jumlah_tanggal_tanpa_data"`
	JumlahTanggalGagal     int    `json:"jumlah_tanggal_gagal"`
}

type TanggalGagal struct {
	Tanggal string `json:"tanggal"`
	Pesan   string `json:"pesan"`
}

type KesalahanBPJS struct {
	Code    string
	Message string
}

func (e *KesalahanBPJS) Error() string {
	return fmt.Sprintf("VClaim [%s]: %s", e.Code, e.Message)
}

type httpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

type Layanan struct {
	konfigurasi   Konfigurasi
	client        httpClient
	waktuSekarang func() time.Time
}

func NewLayanan(konfigurasi Konfigurasi) *Layanan {
	return &Layanan{
		konfigurasi: konfigurasi,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		waktuSekarang: time.Now,
	}
}

func (l *Layanan) Data(ctx context.Context, input Input) (Hasil, error) {
	periode, err := validasi(input)
	if err != nil {
		return Hasil{}, err
	}

	if periode == nil {
		return l.dataPerTanggalDenganPercobaanUlang(ctx, strings.TrimSpace(input.TanggalPulang), input)
	}

	semuaKlaim := make([]any, 0)
	tanggalTanpaData := make([]string, 0)
	tanggalGagal := make([]TanggalGagal, 0)
	jumlahTanggalBerhasil := 0
	for tanggal := periode.mulai; !tanggal.After(periode.selesai); tanggal = tanggal.AddDate(0, 0, 1) {
		tanggalTeks := tanggal.Format("2006-01-02")
		hasil, err := l.dataPerTanggal(ctx, tanggalTeks, input)
		if err != nil {
			if dataTidakDitemukan(err) {
				tanggalTanpaData = append(tanggalTanpaData, tanggalTeks)
				continue
			}
			if gangguanSementara(err) {
				tanggalGagal = append(tanggalGagal, TanggalGagal{
					Tanggal: tanggalTeks,
					Pesan:   err.Error(),
				})
				continue
			}
			return Hasil{}, fmt.Errorf("data klaim tanggal %s: %w", tanggalTeks, err)
		}
		jumlahTanggalBerhasil++

		klaim, err := ekstrakKlaim(hasil.Response)
		if err != nil {
			return Hasil{}, err
		}
		semuaKlaim = append(semuaKlaim, klaim...)
	}

	pesan := "Sukses"
	if len(tanggalGagal) > 0 {
		pesan = "Sukses dengan sebagian tanggal gagal diproses VClaim"
	}

	return Hasil{
		MetaData:         MetaData{Code: "200", Message: pesan},
		Response:         map[string]any{"klaim": semuaKlaim},
		TanggalTanpaData: tanggalTanpaData,
		TanggalGagal:     tanggalGagal,
		Periode: &Periode{
			TanggalMulai:           periode.mulai.Format("2006-01-02"),
			TanggalSelesai:         periode.selesai.Format("2006-01-02"),
			JumlahHari:             int(periode.selesai.Sub(periode.mulai).Hours()/24) + 1,
			JumlahTanggalBerhasil:  jumlahTanggalBerhasil,
			JumlahTanggalTanpaData: len(tanggalTanpaData),
			JumlahTanggalGagal:     len(tanggalGagal),
		},
	}, nil
}

func (l *Layanan) dataPerTanggalDenganPercobaanUlang(ctx context.Context, tanggalPulang string, input Input) (Hasil, error) {
	const maksimalPercobaan = 3

	var errTerakhir error
	for percobaan := 1; percobaan <= maksimalPercobaan; percobaan++ {
		hasil, err := l.dataPerTanggal(ctx, tanggalPulang, input)
		if err == nil {
			return hasil, nil
		}
		errTerakhir = err
		if !gangguanSementara(err) || percobaan == maksimalPercobaan {
			break
		}

		waktuTunggu := time.Duration(percobaan) * 500 * time.Millisecond
		timer := time.NewTimer(waktuTunggu)
		select {
		case <-ctx.Done():
			timer.Stop()
			return Hasil{}, ctx.Err()
		case <-timer.C:
		}
	}

	return Hasil{}, fmt.Errorf("VClaim tetap menolak setelah %d percobaan: %w", maksimalPercobaan, errTerakhir)
}

func gangguanSementara(err error) bool {
	var kesalahanBPJS *KesalahanBPJS
	if !errors.As(err, &kesalahanBPJS) {
		return false
	}
	pesan := strings.ToLower(kesalahanBPJS.Message)
	return kesalahanBPJS.Code == "404" && strings.Contains(pesan, "coba lagi nanti")
}

func dataTidakDitemukan(err error) bool {
	var kesalahanBPJS *KesalahanBPJS
	if !errors.As(err, &kesalahanBPJS) {
		return false
	}
	pesan := strings.ToLower(kesalahanBPJS.Message)
	return strings.Contains(pesan, "data tidak ditemukan") ||
		strings.Contains(pesan, "data klaim tidak ditemukan")
}

func (l *Layanan) dataPerTanggal(ctx context.Context, tanggalPulang string, input Input) (Hasil, error) {

	timestamp := strconv.FormatInt(l.waktuSekarang().UTC().Unix(), 10)
	header, err := bpjs.BuatHeader(
		l.konfigurasi.ConsumerID,
		l.konfigurasi.SecretKey,
		l.konfigurasi.UserKey,
		timestamp,
	)
	if err != nil {
		return Hasil{}, err
	}

	alamat := strings.TrimRight(strings.TrimSpace(l.konfigurasi.BaseURL), "/") +
		"/Monitoring/Klaim/Tanggal/" + url.PathEscape(tanggalPulang) +
		"/JnsPelayanan/" + url.PathEscape(input.JenisPelayanan) +
		"/Status/" + url.PathEscape(input.StatusKlaim)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, alamat, nil)
	if err != nil {
		return Hasil{}, fmt.Errorf("membuat request VClaim: %w", err)
	}
	request.Header = header

	response, err := l.client.Do(request)
	if err != nil {
		return Hasil{}, fmt.Errorf("%w: %v", ErrVClaimTidakTersedia, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 10<<20))
	if err != nil {
		return Hasil{}, fmt.Errorf("%w: membaca body", ErrResponsTidakValid)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return Hasil{}, fmt.Errorf("%w: HTTP %d", ErrVClaimTidakTersedia, response.StatusCode)
	}

	hasil, err := l.bacaResponse(body, timestamp)
	if err != nil {
		return Hasil{}, err
	}
	if hasil.MetaData.Code != "200" {
		return Hasil{}, &KesalahanBPJS{
			Code:    hasil.MetaData.Code,
			Message: hasil.MetaData.Message,
		}
	}
	return hasil, nil
}

type rentangTanggal struct {
	mulai   time.Time
	selesai time.Time
}

func validasi(input Input) (*rentangTanggal, error) {
	tanggalPulang := strings.TrimSpace(input.TanggalPulang)
	tanggalMulai := strings.TrimSpace(input.TanggalMulai)
	tanggalSelesai := strings.TrimSpace(input.TanggalSelesai)

	if tanggalMulai != "" || tanggalSelesai != "" {
		if tanggalMulai == "" || tanggalSelesai == "" {
			return nil, fmt.Errorf("%w: tanggal_mulai dan tanggal_selesai wajib diisi bersamaan", ErrInputTidakValid)
		}
		mulai, err := time.Parse("2006-01-02", tanggalMulai)
		if err != nil {
			return nil, fmt.Errorf("%w: tanggal_mulai wajib berformat yyyy-mm-dd", ErrInputTidakValid)
		}
		selesai, err := time.Parse("2006-01-02", tanggalSelesai)
		if err != nil {
			return nil, fmt.Errorf("%w: tanggal_selesai wajib berformat yyyy-mm-dd", ErrInputTidakValid)
		}
		if selesai.Before(mulai) {
			return nil, fmt.Errorf("%w: tanggal_selesai tidak boleh sebelum tanggal_mulai", ErrInputTidakValid)
		}
		if jumlahHari := int(selesai.Sub(mulai).Hours()/24) + 1; jumlahHari > 31 {
			return nil, fmt.Errorf("%w: rentang tanggal maksimal 31 hari", ErrInputTidakValid)
		}
		if err := validasiPilihan(input); err != nil {
			return nil, err
		}
		return &rentangTanggal{mulai: mulai, selesai: selesai}, nil
	}

	if _, err := time.Parse("2006-01-02", tanggalPulang); err != nil {
		return nil, fmt.Errorf("%w: tanggal_pulang wajib berformat yyyy-mm-dd, atau gunakan tanggal_mulai dan tanggal_selesai", ErrInputTidakValid)
	}
	if err := validasiPilihan(input); err != nil {
		return nil, err
	}
	return nil, nil
}

func validasiPilihan(input Input) error {
	if input.JenisPelayanan != "1" && input.JenisPelayanan != "2" {
		return fmt.Errorf("%w: jenis_pelayanan hanya boleh 1 (inap) atau 2 (jalan)", ErrInputTidakValid)
	}
	if input.StatusKlaim != "1" && input.StatusKlaim != "2" && input.StatusKlaim != "3" {
		return fmt.Errorf("%w: status_klaim hanya boleh 1, 2, atau 3", ErrInputTidakValid)
	}
	return nil
}

func ekstrakKlaim(response any) ([]any, error) {
	if response == nil {
		return []any{}, nil
	}
	objek, ok := response.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: struktur response klaim tidak dikenali", ErrResponsTidakValid)
	}
	klaimRaw, ada := objek["klaim"]
	if !ada || klaimRaw == nil {
		return []any{}, nil
	}
	klaim, ok := klaimRaw.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: daftar klaim tidak dikenali", ErrResponsTidakValid)
	}
	return klaim, nil
}

func (l *Layanan) bacaResponse(body []byte, timestamp string) (Hasil, error) {
	var pembungkus struct {
		MetaData MetaData        `json:"metaData"`
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(body, &pembungkus); err != nil {
		return Hasil{}, fmt.Errorf("%w: JSON tidak dapat dibaca", ErrResponsTidakValid)
	}

	hasil := Hasil{MetaData: pembungkus.MetaData}
	// Respons gagal dari BPJS umumnya tidak memiliki payload terenkripsi.
	// Kembalikan metadata lebih dahulu agar pesan asli BPJS tetap terbaca.
	if hasil.MetaData.Code != "200" {
		return hasil, nil
	}
	if len(pembungkus.Response) == 0 || string(pembungkus.Response) == "null" {
		return hasil, nil
	}

	var terenkripsi string
	if err := json.Unmarshal(pembungkus.Response, &terenkripsi); err == nil {
		plaintext, err := bpjs.DekripsiResponse(
			terenkripsi,
			l.konfigurasi.ConsumerID,
			l.konfigurasi.SecretKey,
			timestamp,
		)
		if err != nil {
			return Hasil{}, fmt.Errorf("%w: dekripsi response gagal: %v", ErrResponsTidakValid, err)
		}
		if err := json.Unmarshal(plaintext, &hasil.Response); err != nil {
			return Hasil{}, fmt.Errorf("%w: hasil dekripsi bukan JSON", ErrResponsTidakValid)
		}
		return hasil, nil
	}

	if err := json.Unmarshal(pembungkus.Response, &hasil.Response); err != nil {
		return Hasil{}, fmt.Errorf("%w: field response tidak dapat dibaca", ErrResponsTidakValid)
	}
	return hasil, nil
}
