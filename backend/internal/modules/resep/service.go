package resep

import (
	"context"
	"strings"
	"time"
)

type Layanan struct{ repositori *Repositori }

func NewLayanan(repositori *Repositori) *Layanan { return &Layanan{repositori: repositori} }

func (l *Layanan) InfoPasien(ctx context.Context, noRawat string) (InfoPasien, error) {
	if strings.TrimSpace(noRawat) == "" {
		return InfoPasien{}, inputTidakValid("nomor rawat wajib diisi")
	}
	return l.repositori.InfoPasien(ctx, strings.TrimSpace(noRawat))
}
func (l *Layanan) CariDokter(ctx context.Context, q string) ([]InfoDokter, error) {
	if len(strings.TrimSpace(q)) < 2 {
		return []InfoDokter{}, nil
	}
	return l.repositori.CariDokter(ctx, strings.TrimSpace(q))
}
func (l *Layanan) DepoDefault(ctx context.Context, noRawat, status string) (string, string) {
	return l.repositori.DepoDefault(ctx, noRawat, status)
}
func (l *Layanan) CariDepoGudang(ctx context.Context, q string) ([]DepoGudang, error) {
	if len(strings.TrimSpace(q)) < 2 {
		return []DepoGudang{}, nil
	}
	return l.repositori.CariDepoGudang(ctx, strings.TrimSpace(q))
}
func (l *Layanan) CariObat(ctx context.Context, q, kdBangsal, kelas string) ([]Obat, error) {
	if len(strings.TrimSpace(q)) < 2 {
		return []Obat{}, nil
	}
	return l.repositori.CariObat(ctx, strings.TrimSpace(q), kdBangsal, kelas)
}
func (l *Layanan) DaftarMetodeRacik(ctx context.Context) ([]MetodeRacik, error) {
	return l.repositori.DaftarMetodeRacik(ctx)
}
func (l *Layanan) AutoNomorResep(ctx context.Context, tanggal string) (string, error) {
	if strings.TrimSpace(tanggal) == "" {
		tanggal = time.Now().Format("2006-01-02")
	}
	return l.repositori.AutoNomorResep(ctx, tanggal)
}
func (l *Layanan) DaftarResep(ctx context.Context, noRawat string) ([]ResepHeader, error) {
	return l.repositori.DaftarResep(ctx, strings.TrimSpace(noRawat))
}
func (l *Layanan) DaftarResepPasien(ctx context.Context, noRKM string) ([]ResepHeader, error) {
	noRKM = strings.TrimSpace(noRKM)
	if noRKM == "" {
		return []ResepHeader{}, nil
	}
	return l.repositori.DaftarResepPasien(ctx, noRKM)
}
func (l *Layanan) DetailResep(ctx context.Context, noResep string) (ResepLengkap, error) {
	if strings.TrimSpace(noResep) == "" {
		return ResepLengkap{}, inputTidakValid("nomor resep wajib diisi")
	}
	return l.repositori.DetailResep(ctx, strings.TrimSpace(noResep))
}
func (l *Layanan) SimpanResep(ctx context.Context, in InputSimpanResep) error {
	in.NoRawat, in.NoResep, in.KdDokter = strings.TrimSpace(in.NoRawat), strings.TrimSpace(in.NoResep), strings.TrimSpace(in.KdDokter)
	if in.NoRawat == "" {
		return inputTidakValid("nomor rawat wajib diisi")
	}
	if in.NoResep == "" {
		return inputTidakValid("nomor resep wajib diisi")
	}
	if in.KdDokter == "" {
		return inputTidakValid("dokter peresep wajib diisi")
	}
	if len(in.Obat) == 0 && len(in.Racikan) == 0 {
		return inputTidakValid("resep harus berisi minimal satu obat atau racikan")
	}
	if in.TglPeresepan == "" {
		in.TglPeresepan = time.Now().Format("2006-01-02")
	}
	if in.Jam == "" {
		in.Jam = time.Now().Format("15:04:05")
	}
	return l.repositori.SimpanResep(ctx, in)
}
func (l *Layanan) HapusResep(ctx context.Context, noResep string) error {
	if strings.TrimSpace(noResep) == "" {
		return inputTidakValid("nomor resep wajib diisi")
	}
	return l.repositori.HapusResep(ctx, strings.TrimSpace(noResep))
}
