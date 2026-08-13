package riwayat_perawatan

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInputTidakValid = errors.New("filter riwayat perawatan tidak valid")

type Layanan struct{ repo *Repositori }

func NewLayanan(repo *Repositori) *Layanan { return &Layanan{repo: repo} }

func (l *Layanan) Data(ctx context.Context, filter Filter) (Data, error) {
	filter.NoRawat = strings.TrimSpace(filter.NoRawat)
	filter.Mode = strings.ToLower(strings.TrimSpace(filter.Mode))
	filter.TanggalMulai = strings.TrimSpace(filter.TanggalMulai)
	filter.TanggalSelesai = strings.TrimSpace(filter.TanggalSelesai)
	filter.NomorRawat = strings.TrimSpace(filter.NomorRawat)
	if filter.NoRawat == "" {
		return Data{}, fmt.Errorf("%w: nomor rawat pasien aktif wajib diisi", ErrInputTidakValid)
	}
	if filter.Mode == "" {
		filter.Mode = "terakhir"
	}
	if filter.Mode != "terakhir" && filter.Mode != "semua" && filter.Mode != "tanggal" && filter.Mode != "nomor" {
		return Data{}, fmt.Errorf("%w: mode pencarian tidak dikenal", ErrInputTidakValid)
	}
	if filter.Mode == "tanggal" {
		mulai, err1 := time.Parse("2006-01-02", filter.TanggalMulai)
		selesai, err2 := time.Parse("2006-01-02", filter.TanggalSelesai)
		if err1 != nil || err2 != nil || mulai.After(selesai) {
			return Data{}, fmt.Errorf("%w: rentang tanggal tidak valid", ErrInputTidakValid)
		}
	}
	if filter.Mode == "nomor" && filter.NomorRawat == "" {
		return Data{}, fmt.Errorf("%w: nomor rawat pencarian wajib diisi", ErrInputTidakValid)
	}
	return l.repo.Data(ctx, filter)
}
