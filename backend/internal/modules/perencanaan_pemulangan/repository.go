package perencanaan_pemulangan

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"simrs-backend/internal/shared/khanzamutasi"
)

var ErrValidasi = errors.New("perencanaan pemulangan tidak valid")
var ErrAkses = errors.New("hanya petugas pencatat atau administrator yang boleh mengubah catatan")

const tabel = "perencanaan_pemulangan"

var Penilaian = []string{"pengaruh_ri_pasien_dan_keluarga", "pengaruh_ri_pekerjaan_sekolah", "pengaruh_ri_keuangan", "antisipasi_masalah_saat_pulang", "bantuan_diperlukan_dalam", "adakah_yang_membantu_keperluan", "pasien_tinggal_sendiri", "pasien_menggunakan_peralatan_medis", "pasien_memerlukan_alat_bantu", "memerlukan_perawatan_khusus", "bermasalah_memenuhi_kebutuhan", "memiliki_nyeri_kronis", "memerlukan_edukasi_kesehatan", "memerlukan_keterampilkan_khusus"}
var Bantuan = []string{"Menyiapkan Makanan", "Edukasi Kesehatan", "Makan", "Mandi", "Diet", "Berpakaian", "Menyiapkan Obat", "Transportasi", "Minum Obat"}

type Input struct {
	NoRawat string            `json:"no_rawat"`
	Data    map[string]string `json:"data"`
	Asli    map[string]string `json:"asli"`
}
type Catatan struct {
	Data        map[string]string `json:"data"`
	NamaPetugas string            `json:"nama_petugas"`
	BisaUbah    bool              `json:"bisa_ubah"`
}
type Pilihan struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}
type Hasil struct {
	Catatan           []Catatan `json:"catatan"`
	PetugasLogin      Pilihan   `json:"petugas_login"`
	BolehPilihPetugas bool      `json:"boleh_pilih_petugas"`
}
type Repositori struct{ simrs *sql.DB }

func NewRepositori(simrs *sql.DB) *Repositori { return &Repositori{simrs: simrs} }
func Kolom() []string {
	k := []string{"no_rawat", "rencana_pulang", "alasan_masuk", "diagnosa_medis"}
	for _, p := range Penilaian {
		k = append(k, p, "keterangan_"+p)
	}
	return append(k, "nama_pasien_keluarga", "nip")
}
func Validasi(in *Input) error {
	if strings.TrimSpace(in.NoRawat) == "" || len(in.NoRawat) > 17 || in.Data == nil {
		return fmt.Errorf("%w: nomor rawat dan form wajib diisi", ErrValidasi)
	}
	if _, err := time.Parse("2006-01-02", in.Data["rencana_pulang"]); err != nil {
		return fmt.Errorf("%w: tanggal rencana pulang tidak valid", ErrValidasi)
	}
	for k, n := range map[string]int{"alasan_masuk": 150, "diagnosa_medis": 50, "nama_pasien_keluarga": 50, "nip": 20} {
		in.Data[k] = strings.TrimSpace(in.Data[k])
		if in.Data[k] == "" || utf8.RuneCountInString(in.Data[k]) > n {
			return fmt.Errorf("%w: %s wajib diisi, maksimal %d karakter", ErrValidasi, k, n)
		}
	}
	for _, p := range Penilaian {
		opsi := []string{"Tidak", "Ya"}
		if p == "bantuan_diperlukan_dalam" {
			opsi = Bantuan
		}
		valid := false
		for _, v := range opsi {
			if in.Data[p] == v {
				valid = true
			}
		}
		if !valid {
			return fmt.Errorf("%w: pilihan %s tidak sesuai", ErrValidasi, p)
		}
		if utf8.RuneCountInString(in.Data["keterangan_"+p]) > 100 {
			return fmt.Errorf("%w: keterangan %s maksimal 100 karakter", ErrValidasi, p)
		}
	}
	return nil
}
func (r *Repositori) Referensi(ctx context.Context, q string) ([]Pilihan, error) {
	rows, err := r.simrs.QueryContext(ctx, "SELECT nip,nama FROM petugas WHERE nip LIKE ? OR nama LIKE ? ORDER BY nama LIMIT 30", "%"+q+"%", "%"+q+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hasil := []Pilihan{}
	for rows.Next() {
		var p Pilihan
		if err := rows.Scan(&p.Kode, &p.Nama); err != nil {
			return nil, err
		}
		hasil = append(hasil, p)
	}
	return hasil, rows.Err()
}
func (r *Repositori) Daftar(ctx context.Context, no, user string, admin bool) (Hasil, error) {
	h := Hasil{Catatan: []Catatan{}, BolehPilihPetugas: admin}
	if no == "" || len(no) > 17 {
		return h, ErrValidasi
	}
	err := r.simrs.QueryRowContext(ctx, "SELECT nip,nama FROM petugas WHERE nip=?", user).Scan(&h.PetugasLogin.Kode, &h.PetugasLogin.Nama)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return h, err
	}
	selects := []string{}
	for _, k := range Kolom() {
		if k == "rencana_pulang" {
			selects = append(selects, "DATE_FORMAT(c.rencana_pulang,'%Y-%m-%d')")
		} else {
			selects = append(selects, "COALESCE(c."+k+",'')")
		}
	}
	rows, err := r.simrs.QueryContext(ctx, "SELECT "+strings.Join(selects, ",")+",COALESCE(p.nama,'') FROM "+tabel+" c LEFT JOIN petugas p ON p.nip=c.nip WHERE c.no_rawat=?", no)
	if err != nil {
		return h, err
	}
	defer rows.Close()
	for rows.Next() {
		nilai := make([]string, len(selects)+1)
		dest := make([]any, len(nilai))
		for i := range nilai {
			dest[i] = &nilai[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return h, err
		}
		c := Catatan{Data: map[string]string{}, NamaPetugas: nilai[len(selects)]}
		for i, k := range Kolom() {
			c.Data[k] = nilai[i]
		}
		c.BisaUbah = admin || c.Data["nip"] == user
		h.Catatan = append(h.Catatan, c)
	}
	return h, rows.Err()
}
func (r *Repositori) Mutasi(ctx context.Context, in Input, method, user string, admin bool) error {
	if in.NoRawat == "" || len(in.NoRawat) > 17 {
		return ErrValidasi
	}
	if method != "POST" && method != "PUT" && method != "DELETE" {
		return ErrValidasi
	}
	if method != "DELETE" {
		if err := Validasi(&in); err != nil {
			return err
		}
		if !admin && in.Data["nip"] != user {
			return ErrAkses
		}
	}
	if method != "POST" {
		for _, k := range Kolom() {
			if _, ok := in.Asli[k]; !ok {
				return fmt.Errorf("%w: muat ulang catatan sebelum mengubah", ErrValidasi)
			}
		}
		if in.Asli["no_rawat"] != in.NoRawat {
			return ErrValidasi
		}
		if !admin && in.Asli["nip"] != user {
			return ErrAkses
		}
	}
	if method != "DELETE" {
		var ada bool
		if err := r.simrs.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM reg_periksa WHERE no_rawat=?) AND EXISTS(SELECT 1 FROM petugas WHERE nip=?)", in.NoRawat, in.Data["nip"]).Scan(&ada); err != nil {
			return err
		}
		if !ada {
			return fmt.Errorf("%w: kunjungan atau petugas tidak ditemukan", ErrValidasi)
		}
	}
	lama, baru := []any{}, []any{}
	for _, k := range Kolom() {
		lama = append(lama, in.Asli[k])
		if k == "no_rawat" {
			baru = append(baru, in.NoRawat)
		} else {
			baru = append(baru, in.Data[k])
		}
	}
	if method == "POST" {
		_, err := r.simrs.ExecContext(ctx, "INSERT INTO "+tabel+" ("+strings.Join(Kolom(), ",")+") VALUES ("+strings.TrimSuffix(strings.Repeat("?,", len(Kolom())), ",")+")", baru...)
		return err
	}
	return khanzamutasi.Jalankan(ctx, r.simrs, tabel, Kolom(), lama, baru, method == "DELETE")
}
