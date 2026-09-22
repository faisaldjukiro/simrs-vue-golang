package permintaan_laboratorium

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Nama tabel hanya berasal dari kategori yang diizinkan, bukan input SQL pengguna.
func kategoriValid(kategori string) (string, error) {
	kategori = strings.ToUpper(strings.TrimSpace(kategori))
	if kategori == "" {
		kategori = "PK"
	}
	switch kategori {
	case "PK", "PA", "MB":
		return kategori, nil
	default:
		return "", fmt.Errorf("%w: kategori laboratorium harus PK, PA, atau MB", ErrInputTidakValid)
	}
}

func (r *Repositori) untukKategori(kategori string) (*Repositori, error) {
	kategori, err := kategoriValid(kategori)
	if err != nil {
		return nil, err
	}
	salinan := *r
	salinan.kategori = kategori
	return &salinan, nil
}

func (r *Repositori) sqlKategori(query string) string {
	suffix := ""
	switch r.kategori {
	case "PA":
		suffix = "pa"
	case "MB":
		suffix = "mb"
	}
	return strings.NewReplacer(
		"permintaan_detail_permintaan_lab", "permintaan_detail_permintaan_lab"+suffix,
		"permintaan_pemeriksaan_lab", "permintaan_pemeriksaan_lab"+suffix,
		"permintaan_lab", "permintaan_lab"+suffix,
	).Replace(query)
}

// Field spesimen mengikuti permintaan_labpa pada Khanza.
type SpesimenPA struct {
	PengambilanBahan     string `json:"pengambilan_bahan"`
	DiperolehDengan      string `json:"diperoleh_dengan"`
	LokasiJaringan       string `json:"lokasi_jaringan"`
	DiawetkanDengan      string `json:"diawetkan_dengan"`
	PernahDilakukanDi    string `json:"pernah_dilakukan_di"`
	TanggalPASebelumnya  string `json:"tanggal_pa_sebelumnya"`
	NomorPASebelumnya    string `json:"nomor_pa_sebelumnya"`
	DiagnosaPASebelumnya string `json:"diagnosa_pa_sebelumnya"`
}

func (r *Repositori) bacaSpesimen(ctx context.Context, nomor string) (SpesimenPA, error) {
	var p SpesimenPA
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT COALESCE(DATE_FORMAT(pengambilan_bahan,'%Y-%m-%d'),''),
			COALESCE(diperoleh_dengan,''),COALESCE(lokasi_jaringan,''),COALESCE(diawetkan_dengan,''),
			COALESCE(pernah_dilakukan_di,''),
			IF(tanggal_pa_sebelumnya='0000-00-00','',COALESCE(DATE_FORMAT(tanggal_pa_sebelumnya,'%Y-%m-%d'),'')),
			COALESCE(nomor_pa_sebelumnya,''),COALESCE(diagnosa_pa_sebelumnya,'')
		FROM permintaan_labpa WHERE noorder=?`, nomor).Scan(
		&p.PengambilanBahan, &p.DiperolehDengan, &p.LokasiJaringan, &p.DiawetkanDengan,
		&p.PernahDilakukanDi, &p.TanggalPASebelumnya, &p.NomorPASebelumnya, &p.DiagnosaPASebelumnya,
	)
	return p, err
}

func (r *Repositori) simpanSpesimen(ctx context.Context, tx *sql.Tx, nomor string, p SpesimenPA) error {
	if r.kategori != "PA" {
		return nil
	}
	tanggal := p.TanggalPASebelumnya
	if p.PernahDilakukanDi == "" {
		tanggal = "0000-00-00"
	}
	_, err := tx.ExecContext(ctx, `
		UPDATE permintaan_labpa SET pengambilan_bahan=?,diperoleh_dengan=?,lokasi_jaringan=?,
			diawetkan_dengan=?,pernah_dilakukan_di=?,tanggal_pa_sebelumnya=?,
			nomor_pa_sebelumnya=?,diagnosa_pa_sebelumnya=? WHERE noorder=?`,
		p.PengambilanBahan, p.DiperolehDengan, p.LokasiJaringan, p.DiawetkanDengan,
		p.PernahDilakukanDi, tanggal, p.NomorPASebelumnya, p.DiagnosaPASebelumnya, nomor,
	)
	return err
}

func (r *Repositori) kunciPermintaan(ctx context.Context, tx *sql.Tx, noRawat, nomor string) error {
	var diterima int
	err := tx.QueryRowContext(ctx, r.sqlKategori(`
		SELECT CASE WHEN tgl_sampel<>'0000-00-00' OR jam_sampel<>'00:00:00'
			OR tgl_hasil<>'0000-00-00' OR jam_hasil<>'00:00:00' THEN 1 ELSE 0 END
		FROM permintaan_lab WHERE noorder=? AND no_rawat=? FOR UPDATE`), nomor, noRawat).Scan(&diterima)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTidakDitemukan
	}
	if err != nil {
		return err
	}
	if diterima > 0 {
		return ErrSudahDiterima
	}
	var dibayar int
	err = tx.QueryRowContext(ctx, r.sqlKategori(`
		SELECT COUNT(*) FROM permintaan_detail_permintaan_lab
		WHERE noorder=? AND stts_bayar='Sudah'`), nomor).Scan(&dibayar)
	if err != nil {
		return err
	}
	if dibayar > 0 {
		return ErrSudahDiproses
	}
	return nil
}
