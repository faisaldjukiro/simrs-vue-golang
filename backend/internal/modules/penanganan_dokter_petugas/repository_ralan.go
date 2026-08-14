package penanganan_dokter_petugas

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// DaftarRalan membaca tindakan gabungan dokter dan petugas pada tab
// Penanganan Dokter & Petugas di DlgRawatJalan.java.
func (r *Repositori) DaftarRalan(ctx context.Context, noRawat string) ([]Catatan, bool, error) {
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT r.no_rawat,r.kd_jenis_prw,COALESCE(j.nm_perawatan,''),'-',
			r.kd_dokter,COALESCE(d.nm_dokter,''),r.nip,COALESCE(p.nama,''),
			DATE_FORMAT(r.tgl_perawatan,'%Y-%m-%d'),TIME_FORMAT(r.jam_rawat,'%H:%i:%s'),
			COALESCE(r.material,0),COALESCE(r.bhp,0),COALESCE(r.tarif_tindakandr,0),
			COALESCE(r.tarif_tindakanpr,0),COALESCE(r.kso,0),COALESCE(r.menejemen,0),
			COALESCE(r.biaya_rawat,0),COALESCE(r.stts_bayar,'Belum')
		FROM rawat_jl_drpr r
		INNER JOIN jns_perawatan j ON j.kd_jenis_prw=r.kd_jenis_prw
		INNER JOIN dokter d ON d.kd_dokter=r.kd_dokter
		INNER JOIN petugas p ON p.nip=r.nip
		WHERE r.no_rawat=?
		ORDER BY r.tgl_perawatan DESC,r.jam_rawat DESC
	`, noRawat)
	if err != nil {
		return nil, false, fmt.Errorf("baca penanganan dokter dan petugas IGD/rawat jalan: %w", err)
	}
	defer rows.Close()

	daftar := make([]Catatan, 0)
	for rows.Next() {
		item := Catatan{JenisRawat: "ralan"}
		if err := rows.Scan(&item.NoRawat, &item.KodeTindakan, &item.NamaTindakan, &item.Kelas,
			&item.KodeDokter, &item.NamaDokter, &item.KodePetugas, &item.NamaPetugas,
			&item.Tanggal, &item.Jam, &item.Material, &item.BHP, &item.TarifDokter,
			&item.TarifPetugas, &item.KSO, &item.Manajemen, &item.Total, &item.StatusBayar); err != nil {
			return nil, false, fmt.Errorf("scan penanganan dokter dan petugas IGD/rawat jalan: %w", err)
		}
		daftar = append(daftar, item)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterasi penanganan dokter dan petugas IGD/rawat jalan: %w", err)
	}
	terkunci, err := r.billingTerkunci(ctx, r.simrsDB, noRawat)
	return daftar, terkunci, err
}

func (r *Repositori) DokterRalan(ctx context.Context, noRawat string) (Dokter, error) {
	var item Dokter
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT d.kd_dokter,COALESCE(d.nm_dokter,''),COALESCE(s.nm_sps,'')
		FROM reg_periksa rp
		INNER JOIN dokter d ON d.kd_dokter=rp.kd_dokter
		LEFT JOIN spesialis s ON s.kd_sps=d.kd_sps
		WHERE rp.no_rawat=? LIMIT 1
	`, noRawat).Scan(&item.Kode, &item.Nama, &item.Spesialis)
	return item, err
}

func (r *Repositori) CariTindakanRalan(ctx context.Context, noRawat, kata string) ([]Tindakan, error) {
	seperti := "%" + strings.TrimSpace(kata) + "%"
	aturan, err := bacaAturanTarifRalan(ctx, r.simrsDB)
	if err != nil {
		return nil, fmt.Errorf("baca aturan tarif IGD/rawat jalan: %w", err)
	}
	query := `
		SELECT j.kd_jenis_prw,COALESCE(j.nm_perawatan,''),COALESCE(k.nm_kategori,''),'-',
			COALESCE(j.material,0),COALESCE(j.bhp,0),COALESCE(j.tarif_tindakandr,0),
			COALESCE(j.tarif_tindakanpr,0),COALESCE(j.kso,0),COALESCE(j.menejemen,0),
			COALESCE(j.total_byrdrpr,0)
		FROM reg_periksa rp
		INNER JOIN jns_perawatan j ON j.status='1'` + kondisiTarifRalan(aturan) + `
		LEFT JOIN kategori_perawatan k ON k.kd_kategori=j.kd_kategori
		WHERE rp.no_rawat=? AND j.total_byrdrpr>0
			AND (j.kd_jenis_prw LIKE ? OR j.nm_perawatan LIKE ? OR k.nm_kategori LIKE ?)
		ORDER BY j.nm_perawatan
		LIMIT 40
	`
	rows, err := r.simrsDB.QueryContext(ctx, query, noRawat, seperti, seperti, seperti)
	if err != nil {
		return nil, fmt.Errorf("cari tindakan IGD/rawat jalan: %w", err)
	}
	defer rows.Close()
	daftar := make([]Tindakan, 0)
	for rows.Next() {
		var item Tindakan
		if err := rows.Scan(&item.Kode, &item.Nama, &item.Kategori, &item.Kelas, &item.Material,
			&item.BHP, &item.TarifDokter, &item.TarifPetugas, &item.KSO, &item.Manajemen, &item.Total); err != nil {
			return nil, err
		}
		daftar = append(daftar, item)
	}
	return daftar, rows.Err()
}

func (r *Repositori) SimpanBanyakRalan(ctx context.Context, daftar []Catatan) error {
	return r.dalamTransaksi(ctx, daftar[0].NoRawat, func(tx *sql.Tx) error {
		for _, item := range daftar {
			tarif, err := r.tarifRalan(ctx, tx, item.NoRawat, item.KodeTindakan)
			if err != nil {
				return err
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO rawat_jl_drpr
				(no_rawat,kd_jenis_prw,kd_dokter,nip,tgl_perawatan,jam_rawat,material,bhp,tarif_tindakandr,tarif_tindakanpr,kso,menejemen,biaya_rawat,stts_bayar)
				VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,'Belum')`, item.NoRawat, tarif.Kode, item.KodeDokter,
				item.KodePetugas, item.Tanggal, item.Jam, tarif.Material, tarif.BHP, tarif.TarifDokter,
				tarif.TarifPetugas, tarif.KSO, tarif.Manajemen, tarif.Total)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repositori) UbahRalan(ctx context.Context, lama Kunci, item Catatan) error {
	return r.dalamTransaksi(ctx, lama.NoRawat, func(tx *sql.Tx) error {
		tarif, err := r.tarifRalan(ctx, tx, item.NoRawat, item.KodeTindakan)
		if err != nil {
			return err
		}
		hasil, err := tx.ExecContext(ctx, `UPDATE rawat_jl_drpr SET no_rawat=?,kd_jenis_prw=?,kd_dokter=?,nip=?,tgl_perawatan=?,jam_rawat=?,material=?,bhp=?,tarif_tindakandr=?,tarif_tindakanpr=?,kso=?,menejemen=?,biaya_rawat=?
			WHERE no_rawat=? AND kd_jenis_prw=? AND kd_dokter=? AND nip=? AND tgl_perawatan=? AND jam_rawat=?`,
			item.NoRawat, tarif.Kode, item.KodeDokter, item.KodePetugas, item.Tanggal, item.Jam,
			tarif.Material, tarif.BHP, tarif.TarifDokter, tarif.TarifPetugas, tarif.KSO, tarif.Manajemen,
			tarif.Total, lama.NoRawat, lama.KodeTindakan, lama.KodeDokter, lama.KodePetugas, lama.Tanggal, lama.Jam)
		if err != nil {
			return err
		}
		jumlah, err := hasil.RowsAffected()
		if err == nil && jumlah == 0 {
			return ErrTidakDitemukan
		}
		return err
	})
}

func (r *Repositori) HapusRalan(ctx context.Context, kunci Kunci) error {
	return r.dalamTransaksi(ctx, kunci.NoRawat, func(tx *sql.Tx) error {
		hasil, err := tx.ExecContext(ctx, `DELETE FROM rawat_jl_drpr WHERE no_rawat=? AND kd_jenis_prw=? AND kd_dokter=? AND nip=? AND tgl_perawatan=? AND jam_rawat=?`,
			kunci.NoRawat, kunci.KodeTindakan, kunci.KodeDokter, kunci.KodePetugas, kunci.Tanggal, kunci.Jam)
		if err != nil {
			return err
		}
		jumlah, err := hasil.RowsAffected()
		if err == nil && jumlah == 0 {
			return ErrTidakDitemukan
		}
		return err
	})
}

func (r *Repositori) tarifRalan(ctx context.Context, q queryer, noRawat, kode string) (Tindakan, error) {
	var item Tindakan
	aturan, err := bacaAturanTarifRalan(ctx, q)
	if err != nil {
		return item, fmt.Errorf("baca aturan tarif IGD/rawat jalan: %w", err)
	}
	query := `
		SELECT j.kd_jenis_prw,COALESCE(j.nm_perawatan,''),COALESCE(k.nm_kategori,''),'-',
			COALESCE(j.material,0),COALESCE(j.bhp,0),COALESCE(j.tarif_tindakandr,0),
			COALESCE(j.tarif_tindakanpr,0),COALESCE(j.kso,0),COALESCE(j.menejemen,0),
			COALESCE(j.total_byrdrpr,0)
		FROM reg_periksa rp
		INNER JOIN jns_perawatan j ON j.kd_jenis_prw=? AND j.status='1'` + kondisiTarifRalan(aturan) + `
		LEFT JOIN kategori_perawatan k ON k.kd_kategori=j.kd_kategori
		WHERE rp.no_rawat=? AND j.total_byrdrpr>0
		LIMIT 1
	`
	err = q.QueryRowContext(ctx, query, kode, noRawat).Scan(&item.Kode, &item.Nama, &item.Kategori, &item.Kelas, &item.Material,
		&item.BHP, &item.TarifDokter, &item.TarifPetugas, &item.KSO, &item.Manajemen, &item.Total)
	if errors.Is(err, sql.ErrNoRows) {
		return item, fmt.Errorf("tindakan tidak tersedia untuk cara bayar dan poliklinik pasien")
	}
	return item, err
}

type aturanTarifRalan struct {
	Poliklinik bool
	CaraBayar  bool
}

func bacaAturanTarifRalan(ctx context.Context, q queryer) (aturanTarifRalan, error) {
	aturan := aturanTarifRalan{Poliklinik: true, CaraBayar: true}
	var poliklinik, caraBayar string
	err := q.QueryRowContext(ctx, `SELECT poli_ralan,cara_bayar_ralan FROM set_tarif LIMIT 1`).Scan(&poliklinik, &caraBayar)
	if errors.Is(err, sql.ErrNoRows) {
		return aturan, nil
	}
	if err != nil {
		return aturan, err
	}
	aturan.Poliklinik = strings.EqualFold(strings.TrimSpace(poliklinik), "Yes")
	aturan.CaraBayar = strings.EqualFold(strings.TrimSpace(caraBayar), "Yes")
	return aturan, nil
}

func kondisiTarifRalan(aturan aturanTarifRalan) string {
	kondisi := ""
	if aturan.CaraBayar {
		kondisi += " AND (j.kd_pj=rp.kd_pj OR j.kd_pj='-')"
	}
	if aturan.Poliklinik {
		kondisi += " AND (j.kd_poli=rp.kd_poli OR j.kd_poli='-')"
	}
	return kondisi
}
