package rujukan_internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	Ralan = "Ralan"
	Ranap = "Ranap"
)

var (
	ErrInput     = errors.New("data rujukan tidak valid")
	ErrTidakAda  = errors.New("kunjungan pasien tidak ditemukan")
	ErrDuplikat  = errors.New("rujukan yang sama sudah tersedia")
	ErrKunjungan = errors.New("rujukan tidak dapat disimpan: jenis perawatan tidak sesuai atau kunjungan dibatalkan")
	ErrBerubah   = errors.New("rujukan sudah berubah atau dihapus. Muat ulang sebelum melanjutkan")
	ErrBilling   = errors.New("billing sudah terverifikasi; rujukan hanya dapat dilihat")
)

type Referensi struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}

type Input struct {
	NoRawat    string `json:"no_rawat"`
	KodeDokter string `json:"kd_dokter"`
	KodePoli   string `json:"kd_poli"`
	Tanggal    string `json:"tanggal"`
	Jam        string `json:"jam"`
}

type Rujukan struct {
	Input
	ID            uint64 `json:"id"`
	NamaDokter    string `json:"nama_dokter"`
	NamaPoli      string `json:"nama_poli"`
	Sumber        string `json:"sumber"`
	KonflikKhanza bool   `json:"konflik_khanza"`
}

type Data struct {
	Daftar      []Rujukan `json:"daftar"`
	BolehSimpan bool      `json:"boleh_simpan"`
	Pesan       string    `json:"pesan"`
}

// Izin khusus user: CRUD tabel rujukan Khanza. Migration tetap hanya ke aplikasiDB.
type Repositori struct {
	aplikasiDB *sql.DB
	simrsDB    *sql.DB
}

func NewRepositori(aplikasiDB, simrsDB *sql.DB) *Repositori {
	return &Repositori{aplikasiDB: aplikasiDB, simrsDB: simrsDB}
}

func validJenis(jenis string) bool { return jenis == Ralan || jenis == Ranap }

func validNoRawat(noRawat string) bool {
	return strings.TrimSpace(noRawat) != "" && len(noRawat) <= 17
}

func Validasi(jenis string, input Input) (Input, error) {
	input.NoRawat = strings.TrimSpace(input.NoRawat)
	input.KodeDokter = strings.TrimSpace(input.KodeDokter)
	input.KodePoli = strings.TrimSpace(input.KodePoli)
	input.Tanggal = strings.TrimSpace(input.Tanggal)
	input.Jam = strings.TrimSpace(input.Jam)
	if !validJenis(jenis) || !validNoRawat(input.NoRawat) || input.KodeDokter == "" || len(input.KodeDokter) > 20 || input.KodePoli == "" || len(input.KodePoli) > 5 {
		return Input{}, fmt.Errorf("%w: kunjungan, poli dan dokter tujuan wajib dipilih", ErrInput)
	}
	if jenis == Ranap {
		if _, err := time.Parse("2006-01-02", input.Tanggal); err != nil {
			return Input{}, fmt.Errorf("%w: tanggal rujukan wajib diisi dengan benar", ErrInput)
		}
		if len(input.Jam) == 5 {
			input.Jam += ":00"
		}
		if _, err := time.Parse("15:04:05", input.Jam); err != nil {
			return Input{}, fmt.Errorf("%w: jam rujukan wajib diisi dengan benar", ErrInput)
		}
	} else {
		// Khanza tidak menyimpan tanggal/jam pada rujukan_internal_poli.
		input.Tanggal, input.Jam = "", ""
	}
	return input, nil
}

func (r *Repositori) kunjungan(ctx context.Context, noRawat, jenis string) (bool, error) {
	err := periksaKunjungan(ctx, r.simrsDB, noRawat, jenis, false)
	if errors.Is(err, ErrKunjungan) || errors.Is(err, ErrBilling) {
		return false, err
	}
	return err == nil, err
}

func (r *Repositori) Daftar(ctx context.Context, jenis, noRawat string) (Data, error) {
	noRawat = strings.TrimSpace(noRawat)
	if !validJenis(jenis) || !validNoRawat(noRawat) {
		return Data{}, ErrInput
	}
	boleh, err := r.kunjungan(ctx, noRawat, jenis)
	if err != nil && !errors.Is(err, ErrKunjungan) && !errors.Is(err, ErrBilling) {
		return Data{}, err
	}
	data := Data{Daftar: []Rujukan{}, BolehSimpan: boleh}
	if !boleh {
		data.Pesan = err.Error()
	}
	query := `SELECT 0, r.kd_dokter, COALESCE(d.nm_dokter,''), COALESCE(r.kd_poli,''), COALESCE(p.nm_poli,''), '', ''
		FROM rujukan_internal_poli r
		LEFT JOIN dokter d ON d.kd_dokter = r.kd_dokter
		LEFT JOIN poliklinik p ON p.kd_poli = r.kd_poli
		WHERE r.no_rawat = ? ORDER BY p.nm_poli, d.nm_dokter`
	if jenis == Ranap {
		query = `SELECT 0, r.kd_dokter, COALESCE(d.nm_dokter,''), COALESCE(r.kd_poli,''), COALESCE(p.nm_poli,''),
			COALESCE(DATE_FORMAT(r.tanggal,'%Y-%m-%d'),''), COALESCE(TIME_FORMAT(r.jam,'%H:%i:%s'),'')
			FROM rujukan_internal_ranap r
			LEFT JOIN dokter d ON d.kd_dokter = r.kd_dokter
			LEFT JOIN poliklinik p ON p.kd_poli = r.kd_poli
			WHERE r.no_rawat = ? ORDER BY r.tanggal DESC, r.jam DESC, r.kd_poli`
	}
	legacy, err := r.bacaDaftar(ctx, r.simrsDB, query, "Khanza", noRawat, noRawat)
	if err != nil {
		return Data{}, err
	}
	lokal, err := r.bacaDaftar(ctx, r.aplikasiDB, `SELECT id, kd_dokter, nama_dokter, kd_poli, nama_poli,
		COALESCE(DATE_FORMAT(tanggal,'%Y-%m-%d'),''), COALESCE(TIME_FORMAT(jam,'%H:%i:%s'),'')
		FROM sirapi_rujukan_internal WHERE no_rawat = ? AND jenis_rawat = ? AND dikirim_pada IS NULL ORDER BY id DESC`, "SIRAPI", noRawat, noRawat, jenis)
	if err != nil {
		return Data{}, err
	}
	for i := range lokal {
		for _, tujuan := range legacy {
			if kunciTujuan(jenis, lokal[i].Input) == kunciTujuan(jenis, tujuan.Input) && lokal[i].Input != tujuan.Input {
				lokal[i].KonflikKhanza = true
			}
		}
	}
	data.Daftar = append(legacy, lokal...)
	return data, nil
}

func (r *Repositori) bacaDaftar(ctx context.Context, db *sql.DB, query, sumber, noRawat string, args ...any) ([]Rujukan, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	data := []Rujukan{}
	for rows.Next() {
		item := Rujukan{Input: Input{NoRawat: noRawat}, Sumber: sumber}
		if err := rows.Scan(&item.ID, &item.KodeDokter, &item.NamaDokter, &item.KodePoli, &item.NamaPoli, &item.Tanggal, &item.Jam); err != nil {
			return nil, err
		}
		data = append(data, item)
	}
	return data, rows.Err()
}

func (r *Repositori) CariReferensi(ctx context.Context, jenis, kata string) ([]Referensi, error) {
	if jenis != "poli" && jenis != "dokter" {
		return nil, ErrInput
	}
	kata = strings.TrimSpace(kata)
	if len(kata) < 2 {
		return []Referensi{}, nil
	}
	if len(kata) > 100 {
		return nil, ErrInput
	}
	query := `SELECT kd_poli, nm_poli FROM poliklinik WHERE status = '1' AND (kd_poli LIKE ? OR nm_poli LIKE ?) ORDER BY nm_poli LIMIT 30`
	if jenis == "dokter" {
		query = `SELECT kd_dokter, nm_dokter FROM dokter WHERE status = '1' AND (kd_dokter LIKE ? OR nm_dokter LIKE ?) ORDER BY nm_dokter LIMIT 30`
	}
	rows, err := r.simrsDB.QueryContext(ctx, query, "%"+kata+"%", "%"+kata+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	data := []Referensi{}
	for rows.Next() {
		var item Referensi
		if err := rows.Scan(&item.Kode, &item.Nama); err != nil {
			return nil, err
		}
		data = append(data, item)
	}
	return data, rows.Err()
}
