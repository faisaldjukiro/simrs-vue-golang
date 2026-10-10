package edukasi_pasien

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/png"
	"strings"
)

// Hanya mendeteksi kolom; perubahan struktur SIMRS dilakukan pengguna di luar aplikasi.
func (r *Repositori) kolomBuktiTersedia(ctx context.Context) (map[string]bool, error) {
	hasil := map[string]bool{}
	rows, err := r.simrs.QueryContext(ctx, `SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=? AND COLUMN_NAME IN ('paraf_petugas','nama_penerima','paraf_penerima','foto_penerima')`, tabel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		hasil[k] = true
	}
	return hasil, rows.Err()
}

func kolomBuktiOpsional(k string) bool {
	return k == "paraf_petugas" || k == "nama_penerima" || k == "paraf_penerima" || k == "foto_penerima"
}

func kolomGambarBukti(k string) bool {
	return k == "paraf_petugas" || k == "paraf_penerima" || k == "foto" || k == "foto_penerima"
}

func validasiParaf(nilai string) error {
	if nilai == "" {
		return nil
	}
	gagal := fmt.Errorf("%w: paraf harus digambar ulang pada kotak paraf", ErrValidasi)
	const prefix = "data:image/png;base64,"
	if len(nilai) > 256*1024 || !strings.HasPrefix(nilai, prefix) {
		return gagal
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(nilai, prefix))
	if err != nil {
		return gagal
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width != 720 || cfg.Height != 240 {
		return gagal
	}
	gambar, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return gagal
	}
	// Tolak kanvas kosong (putih atau transparan).
	tinta := 0
	for y := 0; y < cfg.Height; y++ {
		for x := 0; x < cfg.Width; x++ {
			r, g, b, a := gambar.At(x, y).RGBA()
			if a > 32767 && r < 49151 && g < 49151 && b < 49151 {
				tinta++
			}
		}
	}
	if tinta < 20 {
		return gagal
	}
	return nil
}
