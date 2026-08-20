package permintaan_radiologi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrTidakDitemukan  = errors.New("permintaan radiologi tidak ditemukan")
	ErrBillingTerkunci = errors.New("billing sudah terverifikasi atau kunjungan dibatalkan")
	ErrSudahDiproses   = errors.New("permintaan radiologi sudah diproses")
	ErrSudahDiterima   = errors.New("permintaan radiologi sudah diterima petugas radiologi")
)

type Dokter struct {
	Kode      string `json:"kode"`
	Nama      string `json:"nama"`
	Spesialis string `json:"spesialis"`
}

type Tindakan struct {
	Kode          string  `json:"kode"`
	Nama          string  `json:"nama"`
	KodeCaraBayar string  `json:"kode_cara_bayar"`
	Kelas         string  `json:"kelas"`
	Total         float64 `json:"total"`
}

type Pemeriksaan struct {
	Tindakan
	StatusBayar string `json:"status_bayar"`
}

type Permintaan struct {
	Nomor             string        `json:"nomor"`
	NoRawat           string        `json:"no_rawat"`
	Tanggal           string        `json:"tanggal"`
	Jam               string        `json:"jam"`
	TanggalDiterima   string        `json:"tanggal_diterima"`
	JamDiterima       string        `json:"jam_diterima"`
	TanggalHasil      string        `json:"tanggal_hasil"`
	JamHasil          string        `json:"jam_hasil"`
	KodeDokter        string        `json:"kode_dokter"`
	NamaDokter        string        `json:"nama_dokter"`
	StatusRawat       string        `json:"status_rawat"`
	StatusPemeriksaan string        `json:"status_pemeriksaan"`
	StatusBayar       string        `json:"status_bayar"`
	InformasiTambahan string        `json:"informasi_tambahan"`
	DiagnosisKlinis   string        `json:"diagnosis_klinis"`
	Pemeriksaan       []Pemeriksaan `json:"pemeriksaan"`
	Total             float64       `json:"total"`
	DapatDiubah       bool          `json:"dapat_diubah"`
	DapatDihapus      bool          `json:"dapat_dihapus"`
}

type Data struct {
	Permintaan      []Permintaan `json:"permintaan"`
	DokterPerujuk   *Dokter      `json:"dokter_perujuk,omitempty"`
	BillingTerkunci bool         `json:"billing_terkunci"`
	StatusRawat     string       `json:"status_rawat"`
	KodeCaraBayar   string       `json:"kode_cara_bayar"`
	KelasPasien     string       `json:"kelas_pasien"`
	FilterCaraBayar bool         `json:"filter_cara_bayar"`
	FilterKelas     bool         `json:"filter_kelas"`
}

type Input struct {
	NoRawat           string   `json:"no_rawat"`
	Tanggal           string   `json:"tanggal"`
	Jam               string   `json:"jam"`
	KodeDokter        string   `json:"kode_dokter"`
	InformasiTambahan string   `json:"informasi_tambahan"`
	DiagnosisKlinis   string   `json:"diagnosis_klinis"`
	KodeTindakan      []string `json:"kode_tindakan"`
}

type lingkupTarif struct {
	KodeCaraBayar   string
	Kelas           string
	StatusRawat     string
	FilterCaraBayar bool
	FilterKelas     bool
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Repositori struct{ simrsDB *sql.DB }

func NewRepositori(simrsDB *sql.DB) *Repositori { return &Repositori{simrsDB: simrsDB} }

func (r *Repositori) Data(ctx context.Context, noRawat string) (Data, error) {
	lingkup, err := r.lingkup(ctx, r.simrsDB, noRawat)
	if err != nil {
		return Data{}, err
	}
	permintaan, err := r.daftar(ctx, noRawat)
	if err != nil {
		return Data{}, err
	}
	terkunci, err := r.billingTerkunci(ctx, r.simrsDB, noRawat)
	if err != nil {
		return Data{}, err
	}
	data := Data{
		Permintaan:      permintaan,
		BillingTerkunci: terkunci,
		StatusRawat:     lingkup.StatusRawat,
		KodeCaraBayar:   lingkup.KodeCaraBayar,
		KelasPasien:     lingkup.Kelas,
		FilterCaraBayar: lingkup.FilterCaraBayar,
		FilterKelas:     lingkup.FilterKelas,
	}
	if dokter, err := r.dokterPerujuk(ctx, noRawat); err == nil {
		data.DokterPerujuk = &dokter
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Data{}, err
	}
	return data, nil
}

func (r *Repositori) daftar(ctx context.Context, noRawat string) ([]Permintaan, error) {
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT pr.noorder, pr.no_rawat, DATE_FORMAT(pr.tgl_permintaan,'%Y-%m-%d'),
			TIME_FORMAT(pr.jam_permintaan,'%H:%i:%s'),
			IF(pr.tgl_sampel='0000-00-00','',DATE_FORMAT(pr.tgl_sampel,'%Y-%m-%d')),
			IF(pr.jam_sampel='00:00:00','',TIME_FORMAT(pr.jam_sampel,'%H:%i:%s')),
			IF(pr.tgl_hasil='0000-00-00','',DATE_FORMAT(pr.tgl_hasil,'%Y-%m-%d')),
			IF(pr.jam_hasil='00:00:00','',TIME_FORMAT(pr.jam_hasil,'%H:%i:%s')),
			pr.dokter_perujuk,
			COALESCE(d.nm_dokter,''), pr.status, pr.informasi_tambahan, pr.diagnosa_klinis,
			ppr.kd_jenis_prw, COALESCE(j.nm_perawatan,''), j.kd_pj, j.kelas,
			COALESCE(j.total_byr,0), COALESCE(ppr.stts_bayar,'Belum')
		FROM permintaan_radiologi pr
		INNER JOIN dokter d ON d.kd_dokter=pr.dokter_perujuk
		INNER JOIN permintaan_pemeriksaan_radiologi ppr ON ppr.noorder=pr.noorder
		INNER JOIN jns_perawatan_radiologi j ON j.kd_jenis_prw=ppr.kd_jenis_prw
		WHERE pr.no_rawat=?
		ORDER BY pr.tgl_permintaan DESC, pr.jam_permintaan DESC, j.nm_perawatan
	`, noRawat)
	if err != nil {
		return nil, fmt.Errorf("baca permintaan radiologi: %w", err)
	}
	defer rows.Close()

	daftar := make([]Permintaan, 0)
	indeks := make(map[string]int)
	for rows.Next() {
		var p Permintaan
		var tindakan Pemeriksaan
		if err := rows.Scan(&p.Nomor, &p.NoRawat, &p.Tanggal, &p.Jam,
			&p.TanggalDiterima, &p.JamDiterima, &p.TanggalHasil, &p.JamHasil, &p.KodeDokter,
			&p.NamaDokter, &p.StatusRawat, &p.InformasiTambahan, &p.DiagnosisKlinis,
			&tindakan.Kode, &tindakan.Nama, &tindakan.KodeCaraBayar, &tindakan.Kelas,
			&tindakan.Total, &tindakan.StatusBayar); err != nil {
			return nil, fmt.Errorf("scan permintaan radiologi: %w", err)
		}
		posisi, ada := indeks[p.Nomor]
		if !ada {
			p.Pemeriksaan = make([]Pemeriksaan, 0)
			p.DapatDiubah = true
			p.DapatDihapus = true
			daftar = append(daftar, p)
			posisi = len(daftar) - 1
			indeks[p.Nomor] = posisi
		}
		daftar[posisi].Pemeriksaan = append(daftar[posisi].Pemeriksaan, tindakan)
		daftar[posisi].Total += tindakan.Total
		perbaruiStatusPermintaan(&daftar[posisi])
	}
	return daftar, rows.Err()
}

func (r *Repositori) dokterPerujuk(ctx context.Context, noRawat string) (Dokter, error) {
	var item Dokter
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT d.kd_dokter, COALESCE(d.nm_dokter,''), COALESCE(s.nm_sps,'')
		FROM reg_periksa rp
		INNER JOIN dokter d ON d.kd_dokter=rp.kd_dokter
		LEFT JOIN spesialis s ON s.kd_sps=d.kd_sps
		WHERE rp.no_rawat=? LIMIT 1
	`, noRawat).Scan(&item.Kode, &item.Nama, &item.Spesialis)
	return item, err
}

func (r *Repositori) CariDokter(ctx context.Context, kata string) ([]Dokter, error) {
	seperti := "%" + strings.TrimSpace(kata) + "%"
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT d.kd_dokter, COALESCE(d.nm_dokter,''), COALESCE(s.nm_sps,'')
		FROM dokter d LEFT JOIN spesialis s ON s.kd_sps=d.kd_sps
		WHERE d.status='1' AND (d.kd_dokter LIKE ? OR d.nm_dokter LIKE ? OR s.nm_sps LIKE ?)
		ORDER BY d.nm_dokter LIMIT 30
	`, seperti, seperti, seperti)
	if err != nil {
		return nil, fmt.Errorf("cari dokter perujuk: %w", err)
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

func (r *Repositori) CariTindakan(ctx context.Context, noRawat, kata string) ([]Tindakan, error) {
	seperti := "%" + strings.TrimSpace(kata) + "%"
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT kd_jenis_prw, COALESCE(nm_perawatan,''), kd_pj, kelas, COALESCE(total_byr,0)
		FROM jns_perawatan_radiologi
		WHERE status='1'
			AND (kd_jenis_prw LIKE ? OR nm_perawatan LIKE ?)
		ORDER BY nm_perawatan LIMIT 50
	`, seperti, seperti)
	if err != nil {
		return nil, fmt.Errorf("cari tindakan radiologi: %w", err)
	}
	defer rows.Close()
	daftar := make([]Tindakan, 0)
	for rows.Next() {
		var item Tindakan
		if err := rows.Scan(&item.Kode, &item.Nama, &item.KodeCaraBayar, &item.Kelas, &item.Total); err != nil {
			return nil, err
		}
		daftar = append(daftar, item)
	}
	return daftar, rows.Err()
}

func (r *Repositori) Simpan(ctx context.Context, input Input) (string, error) {
	var errTerakhir error
	for percobaan := 0; percobaan < 4; percobaan++ {
		nomor, err := r.simpanSekali(ctx, input)
		if err == nil {
			return nomor, nil
		}
		var mysqlErr *mysql.MySQLError
		if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
			return "", err
		}
		errTerakhir = err
	}
	return "", errTerakhir
}

func (r *Repositori) Ubah(ctx context.Context, nomor string, input Input) error {
	tx, err := r.simrsDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	terkunci, err := r.billingTerkunci(ctx, tx, input.NoRawat)
	if err != nil {
		return err
	}
	if terkunci {
		return ErrBillingTerkunci
	}
	var jumlah, sudah, sudahDiterima int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(CASE WHEN ppr.stts_bayar='Sudah' THEN 1 ELSE 0 END),0),
			COALESCE(MAX(CASE WHEN pr.tgl_sampel<>'0000-00-00' OR pr.jam_sampel<>'00:00:00'
				OR pr.tgl_hasil<>'0000-00-00' OR pr.jam_hasil<>'00:00:00' THEN 1 ELSE 0 END),0)
		FROM permintaan_radiologi pr
		LEFT JOIN permintaan_pemeriksaan_radiologi ppr ON ppr.noorder=pr.noorder
		WHERE pr.noorder=? AND pr.no_rawat=?
	`, nomor, input.NoRawat).Scan(&jumlah, &sudah, &sudahDiterima); err != nil {
		return err
	}
	if jumlah == 0 {
		return ErrTidakDitemukan
	}
	if sudah > 0 {
		return ErrSudahDiproses
	}
	if sudahDiterima > 0 {
		return ErrSudahDiterima
	}
	lingkup, err := r.lingkup(ctx, tx, input.NoRawat)
	if err != nil {
		return err
	}
	for _, kode := range input.KodeTindakan {
		if _, err := r.tindakan(ctx, tx, lingkup, kode); err != nil {
			return err
		}
	}
	hasil, err := tx.ExecContext(ctx, `UPDATE permintaan_radiologi SET
		tgl_permintaan=?,jam_permintaan=?,dokter_perujuk=?,status=?,informasi_tambahan=?,diagnosa_klinis=?
		WHERE noorder=? AND no_rawat=?`, input.Tanggal, input.Jam, input.KodeDokter,
		lingkup.StatusRawat, input.InformasiTambahan, input.DiagnosisKlinis, nomor, input.NoRawat)
	if err != nil {
		return err
	}
	jumlahTerubah, err := hasil.RowsAffected()
	if err != nil {
		return err
	}
	if jumlahTerubah == 0 {
		var tetapAda int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM permintaan_radiologi WHERE noorder=? AND no_rawat=?`, nomor, input.NoRawat).Scan(&tetapAda); err != nil {
			return err
		}
		if tetapAda == 0 {
			return ErrTidakDitemukan
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM permintaan_pemeriksaan_radiologi WHERE noorder=?`, nomor); err != nil {
		return err
	}
	for _, kode := range input.KodeTindakan {
		if _, err := tx.ExecContext(ctx, `INSERT INTO permintaan_pemeriksaan_radiologi (noorder,kd_jenis_prw,stts_bayar) VALUES (?,?,'Belum')`, nomor, kode); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repositori) simpanSekali(ctx context.Context, input Input) (string, error) {
	tx, err := r.simrsDB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	terkunci, err := r.billingTerkunci(ctx, tx, input.NoRawat)
	if err != nil {
		return "", err
	}
	if terkunci {
		return "", ErrBillingTerkunci
	}
	lingkup, err := r.lingkup(ctx, tx, input.NoRawat)
	if err != nil {
		return "", err
	}
	nomor, err := r.nomorBerikutnya(ctx, tx, input.Tanggal)
	if err != nil {
		return "", err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO permintaan_radiologi
		(noorder,no_rawat,tgl_permintaan,jam_permintaan,tgl_sampel,jam_sampel,tgl_hasil,jam_hasil,dokter_perujuk,status,informasi_tambahan,diagnosa_klinis)
		VALUES (?,?,?,?,'0000-00-00','00:00:00','0000-00-00','00:00:00',?,?,?,?)`,
		nomor, input.NoRawat, input.Tanggal, input.Jam, input.KodeDokter,
		lingkup.StatusRawat, input.InformasiTambahan, input.DiagnosisKlinis)
	if err != nil {
		return "", err
	}
	for _, kode := range input.KodeTindakan {
		if _, err := r.tindakan(ctx, tx, lingkup, kode); err != nil {
			return "", err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO permintaan_pemeriksaan_radiologi (noorder,kd_jenis_prw,stts_bayar) VALUES (?,?,'Belum')`, nomor, kode); err != nil {
			return "", err
		}
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return nomor, nil
}

func (r *Repositori) Hapus(ctx context.Context, noRawat, nomor string) error {
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
	var jumlahSudah, sudahDiterima int
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(CASE WHEN ppr.stts_bayar='Sudah' THEN 1 ELSE 0 END),0),
			COALESCE(MAX(CASE WHEN pr.tgl_sampel<>'0000-00-00' OR pr.jam_sampel<>'00:00:00'
				OR pr.tgl_hasil<>'0000-00-00' OR pr.jam_hasil<>'00:00:00' THEN 1 ELSE 0 END),0)
		FROM permintaan_radiologi pr
		LEFT JOIN permintaan_pemeriksaan_radiologi ppr ON ppr.noorder=pr.noorder
		WHERE pr.noorder=? AND pr.no_rawat=?
	`, nomor, noRawat).Scan(&jumlahSudah, &sudahDiterima); err != nil {
		return err
	}
	if jumlahSudah > 0 {
		return ErrSudahDiproses
	}
	if sudahDiterima > 0 {
		return ErrSudahDiterima
	}
	hasil, err := tx.ExecContext(ctx, `DELETE FROM permintaan_radiologi WHERE noorder=? AND no_rawat=?`, nomor, noRawat)
	if err != nil {
		return err
	}
	jumlah, err := hasil.RowsAffected()
	if err != nil {
		return err
	}
	if jumlah == 0 {
		return ErrTidakDitemukan
	}
	return tx.Commit()
}

func (r *Repositori) tindakan(ctx context.Context, q queryer, lingkup lingkupTarif, kode string) (Tindakan, error) {
	var item Tindakan
	err := q.QueryRowContext(ctx, `
		SELECT kd_jenis_prw, COALESCE(nm_perawatan,''), kd_pj, kelas, COALESCE(total_byr,0)
		FROM jns_perawatan_radiologi
		WHERE kd_jenis_prw=? AND status='1'
			AND kd_pj=?
		LIMIT 1
	`, kode, lingkup.KodeCaraBayar).Scan(
		&item.Kode, &item.Nama, &item.KodeCaraBayar, &item.Kelas, &item.Total)
	if errors.Is(err, sql.ErrNoRows) {
		return item, fmt.Errorf("tindakan radiologi %s tidak tersedia untuk pasien", kode)
	}
	return item, err
}

func (r *Repositori) lingkup(ctx context.Context, q queryer, noRawat string) (lingkupTarif, error) {
	var hasil lingkupTarif
	var kodeCaraBayarPasien string
	var statusLanjut string
	if err := q.QueryRowContext(ctx, `SELECT kd_pj,status_lanjut FROM reg_periksa WHERE no_rawat=? LIMIT 1`, noRawat).Scan(&kodeCaraBayarPasien, &statusLanjut); err != nil {
		return hasil, err
	}
	hasil.KodeCaraBayar = kodeCaraBayarTarifRadiologi(kodeCaraBayarPasien)
	hasil.StatusRawat = strings.ToLower(statusLanjut)
	if strings.EqualFold(statusLanjut, "Ralan") {
		hasil.Kelas = "Rawat Jalan"
	} else {
		noRawatKamar := noRawat
		var induk string
		if err := q.QueryRowContext(ctx, `SELECT no_rawat FROM ranap_gabung WHERE no_rawat2=? LIMIT 1`, noRawat).Scan(&induk); err == nil && induk != "" {
			noRawatKamar = induk
		}
		if err := q.QueryRowContext(ctx, `
			SELECT k.kelas FROM kamar_inap ki INNER JOIN kamar k ON k.kd_kamar=ki.kd_kamar
			WHERE ki.no_rawat=? ORDER BY STR_TO_DATE(CONCAT(ki.tgl_masuk,' ',ki.jam_masuk),'%Y-%m-%d %H:%i:%s') DESC LIMIT 1
		`, noRawatKamar).Scan(&hasil.Kelas); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return hasil, err
		}
	}
	hasil.FilterCaraBayar = true
	hasil.FilterKelas = false
	return hasil, nil
}

func kodeCaraBayarTarifRadiologi(kode string) string {
	if strings.EqualFold(strings.TrimSpace(kode), "BPJ") || strings.TrimSpace(kode) == "36" {
		return "BPJ"
	}
	return "A09"
}

func (r *Repositori) nomorBerikutnya(ctx context.Context, q queryer, tanggal string) (string, error) {
	var terakhir int
	if err := q.QueryRowContext(ctx, `SELECT COALESCE(MAX(CONVERT(RIGHT(noorder,4),UNSIGNED)),0) FROM permintaan_radiologi WHERE tgl_permintaan=?`, tanggal).Scan(&terakhir); err != nil {
		return "", err
	}
	prefix := "PR" + strings.ReplaceAll(tanggal, "-", "")
	return fmt.Sprintf("%s%04d", prefix, terakhir+1), nil
}

func (r *Repositori) billingTerkunci(ctx context.Context, q queryer, noRawat string) (bool, error) {
	var jumlah int
	err := q.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM billing WHERE no_rawat=?) + (SELECT COUNT(*) FROM reg_periksa WHERE no_rawat=? AND stts='Batal')`, noRawat, noRawat).Scan(&jumlah)
	return jumlah > 0, err
}

func perbaruiStatusPermintaan(permintaan *Permintaan) {
	permintaan.StatusPemeriksaan = "Menunggu Radiologi"
	if permintaan.TanggalDiterima != "" || permintaan.JamDiterima != "" {
		permintaan.StatusPemeriksaan = "Sedang Dikerjakan"
	}
	if permintaan.TanggalHasil != "" || permintaan.JamHasil != "" {
		permintaan.StatusPemeriksaan = "Selesai"
	}

	jumlahDibayar := 0
	for _, pemeriksaan := range permintaan.Pemeriksaan {
		if strings.EqualFold(pemeriksaan.StatusBayar, "Sudah") {
			jumlahDibayar++
		}
	}
	permintaan.StatusBayar = "Belum Bayar"
	if jumlahDibayar > 0 && jumlahDibayar < len(permintaan.Pemeriksaan) {
		permintaan.StatusBayar = "Sebagian Dibayar"
	} else if jumlahDibayar > 0 && jumlahDibayar == len(permintaan.Pemeriksaan) {
		permintaan.StatusBayar = "Sudah Bayar"
	}

	terkunci := jumlahDibayar > 0 || permintaan.StatusPemeriksaan != "Menunggu Radiologi"
	permintaan.DapatDiubah = !terkunci
	permintaan.DapatDihapus = !terkunci
}
