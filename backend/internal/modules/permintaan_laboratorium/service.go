package permintaan_laboratorium

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInputTidakValid = errors.New("input permintaan laboratorium tidak valid")

type Layanan struct{ repo *Repositori }

func NewLayanan(repo *Repositori) *Layanan { return &Layanan{repo: repo} }

func (l *Layanan) Data(ctx context.Context, noRawat, kategori string) (Data, error) {
	noRawat = strings.TrimSpace(noRawat)
	if noRawat == "" {
		return Data{}, fmt.Errorf("%w: nomor rawat wajib diisi", ErrInputTidakValid)
	}
	repo, err := l.repo.untukKategori(kategori)
	if err != nil {
		return Data{}, err
	}
	return repo.Data(ctx, noRawat)
}

func (l *Layanan) CariDokter(ctx context.Context, kata string) ([]Dokter, error) {
	if len(strings.TrimSpace(kata)) < 2 {
		return []Dokter{}, nil
	}
	return l.repo.CariDokter(ctx, kata)
}

func (l *Layanan) CariTindakan(ctx context.Context, noRawat, kata, kategori string) ([]Tindakan, error) {
	repo, err := l.repo.untukKategori(kategori)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(noRawat) == "" || len(strings.TrimSpace(kata)) < 2 {
		return []Tindakan{}, nil
	}
	return repo.CariTindakan(ctx, noRawat, kata)
}

func (l *Layanan) DetailTindakan(ctx context.Context, noRawat, kode, kategori string) ([]DetailPemeriksaan, error) {
	noRawat, kode = strings.TrimSpace(noRawat), strings.TrimSpace(kode)
	if noRawat == "" || kode == "" {
		return []DetailPemeriksaan{}, fmt.Errorf("%w: nomor rawat dan kode pemeriksaan wajib diisi", ErrInputTidakValid)
	}
	repo, err := l.repo.untukKategori(kategori)
	if err != nil {
		return nil, err
	}
	return repo.DetailTindakan(ctx, noRawat, kode)
}

func (l *Layanan) Simpan(ctx context.Context, input Input) (string, error) {
	bersihkan(&input)
	if err := validasi(input); err != nil {
		return "", err
	}
	repo, err := l.repo.untukKategori(input.Kategori)
	if err != nil {
		return "", err
	}
	return repo.Simpan(ctx, input)
}

func (l *Layanan) Ubah(ctx context.Context, nomor string, input Input) error {
	nomor = strings.TrimSpace(nomor)
	bersihkan(&input)
	if nomor == "" {
		return fmt.Errorf("%w: nomor permintaan wajib diisi", ErrInputTidakValid)
	}
	if err := validasi(input); err != nil {
		return err
	}
	repo, err := l.repo.untukKategori(input.Kategori)
	if err != nil {
		return err
	}
	return repo.Ubah(ctx, nomor, input)
}

func (l *Layanan) Hapus(ctx context.Context, noRawat, nomor, kategori string) error {
	noRawat, nomor = strings.TrimSpace(noRawat), strings.TrimSpace(nomor)
	if noRawat == "" || nomor == "" {
		return fmt.Errorf("%w: nomor rawat dan nomor permintaan wajib diisi", ErrInputTidakValid)
	}
	repo, err := l.repo.untukKategori(kategori)
	if err != nil {
		return err
	}
	return repo.Hapus(ctx, noRawat, nomor)
}

func bersihkan(input *Input) {
	input.Kategori = strings.ToUpper(strings.TrimSpace(input.Kategori))
	if input.Kategori == "" {
		input.Kategori = "PK"
	}
	p := &input.Spesimen
	p.PengambilanBahan = strings.TrimSpace(p.PengambilanBahan)
	p.DiperolehDengan = strings.TrimSpace(p.DiperolehDengan)
	p.LokasiJaringan = strings.TrimSpace(p.LokasiJaringan)
	p.DiawetkanDengan = strings.TrimSpace(p.DiawetkanDengan)
	p.PernahDilakukanDi = strings.TrimSpace(p.PernahDilakukanDi)
	p.TanggalPASebelumnya = strings.TrimSpace(p.TanggalPASebelumnya)
	p.NomorPASebelumnya = strings.TrimSpace(p.NomorPASebelumnya)
	p.DiagnosaPASebelumnya = strings.TrimSpace(p.DiagnosaPASebelumnya)
	input.NoRawat = strings.TrimSpace(input.NoRawat)
	input.Tanggal = strings.TrimSpace(input.Tanggal)
	input.Jam = strings.TrimSpace(input.Jam)
	if len(input.Jam) == 5 {
		input.Jam += ":00"
	}
	input.KodeDokter = strings.TrimSpace(input.KodeDokter)
	input.InformasiTambahan = strings.TrimSpace(input.InformasiTambahan)
	input.DiagnosisKlinis = strings.TrimSpace(input.DiagnosisKlinis)
	unik := make([]PilihanPemeriksaan, 0, len(input.Pemeriksaan))
	ada := make(map[string]bool)
	for _, pemeriksaan := range input.Pemeriksaan {
		pemeriksaan.Kode = strings.TrimSpace(pemeriksaan.Kode)
		if pemeriksaan.Kode != "" && !ada[pemeriksaan.Kode] {
			ada[pemeriksaan.Kode] = true
			detail := make([]int64, 0, len(pemeriksaan.IDDetail))
			adaDetail := make(map[int64]bool)
			for _, id := range pemeriksaan.IDDetail {
				if id > 0 && !adaDetail[id] {
					adaDetail[id] = true
					detail = append(detail, id)
				}
			}
			pemeriksaan.IDDetail = detail
			unik = append(unik, pemeriksaan)
		}
	}
	input.Pemeriksaan = unik
}

func validasi(input Input) error {
	if _, err := kategoriValid(input.Kategori); err != nil {
		return err
	}
	if input.Kategori == "PA" {
		p := input.Spesimen
		if _, err := time.Parse("2006-01-02", p.PengambilanBahan); err != nil {
			return fmt.Errorf("%w: tanggal pengambilan bahan PA wajib diisi dengan benar", ErrInputTidakValid)
		}
		if p.PernahDilakukanDi != "" {
			if _, err := time.Parse("2006-01-02", p.TanggalPASebelumnya); err != nil {
				return fmt.Errorf("%w: tanggal PA sebelumnya wajib diisi", ErrInputTidakValid)
			}
		}
		if utf8.RuneCountInString(p.DiperolehDengan) > 40 || utf8.RuneCountInString(p.LokasiJaringan) > 40 || utf8.RuneCountInString(p.DiawetkanDengan) > 40 || utf8.RuneCountInString(p.PernahDilakukanDi) > 100 || utf8.RuneCountInString(p.NomorPASebelumnya) > 20 || utf8.RuneCountInString(p.DiagnosaPASebelumnya) > 100 {
			return fmt.Errorf("%w: panjang informasi spesimen PA melebihi batas", ErrInputTidakValid)
		}
		for _, pemeriksaan := range input.Pemeriksaan {
			if len(pemeriksaan.IDDetail) > 0 {
				return fmt.Errorf("%w: pemeriksaan PA tidak memakai detail template PK/MB", ErrInputTidakValid)
			}
		}
	}
	if input.NoRawat == "" || input.KodeDokter == "" {
		return fmt.Errorf("%w: nomor rawat dan dokter perujuk wajib diisi", ErrInputTidakValid)
	}
	if input.InformasiTambahan == "" || input.DiagnosisKlinis == "" {
		return fmt.Errorf("%w: informasi tambahan dan diagnosis klinis wajib diisi", ErrInputTidakValid)
	}
	if utf8.RuneCountInString(input.InformasiTambahan) > 60 || utf8.RuneCountInString(input.DiagnosisKlinis) > 80 {
		return fmt.Errorf("%w: informasi tambahan maksimal 60 karakter dan diagnosis klinis maksimal 80 karakter", ErrInputTidakValid)
	}
	if len(input.Pemeriksaan) == 0 {
		return fmt.Errorf("%w: minimal satu pemeriksaan laboratorium wajib dipilih", ErrInputTidakValid)
	}
	if len(input.Pemeriksaan) > 100 {
		return fmt.Errorf("%w: maksimal 100 pemeriksaan dalam satu permintaan", ErrInputTidakValid)
	}
	if _, err := time.Parse("2006-01-02", input.Tanggal); err != nil {
		return fmt.Errorf("%w: tanggal permintaan tidak valid", ErrInputTidakValid)
	}
	if _, err := time.Parse("15:04:05", input.Jam); err != nil {
		return fmt.Errorf("%w: jam permintaan tidak valid", ErrInputTidakValid)
	}
	return nil
}
