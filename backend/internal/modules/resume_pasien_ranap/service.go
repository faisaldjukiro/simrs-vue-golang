package resume_pasien_ranap

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInputTidakValid = errors.New("input resume pasien rawat inap tidak valid")

type Layanan struct{ repo *Repositori }

func NewLayanan(repo *Repositori) *Layanan { return &Layanan{repo: repo} }

func (l *Layanan) Data(ctx context.Context, noRawat string) (Data, error) {
	noRawat = strings.TrimSpace(noRawat)
	if noRawat == "" {
		return Data{}, fmt.Errorf("%w: nomor rawat wajib diisi", ErrInputTidakValid)
	}
	return l.repo.Data(ctx, noRawat)
}

func (l *Layanan) ValidasiCoding(ctx context.Context, jenis, kode string, utama bool) (HasilValidasiCoding, error) {
	jenis = strings.ToLower(strings.TrimSpace(jenis))
	kode = normalisasiKode(kode)
	if jenis != "diagnosa" && jenis != "prosedur" {
		return HasilValidasiCoding{}, fmt.Errorf("%w: jenis coding harus diagnosa atau prosedur", ErrInputTidakValid)
	}
	if kode == "" {
		return HasilValidasiCoding{}, fmt.Errorf("%w: kode %s wajib diisi", ErrInputTidakValid, jenis)
	}
	return l.repo.ValidasiCoding(ctx, jenis, kode, utama)
}

func (l *Layanan) CariCoding(ctx context.Context, jenis, kataKunci string, utama bool) ([]HasilValidasiCoding, error) {
	jenis = strings.ToLower(strings.TrimSpace(jenis))
	kataKunci = strings.TrimSpace(kataKunci)
	if jenis != "diagnosa" && jenis != "prosedur" {
		return nil, fmt.Errorf("%w: jenis coding harus diagnosa atau prosedur", ErrInputTidakValid)
	}
	if len(kataKunci) < 2 {
		return []HasilValidasiCoding{}, nil
	}
	return l.repo.CariCoding(ctx, jenis, kataKunci, utama)
}

func (l *Layanan) Simpan(ctx context.Context, input Input) error {
	bersihkan(&input)
	if err := validasi(input); err != nil {
		return err
	}
	return l.repo.Simpan(ctx, input)
}

func (l *Layanan) Ubah(ctx context.Context, input Input) error {
	bersihkan(&input)
	if err := validasi(input); err != nil {
		return err
	}
	return l.repo.Ubah(ctx, input)
}

func (l *Layanan) Hapus(ctx context.Context, noRawat string) error {
	noRawat = strings.TrimSpace(noRawat)
	if noRawat == "" {
		return fmt.Errorf("%w: nomor rawat wajib diisi", ErrInputTidakValid)
	}
	return l.repo.Hapus(ctx, noRawat)
}

func bersihkan(i *Input) {
	i.NoRawat = strings.TrimSpace(i.NoRawat)
	i.KodeDokter = strings.TrimSpace(i.KodeDokter)
	i.Kontrol = strings.ReplaceAll(strings.TrimSpace(i.Kontrol), "T", " ")
	if len(i.Kontrol) == 16 {
		i.Kontrol += ":00"
	}
	for _, nilai := range []*string{
		&i.DiagnosaAwal, &i.Alasan, &i.KeluhanUtama, &i.PemeriksaanFisik, &i.JalannyaPenyakit,
		&i.PemeriksaanPenunjang, &i.HasilLaborat, &i.TindakanDanOperasi, &i.ObatDiRS,
		&i.DiagnosaUtama, &i.KodeDiagnosaUtama, &i.DiagnosaSekunder, &i.KodeDiagnosaSekunder,
		&i.DiagnosaSekunder2, &i.KodeDiagnosaSekunder2, &i.DiagnosaSekunder3, &i.KodeDiagnosaSekunder3,
		&i.DiagnosaSekunder4, &i.KodeDiagnosaSekunder4, &i.ProsedurUtama, &i.KodeProsedurUtama,
		&i.ProsedurSekunder, &i.KodeProsedurSekunder, &i.ProsedurSekunder2, &i.KodeProsedurSekunder2,
		&i.ProsedurSekunder3, &i.KodeProsedurSekunder3, &i.Alergi, &i.Diet, &i.LabBelum, &i.Edukasi,
		&i.CaraKeluar, &i.KeteranganKeluar, &i.Keadaan, &i.KeteranganKeadaan,
		&i.Dilanjutkan, &i.KeteranganDilanjutkan, &i.ObatPulang,
	} {
		*nilai = strings.TrimSpace(*nilai)
	}
}

func validasi(i Input) error {
	if i.NoRawat == "" {
		return fmt.Errorf("%w: nomor rawat wajib diisi", ErrInputTidakValid)
	}
	if i.KodeDokter == "" {
		return fmt.Errorf("%w: dokter penanggung jawab wajib dipilih", ErrInputTidakValid)
	}
	if i.KeluhanUtama == "" {
		return fmt.Errorf("%w: keluhan utama dan riwayat penyakit wajib diisi", ErrInputTidakValid)
	}
	if i.JalannyaPenyakit == "" {
		return fmt.Errorf("%w: jalannya penyakit selama perawatan wajib diisi", ErrInputTidakValid)
	}
	if i.KodeDiagnosaUtama == "" || i.DiagnosaUtama == "" {
		return fmt.Errorf("%w: diagnosa utama wajib diisi", ErrInputTidakValid)
	}
	if _, err := time.Parse("2006-01-02 15:04:05", i.Kontrol); err != nil {
		return fmt.Errorf("%w: tanggal dan jam kontrol tidak valid", ErrInputTidakValid)
	}
	if !ada(i.CaraKeluar, "Atas Izin Dokter", "Pindah RS", "Pulang Atas Permintaan Sendiri", "Lainnya") {
		return fmt.Errorf("%w: cara keluar tidak valid", ErrInputTidakValid)
	}
	if !ada(i.Keadaan, "Membaik", "Sembuh", "Keadaan Khusus", "Meninggal") {
		return fmt.Errorf("%w: keadaan pulang tidak valid", ErrInputTidakValid)
	}
	if !ada(i.Dilanjutkan, "Kembali Ke RS", "RS Lain", "Dokter Luar", "Puskesmes", "Lainnya") {
		return fmt.Errorf("%w: layanan lanjutan tidak valid", ErrInputTidakValid)
	}
	return nil
}

func ada(nilai string, pilihan ...string) bool {
	for _, item := range pilihan {
		if nilai == item {
			return true
		}
	}
	return false
}
