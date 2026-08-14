package penanganan_dokter_petugas

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrTidakDitemukan  = errors.New("data penanganan dokter dan petugas tidak ditemukan")
	ErrBillingTerkunci = errors.New("billing sudah terverifikasi atau kunjungan dibatalkan")
)

type Dokter struct {
	Kode      string `json:"kode"`
	Nama      string `json:"nama"`
	Spesialis string `json:"spesialis"`
}

type Petugas struct {
	Kode    string `json:"kode"`
	Nama    string `json:"nama"`
	Jabatan string `json:"jabatan"`
}

type Tindakan struct {
	Kode         string  `json:"kode"`
	Nama         string  `json:"nama"`
	Kategori     string  `json:"kategori"`
	Kelas        string  `json:"kelas"`
	Material     float64 `json:"material"`
	BHP          float64 `json:"bhp"`
	TarifDokter  float64 `json:"tarif_dokter"`
	TarifPetugas float64 `json:"tarif_petugas"`
	KSO          float64 `json:"kso"`
	Manajemen    float64 `json:"manajemen"`
	Total        float64 `json:"total"`
}

type Catatan struct {
	JenisRawat   string  `json:"jenis_rawat"`
	NoRawat      string  `json:"no_rawat"`
	KodeTindakan string  `json:"kode_tindakan"`
	NamaTindakan string  `json:"nama_tindakan"`
	Kelas        string  `json:"kelas"`
	KodeDokter   string  `json:"kode_dokter"`
	NamaDokter   string  `json:"nama_dokter"`
	KodePetugas  string  `json:"kode_petugas"`
	NamaPetugas  string  `json:"nama_petugas"`
	Tanggal      string  `json:"tanggal"`
	Jam          string  `json:"jam"`
	Material     float64 `json:"material"`
	BHP          float64 `json:"bhp"`
	TarifDokter  float64 `json:"tarif_dokter"`
	TarifPetugas float64 `json:"tarif_petugas"`
	KSO          float64 `json:"kso"`
	Manajemen    float64 `json:"manajemen"`
	Total        float64 `json:"total"`
	StatusBayar  string  `json:"status_bayar,omitempty"`
}

type Kunci struct {
	JenisRawat   string `json:"jenis_rawat"`
	NoRawat      string `json:"no_rawat"`
	KodeTindakan string `json:"kode_tindakan"`
	KodeDokter   string `json:"kode_dokter"`
	KodePetugas  string `json:"kode_petugas"`
	Tanggal      string `json:"tanggal"`
	Jam          string `json:"jam"`
}

type Repositori struct{ simrsDB *sql.DB }

func NewRepositori(simrsDB *sql.DB) *Repositori { return &Repositori{simrsDB: simrsDB} }

func (r *Repositori) Daftar(ctx context.Context, noRawat string) ([]Catatan, bool, error) {
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT r.no_rawat, r.kd_jenis_prw, COALESCE(j.nm_perawatan, ''), j.kelas,
			r.kd_dokter, COALESCE(d.nm_dokter, ''), r.nip, COALESCE(p.nama, ''),
			DATE_FORMAT(r.tgl_perawatan, '%Y-%m-%d'), TIME_FORMAT(r.jam_rawat, '%H:%i:%s'),
			COALESCE(r.material, 0), COALESCE(r.bhp, 0), COALESCE(r.tarif_tindakandr, 0),
			COALESCE(r.tarif_tindakanpr, 0), COALESCE(r.kso, 0), COALESCE(r.menejemen, 0),
			COALESCE(r.biaya_rawat, 0)
		FROM rawat_inap_drpr r
		INNER JOIN jns_perawatan_inap j ON j.kd_jenis_prw = r.kd_jenis_prw
		INNER JOIN dokter d ON d.kd_dokter = r.kd_dokter
		INNER JOIN petugas p ON p.nip = r.nip
		WHERE r.no_rawat = ?
		ORDER BY r.tgl_perawatan DESC, r.jam_rawat DESC
	`, noRawat)
	if err != nil {
		return nil, false, fmt.Errorf("baca penanganan dokter dan petugas: %w", err)
	}
	defer rows.Close()

	daftar := make([]Catatan, 0)
	for rows.Next() {
		item := Catatan{JenisRawat: "ranap"}
		if err := rows.Scan(&item.NoRawat, &item.KodeTindakan, &item.NamaTindakan, &item.Kelas,
			&item.KodeDokter, &item.NamaDokter, &item.KodePetugas, &item.NamaPetugas,
			&item.Tanggal, &item.Jam, &item.Material, &item.BHP, &item.TarifDokter,
			&item.TarifPetugas, &item.KSO, &item.Manajemen, &item.Total); err != nil {
			return nil, false, fmt.Errorf("scan penanganan dokter dan petugas: %w", err)
		}
		daftar = append(daftar, item)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterasi penanganan dokter dan petugas: %w", err)
	}
	terkunci, err := r.billingTerkunci(ctx, r.simrsDB, noRawat)
	return daftar, terkunci, err
}

func (r *Repositori) DokterDPJP(ctx context.Context, noRawat string) (Dokter, error) {
	var item Dokter
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT d.kd_dokter, COALESCE(d.nm_dokter, ''), COALESCE(s.nm_sps, '')
		FROM dpjp_ranap dp
		INNER JOIN dokter d ON d.kd_dokter = dp.kd_dokter
		LEFT JOIN spesialis s ON s.kd_sps = d.kd_sps
		WHERE dp.no_rawat = ? ORDER BY d.nm_dokter LIMIT 1
	`, noRawat).Scan(&item.Kode, &item.Nama, &item.Spesialis)
	return item, err
}

func (r *Repositori) PetugasLogin(ctx context.Context, kode string) (Petugas, error) {
	var item Petugas
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT p.nip, COALESCE(p.nama, ''), COALESCE(j.nm_jbtn, '')
		FROM petugas p LEFT JOIN jabatan j ON j.kd_jbtn = p.kd_jbtn
		WHERE p.nip = ? LIMIT 1
	`, strings.TrimSpace(kode)).Scan(&item.Kode, &item.Nama, &item.Jabatan)
	return item, err
}

func (r *Repositori) CariDokter(ctx context.Context, kata string) ([]Dokter, error) {
	seperti := "%" + strings.TrimSpace(kata) + "%"
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT d.kd_dokter, COALESCE(d.nm_dokter, ''), COALESCE(s.nm_sps, '')
		FROM dokter d LEFT JOIN spesialis s ON s.kd_sps = d.kd_sps
		WHERE d.status = '1' AND (d.kd_dokter LIKE ? OR d.nm_dokter LIKE ? OR s.nm_sps LIKE ?)
		ORDER BY d.nm_dokter LIMIT 30
	`, seperti, seperti, seperti)
	if err != nil {
		return nil, fmt.Errorf("cari dokter: %w", err)
	}
	defer rows.Close()
	daftar := make([]Dokter, 0)
	for rows.Next() {
		var item Dokter
		if err := rows.Scan(&item.Kode, &item.Nama, &item.Spesialis); err != nil {
			return nil, err
		}
		daftar = append(daftar, item)
	}
	return daftar, rows.Err()
}

func (r *Repositori) CariPetugas(ctx context.Context, kata string) ([]Petugas, error) {
	seperti := "%" + strings.TrimSpace(kata) + "%"
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT p.nip, COALESCE(p.nama, ''), COALESCE(j.nm_jbtn, '')
		FROM petugas p LEFT JOIN jabatan j ON j.kd_jbtn = p.kd_jbtn
		WHERE p.status = '1' AND (p.nip LIKE ? OR p.nama LIKE ? OR j.nm_jbtn LIKE ?)
		ORDER BY p.nama LIMIT 30
	`, seperti, seperti, seperti)
	if err != nil {
		return nil, fmt.Errorf("cari petugas: %w", err)
	}
	defer rows.Close()
	daftar := make([]Petugas, 0)
	for rows.Next() {
		var item Petugas
		if err := rows.Scan(&item.Kode, &item.Nama, &item.Jabatan); err != nil {
			return nil, err
		}
		daftar = append(daftar, item)
	}
	return daftar, rows.Err()
}

func (r *Repositori) CariTindakan(ctx context.Context, noRawat, kata string) ([]Tindakan, error) {
	kodeCaraBayar, err := r.kodeCaraBayarTarif(ctx, noRawat)
	if err != nil {
		return nil, err
	}
	seperti := "%" + strings.TrimSpace(kata) + "%"
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT j.kd_jenis_prw, COALESCE(j.nm_perawatan, ''), COALESCE(k.nm_kategori, ''), j.kelas,
			COALESCE(j.material, 0), COALESCE(j.bhp, 0), COALESCE(j.tarif_tindakandr, 0),
			COALESCE(j.tarif_tindakanpr, 0), COALESCE(j.kso, 0), COALESCE(j.menejemen, 0),
			COALESCE(j.total_byrdrpr, 0)
		FROM jns_perawatan_inap j
		LEFT JOIN kategori_perawatan k ON k.kd_kategori = j.kd_kategori
		WHERE j.status = '1'
			AND j.kd_pj = ?
			AND (j.kd_jenis_prw LIKE ? OR j.nm_perawatan LIKE ? OR k.nm_kategori LIKE ?)
		ORDER BY j.nm_perawatan LIMIT 40
	`, kodeCaraBayar, seperti, seperti, seperti)
	if err != nil {
		return nil, fmt.Errorf("cari tindakan rawat inap: %w", err)
	}
	defer rows.Close()
	daftar := make([]Tindakan, 0)
	for rows.Next() {
		var item Tindakan
		if err := rows.Scan(&item.Kode, &item.Nama, &item.Kategori, &item.Kelas, &item.Material, &item.BHP, &item.TarifDokter, &item.TarifPetugas, &item.KSO, &item.Manajemen, &item.Total); err != nil {
			return nil, err
		}
		daftar = append(daftar, item)
	}
	return daftar, rows.Err()
}

func (r *Repositori) Simpan(ctx context.Context, item Catatan) error {
	return r.SimpanBanyak(ctx, []Catatan{item})
}

func (r *Repositori) SimpanBanyak(ctx context.Context, daftar []Catatan) error {
	return r.dalamTransaksi(ctx, daftar[0].NoRawat, func(tx *sql.Tx) error {
		for _, item := range daftar {
			tarif, err := r.tarif(ctx, tx, item.NoRawat, item.KodeTindakan)
			if err != nil {
				return err
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO rawat_inap_drpr
				(no_rawat,kd_jenis_prw,kd_dokter,nip,tgl_perawatan,jam_rawat,material,bhp,tarif_tindakandr,tarif_tindakanpr,kso,menejemen,biaya_rawat)
				VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, item.NoRawat, tarif.Kode, item.KodeDokter, item.KodePetugas,
				item.Tanggal, item.Jam, tarif.Material, tarif.BHP, tarif.TarifDokter, tarif.TarifPetugas, tarif.KSO, tarif.Manajemen, tarif.Total)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repositori) Ubah(ctx context.Context, lama Kunci, item Catatan) error {
	return r.dalamTransaksi(ctx, lama.NoRawat, func(tx *sql.Tx) error {
		tarif, err := r.tarif(ctx, tx, item.NoRawat, item.KodeTindakan)
		if err != nil {
			return err
		}
		hasil, err := tx.ExecContext(ctx, `UPDATE rawat_inap_drpr SET no_rawat=?,kd_jenis_prw=?,kd_dokter=?,nip=?,tgl_perawatan=?,jam_rawat=?,material=?,bhp=?,tarif_tindakandr=?,tarif_tindakanpr=?,kso=?,menejemen=?,biaya_rawat=?
			WHERE no_rawat=? AND kd_jenis_prw=? AND kd_dokter=? AND nip=? AND tgl_perawatan=? AND jam_rawat=?`,
			item.NoRawat, tarif.Kode, item.KodeDokter, item.KodePetugas, item.Tanggal, item.Jam,
			tarif.Material, tarif.BHP, tarif.TarifDokter, tarif.TarifPetugas, tarif.KSO, tarif.Manajemen, tarif.Total,
			lama.NoRawat, lama.KodeTindakan, lama.KodeDokter, lama.KodePetugas, lama.Tanggal, lama.Jam)
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

func (r *Repositori) Hapus(ctx context.Context, kunci Kunci) error {
	return r.dalamTransaksi(ctx, kunci.NoRawat, func(tx *sql.Tx) error {
		hasil, err := tx.ExecContext(ctx, `DELETE FROM rawat_inap_drpr WHERE no_rawat=? AND kd_jenis_prw=? AND kd_dokter=? AND nip=? AND tgl_perawatan=? AND jam_rawat=?`, kunci.NoRawat, kunci.KodeTindakan, kunci.KodeDokter, kunci.KodePetugas, kunci.Tanggal, kunci.Jam)
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

func (r *Repositori) dalamTransaksi(ctx context.Context, noRawat string, proses func(*sql.Tx) error) error {
	tx, err := r.simrsDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	terkunci, err := r.billingTerkunci(ctx, tx, noRawat)
	if err != nil {
		return err
	}
	if terkunci {
		return ErrBillingTerkunci
	}
	if err := proses(tx); err != nil {
		return err
	}
	return tx.Commit()
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (r *Repositori) billingTerkunci(ctx context.Context, q queryer, noRawat string) (bool, error) {
	var jumlah int
	err := q.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM billing WHERE no_rawat=?) + (SELECT COUNT(*) FROM reg_periksa WHERE no_rawat=? AND stts='Batal')`, noRawat, noRawat).Scan(&jumlah)
	return jumlah > 0, err
}

func (r *Repositori) tarif(ctx context.Context, q queryer, noRawat, kode string) (Tindakan, error) {
	var item Tindakan
	err := q.QueryRowContext(ctx, `SELECT j.kd_jenis_prw,COALESCE(j.nm_perawatan,''),COALESCE(k.nm_kategori,''),j.kelas,COALESCE(j.material,0),COALESCE(j.bhp,0),COALESCE(j.tarif_tindakandr,0),COALESCE(j.tarif_tindakanpr,0),COALESCE(j.kso,0),COALESCE(j.menejemen,0),COALESCE(j.total_byrdrpr,0)
		FROM jns_perawatan_inap j
		LEFT JOIN kategori_perawatan k ON k.kd_kategori=j.kd_kategori
		WHERE j.kd_jenis_prw=?
			AND j.status='1'
			AND j.kd_pj=(
				SELECT CASE
					WHEN UPPER(TRIM(rp.kd_pj)) IN ('BPJ','36') THEN 'BPJ'
					ELSE 'A09'
				END
				FROM reg_periksa rp
				WHERE rp.no_rawat=?
				LIMIT 1
			)
		LIMIT 1`, kode, noRawat).Scan(&item.Kode, &item.Nama, &item.Kategori, &item.Kelas, &item.Material, &item.BHP, &item.TarifDokter, &item.TarifPetugas, &item.KSO, &item.Manajemen, &item.Total)
	if errors.Is(err, sql.ErrNoRows) {
		return item, fmt.Errorf("tindakan tidak tersedia untuk cara bayar pasien")
	}
	return item, err
}

func (r *Repositori) kodeCaraBayarTarif(ctx context.Context, noRawat string) (string, error) {
	var kode string
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT CASE
			WHEN UPPER(TRIM(kd_pj)) IN ('BPJ','36') THEN 'BPJ'
			ELSE 'A09'
		END
		FROM reg_periksa
		WHERE no_rawat=?
		LIMIT 1
	`, noRawat).Scan(&kode)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("data registrasi pasien tidak ditemukan")
	}
	return kode, err
}
