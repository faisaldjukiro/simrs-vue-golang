package perencanaan_pemulangan

import (
	"context"
	"fmt"
	"strings"
)

type KopCetak struct {
	Nama      string `json:"nama"`
	Alamat    string `json:"alamat"`
	Kabupaten string `json:"kabupaten"`
	Propinsi  string `json:"propinsi"`
	Kontak    string `json:"kontak"`
	Email     string `json:"email"`
}

// Hanya baca identitas instansi. Tidak mengambil blob wallpaper atau mengubah setting.
func (r *Repositori) Kop(ctx context.Context) (KopCetak, error) {
	var k KopCetak
	err := r.simrs.QueryRowContext(ctx, `SELECT nama_instansi,
 COALESCE(alamat_instansi,''),COALESCE(kabupaten,''),COALESCE(propinsi,''),
 COALESCE(kontak,''),COALESCE(email,'')
 FROM setting ORDER BY (aktifkan='Yes') DESC,nama_instansi LIMIT 1`).Scan(
		&k.Nama, &k.Alamat, &k.Kabupaten, &k.Propinsi, &k.Kontak, &k.Email)
	if err != nil {
		return k, err
	}
	if strings.TrimSpace(k.Nama) == "" {
		return k, fmt.Errorf("%w: nama instansi pada setting belum diisi", ErrValidasi)
	}
	return k, nil
}
