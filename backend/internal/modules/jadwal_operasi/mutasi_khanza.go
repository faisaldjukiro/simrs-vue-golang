package jadwal_operasi

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"simrs-backend/internal/shared/khanzamutasi"
)

func snapshot(in Input) ([]string, []any) {
	return []string{"no_rawat", "kode_paket", "tanggal", "jam_mulai", "jam_selesai", "status", "kd_dokter", "kd_ruang_ok", "dokteranastesi", "perawat"},
		[]any{in.NoRawat, in.KodePaket, in.Tanggal, in.JamMulai, in.JamSelesai, in.Status, in.KdDokter, in.KdRuangOK, in.DokterAnestesi, in.Perawat}
}

func (r *Repositori) mutasiKhanza(ctx context.Context, in Input, hapus bool) error {
	if in.Asli == nil || in.NoRawat == "" || in.Asli.NoRawat != in.NoRawat || in.Asli.KodePaket == "" || in.Asli.Tanggal == "" || in.Asli.JamMulai == "" || in.Asli.JamSelesai == "" || in.Asli.KdDokter == "" || in.Asli.KdRuangOK == "" || in.Asli.Status == "" {
		return ErrValidasi
	}
	kolom, lama := snapshot(*in.Asli)
	_, baru := snapshot(in)
	err := khanzamutasi.Jalankan(ctx, r.simrs, "booking_operasi", kolom, lama, baru, hapus)
	if errors.Is(err, khanzamutasi.ErrKonflik) {
		return ErrKonflik
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrBentrok
	}
	return err
}
