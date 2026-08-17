package permintaan_laboratorium

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrTidakDitemukan  = errors.New("permintaan laboratorium tidak ditemukan")
	ErrBillingTerkunci = errors.New("billing sudah terverifikasi atau kunjungan dibatalkan")
	ErrSudahDiproses   = errors.New("permintaan laboratorium sudah diproses")
	ErrSudahDiterima   = errors.New("permintaan laboratorium sudah diterima petugas laboratorium")
)

type Dokter struct {
	Kode      string `json:"kode"`
	Nama      string `json:"nama"`
	Spesialis string `json:"spesialis"`
}

type DetailPemeriksaan struct {
	ID           int64   `json:"id"`
	Nama         string  `json:"nama"`
	Satuan       string  `json:"satuan"`
	NilaiRujukan string  `json:"nilai_rujukan"`
	Biaya        float64 `json:"biaya"`
	StatusBayar  string  `json:"status_bayar"`
}

type Tindakan struct {
	Kode          string  `json:"kode"`
	Nama          string  `json:"nama"`
	KodeCaraBayar string  `json:"kode_cara_bayar"`
	Kelas         string  `json:"kelas"`
	Total         float64 `json:"total"`
	JumlahDetail  int     `json:"jumlah_detail"`
}

type Pemeriksaan struct {
	Tindakan
	StatusBayar string              `json:"status_bayar"`
	Detail      []DetailPemeriksaan `json:"detail"`
}

type Permintaan struct {
	Nomor             string        `json:"nomor"`
	NoRawat           string        `json:"no_rawat"`
	Tanggal           string        `json:"tanggal"`
	Jam               string        `json:"jam"`
	TanggalSampel     string        `json:"tanggal_sampel"`
	JamSampel         string        `json:"jam_sampel"`
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
}

type PilihanPemeriksaan struct {
	Kode     string  `json:"kode"`
	IDDetail []int64 `json:"id_detail"`
}

type Input struct {
	NoRawat           string               `json:"no_rawat"`
	Tanggal           string               `json:"tanggal"`
	Jam               string               `json:"jam"`
	KodeDokter        string               `json:"kode_dokter"`
	InformasiTambahan string               `json:"informasi_tambahan"`
	DiagnosisKlinis   string               `json:"diagnosis_klinis"`
	Pemeriksaan       []PilihanPemeriksaan `json:"pemeriksaan"`
}

type lingkupTarif struct {
	KodeCaraBayar string
	Kelas         string
	StatusRawat   string
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
	hasil := Data{Permintaan: permintaan, BillingTerkunci: terkunci, StatusRawat: lingkup.StatusRawat, KodeCaraBayar: lingkup.KodeCaraBayar, KelasPasien: lingkup.Kelas}
	if dokter, err := r.dokterPerujuk(ctx, noRawat); err == nil {
		hasil.DokterPerujuk = &dokter
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Data{}, err
	}
	return hasil, nil
}

func (r *Repositori) daftar(ctx context.Context, noRawat string) ([]Permintaan, error) {
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT pl.noorder,pl.no_rawat,DATE_FORMAT(pl.tgl_permintaan,'%Y-%m-%d'),TIME_FORMAT(pl.jam_permintaan,'%H:%i:%s'),
			IF(pl.tgl_sampel='0000-00-00','',DATE_FORMAT(pl.tgl_sampel,'%Y-%m-%d')),
			IF(pl.jam_sampel='00:00:00','',TIME_FORMAT(pl.jam_sampel,'%H:%i:%s')),
			IF(pl.tgl_hasil='0000-00-00','',DATE_FORMAT(pl.tgl_hasil,'%Y-%m-%d')),
			IF(pl.jam_hasil='00:00:00','',TIME_FORMAT(pl.jam_hasil,'%H:%i:%s')),
			pl.dokter_perujuk,COALESCE(d.nm_dokter,''),pl.status,pl.informasi_tambahan,pl.diagnosa_klinis,
			ppl.kd_jenis_prw,COALESCE(j.nm_perawatan,''),j.kd_pj,j.kelas,COALESCE(j.total_byr,0),COALESCE(ppl.stts_bayar,'Belum')
		FROM permintaan_lab pl
		INNER JOIN dokter d ON d.kd_dokter=pl.dokter_perujuk
		INNER JOIN permintaan_pemeriksaan_lab ppl ON ppl.noorder=pl.noorder
		INNER JOIN jns_perawatan_lab j ON j.kd_jenis_prw=ppl.kd_jenis_prw
		WHERE pl.no_rawat=? AND j.kategori='PK'
		ORDER BY pl.tgl_permintaan DESC,pl.jam_permintaan DESC,j.nm_perawatan
	`, noRawat)
	if err != nil {
		return nil, fmt.Errorf("baca permintaan laboratorium: %w", err)
	}
	defer rows.Close()

	daftar := make([]Permintaan, 0)
	indeks := make(map[string]int)
	for rows.Next() {
		var p Permintaan
		var tindakan Pemeriksaan
		if err := rows.Scan(&p.Nomor, &p.NoRawat, &p.Tanggal, &p.Jam, &p.TanggalSampel, &p.JamSampel,
			&p.TanggalHasil, &p.JamHasil, &p.KodeDokter, &p.NamaDokter, &p.StatusRawat,
			&p.InformasiTambahan, &p.DiagnosisKlinis, &tindakan.Kode, &tindakan.Nama,
			&tindakan.KodeCaraBayar, &tindakan.Kelas, &tindakan.Total, &tindakan.StatusBayar); err != nil {
			return nil, fmt.Errorf("scan permintaan laboratorium: %w", err)
		}
		posisi, ada := indeks[p.Nomor]
		if !ada {
			p.Pemeriksaan = make([]Pemeriksaan, 0)
			daftar = append(daftar, p)
			posisi = len(daftar) - 1
			indeks[p.Nomor] = posisi
		}
		tindakan.Detail, err = r.detail(ctx, p.Nomor, tindakan.Kode)
		if err != nil {
			return nil, err
		}
		tindakan.JumlahDetail = len(tindakan.Detail)
		daftar[posisi].Pemeriksaan = append(daftar[posisi].Pemeriksaan, tindakan)
		daftar[posisi].Total += tindakan.Total
		perbaruiStatusPermintaan(&daftar[posisi])
	}
	return daftar, rows.Err()
}

func (r *Repositori) detail(ctx context.Context, nomor, kode string) ([]DetailPemeriksaan, error) {
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT t.id_template,t.Pemeriksaan,t.satuan,
			CONCAT_WS(', ',NULLIF(CONCAT('LD: ',t.nilai_rujukan_ld),'LD: '),NULLIF(CONCAT('LA: ',t.nilai_rujukan_la),'LA: '),NULLIF(CONCAT('PD: ',t.nilai_rujukan_pd),'PD: '),NULLIF(CONCAT('PA: ',t.nilai_rujukan_pa),'PA: ')),
			COALESCE(t.biaya_item,0),COALESCE(d.stts_bayar,'Belum')
		FROM permintaan_detail_permintaan_lab d
		INNER JOIN template_laboratorium t ON t.id_template=d.id_template
		WHERE d.noorder=? AND d.kd_jenis_prw=? ORDER BY COALESCE(t.urut,9999),t.Pemeriksaan
	`, nomor, kode)
	if err != nil {
		return nil, fmt.Errorf("baca detail permintaan laboratorium: %w", err)
	}
	defer rows.Close()
	hasil := make([]DetailPemeriksaan, 0)
	for rows.Next() {
		var item DetailPemeriksaan
		if err := rows.Scan(&item.ID, &item.Nama, &item.Satuan, &item.NilaiRujukan, &item.Biaya, &item.StatusBayar); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *Repositori) dokterPerujuk(ctx context.Context, noRawat string) (Dokter, error) {
	var item Dokter
	err := r.simrsDB.QueryRowContext(ctx, `SELECT d.kd_dokter,COALESCE(d.nm_dokter,''),COALESCE(s.nm_sps,'') FROM reg_periksa rp INNER JOIN dokter d ON d.kd_dokter=rp.kd_dokter LEFT JOIN spesialis s ON s.kd_sps=d.kd_sps WHERE rp.no_rawat=? LIMIT 1`, noRawat).Scan(&item.Kode, &item.Nama, &item.Spesialis)
	return item, err
}

func (r *Repositori) CariDokter(ctx context.Context, kata string) ([]Dokter, error) {
	seperti := "%" + strings.TrimSpace(kata) + "%"
	rows, err := r.simrsDB.QueryContext(ctx, `SELECT d.kd_dokter,COALESCE(d.nm_dokter,''),COALESCE(s.nm_sps,'') FROM dokter d LEFT JOIN spesialis s ON s.kd_sps=d.kd_sps WHERE d.status='1' AND (d.kd_dokter LIKE ? OR d.nm_dokter LIKE ? OR s.nm_sps LIKE ?) ORDER BY d.nm_dokter LIMIT 30`, seperti, seperti, seperti)
	if err != nil {
		return nil, fmt.Errorf("cari dokter perujuk laboratorium: %w", err)
	}
	defer rows.Close()
	hasil := make([]Dokter, 0)
	for rows.Next() {
		var item Dokter
		if err := rows.Scan(&item.Kode, &item.Nama, &item.Spesialis); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *Repositori) CariTindakan(ctx context.Context, noRawat, kata string) ([]Tindakan, error) {
	lingkup, err := r.lingkup(ctx, r.simrsDB, noRawat)
	if err != nil {
		return nil, err
	}
	seperti := "%" + strings.TrimSpace(kata) + "%"
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT j.kd_jenis_prw,COALESCE(j.nm_perawatan,''),j.kd_pj,j.kelas,COALESCE(j.total_byr,0),COUNT(t.id_template)
		FROM jns_perawatan_lab j LEFT JOIN template_laboratorium t ON t.kd_jenis_prw=j.kd_jenis_prw
		WHERE j.status='1' AND j.kategori='PK' AND j.kd_pj=? AND (j.kd_jenis_prw LIKE ? OR j.nm_perawatan LIKE ?)
		GROUP BY j.kd_jenis_prw,j.nm_perawatan,j.kd_pj,j.kelas,j.total_byr ORDER BY j.nm_perawatan LIMIT 50
	`, lingkup.KodeCaraBayar, seperti, seperti)
	if err != nil {
		return nil, fmt.Errorf("cari tindakan laboratorium: %w", err)
	}
	defer rows.Close()
	hasil := make([]Tindakan, 0)
	for rows.Next() {
		var item Tindakan
		if err := rows.Scan(&item.Kode, &item.Nama, &item.KodeCaraBayar, &item.Kelas, &item.Total, &item.JumlahDetail); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *Repositori) DetailTindakan(ctx context.Context, noRawat, kode string) ([]DetailPemeriksaan, error) {
	lingkup, err := r.lingkup(ctx, r.simrsDB, noRawat)
	if err != nil {
		return nil, err
	}
	if _, err := r.tindakan(ctx, r.simrsDB, lingkup, kode); err != nil {
		return nil, err
	}
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT id_template,Pemeriksaan,satuan,
			CONCAT_WS(', ',NULLIF(CONCAT('LD: ',nilai_rujukan_ld),'LD: '),NULLIF(CONCAT('LA: ',nilai_rujukan_la),'LA: '),NULLIF(CONCAT('PD: ',nilai_rujukan_pd),'PD: '),NULLIF(CONCAT('PA: ',nilai_rujukan_pa),'PA: ')),
			COALESCE(biaya_item,0)
		FROM template_laboratorium WHERE kd_jenis_prw=? ORDER BY COALESCE(urut,9999),Pemeriksaan
	`, kode)
	if err != nil {
		return nil, fmt.Errorf("baca pilihan detail laboratorium: %w", err)
	}
	defer rows.Close()
	hasil := make([]DetailPemeriksaan, 0)
	for rows.Next() {
		var item DetailPemeriksaan
		if err := rows.Scan(&item.ID, &item.Nama, &item.Satuan, &item.NilaiRujukan, &item.Biaya); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *Repositori) Simpan(ctx context.Context, input Input) (string, error) {
	var terakhir error
	for i := 0; i < 4; i++ {
		nomor, err := r.simpanSekali(ctx, input)
		if err == nil {
			return nomor, nil
		}
		var mysqlErr *mysql.MySQLError
		if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
			return "", err
		}
		terakhir = err
	}
	return "", terakhir
}

func (r *Repositori) simpanSekali(ctx context.Context, input Input) (string, error) {
	tx, err := r.simrsDB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	if terkunci, err := r.billingTerkunci(ctx, tx, input.NoRawat); err != nil {
		return "", err
	} else if terkunci {
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
	if _, err = tx.ExecContext(ctx, `INSERT INTO permintaan_lab (noorder,no_rawat,tgl_permintaan,jam_permintaan,tgl_sampel,jam_sampel,tgl_hasil,jam_hasil,dokter_perujuk,status,informasi_tambahan,diagnosa_klinis) VALUES (?,?,?,?,'0000-00-00','00:00:00','0000-00-00','00:00:00',?,?,?,?)`, nomor, input.NoRawat, input.Tanggal, input.Jam, input.KodeDokter, lingkup.StatusRawat, input.InformasiTambahan, input.DiagnosisKlinis); err != nil {
		return "", err
	}
	for _, pemeriksaan := range input.Pemeriksaan {
		if _, err := r.tindakan(ctx, tx, lingkup, pemeriksaan.Kode); err != nil {
			return "", err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO permintaan_pemeriksaan_lab (noorder,kd_jenis_prw,stts_bayar) VALUES (?,?,'Belum')`, nomor, pemeriksaan.Kode); err != nil {
			return "", err
		}
		if err := r.simpanDetail(ctx, tx, nomor, pemeriksaan.Kode, pemeriksaan.IDDetail); err != nil {
			return "", err
		}
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return nomor, nil
}

func (r *Repositori) Ubah(ctx context.Context, nomor string, input Input) error {
	tx, err := r.simrsDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if terkunci, err := r.billingTerkunci(ctx, tx, input.NoRawat); err != nil {
		return err
	} else if terkunci {
		return ErrBillingTerkunci
	}
	var jumlah, dibayar, diterima int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(CASE WHEN ppl.stts_bayar='Sudah' THEN 1 ELSE 0 END),0),COALESCE(MAX(CASE WHEN pl.tgl_sampel<>'0000-00-00' OR pl.jam_sampel<>'00:00:00' OR pl.tgl_hasil<>'0000-00-00' OR pl.jam_hasil<>'00:00:00' THEN 1 ELSE 0 END),0) FROM permintaan_lab pl LEFT JOIN permintaan_pemeriksaan_lab ppl ON ppl.noorder=pl.noorder WHERE pl.noorder=? AND pl.no_rawat=?`, nomor, input.NoRawat).Scan(&jumlah, &dibayar, &diterima); err != nil {
		return err
	}
	if jumlah == 0 {
		return ErrTidakDitemukan
	}
	if dibayar > 0 {
		return ErrSudahDiproses
	}
	if diterima > 0 {
		return ErrSudahDiterima
	}
	lingkup, err := r.lingkup(ctx, tx, input.NoRawat)
	if err != nil {
		return err
	}
	for _, pemeriksaan := range input.Pemeriksaan {
		if _, err := r.tindakan(ctx, tx, lingkup, pemeriksaan.Kode); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE permintaan_lab SET tgl_permintaan=?,jam_permintaan=?,dokter_perujuk=?,status=?,informasi_tambahan=?,diagnosa_klinis=? WHERE noorder=? AND no_rawat=?`, input.Tanggal, input.Jam, input.KodeDokter, lingkup.StatusRawat, input.InformasiTambahan, input.DiagnosisKlinis, nomor, input.NoRawat); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM permintaan_detail_permintaan_lab WHERE noorder=?`, nomor); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM permintaan_pemeriksaan_lab WHERE noorder=?`, nomor); err != nil {
		return err
	}
	for _, pemeriksaan := range input.Pemeriksaan {
		if _, err := tx.ExecContext(ctx, `INSERT INTO permintaan_pemeriksaan_lab (noorder,kd_jenis_prw,stts_bayar) VALUES (?,?,'Belum')`, nomor, pemeriksaan.Kode); err != nil {
			return err
		}
		if err := r.simpanDetail(ctx, tx, nomor, pemeriksaan.Kode, pemeriksaan.IDDetail); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repositori) Hapus(ctx context.Context, noRawat, nomor string) error {
	tx, err := r.simrsDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if terkunci, err := r.billingTerkunci(ctx, tx, noRawat); err != nil {
		return err
	} else if terkunci {
		return ErrBillingTerkunci
	}
	var dibayar, diterima int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(CASE WHEN ppl.stts_bayar='Sudah' THEN 1 ELSE 0 END),0),COALESCE(MAX(CASE WHEN pl.tgl_sampel<>'0000-00-00' OR pl.jam_sampel<>'00:00:00' OR pl.tgl_hasil<>'0000-00-00' OR pl.jam_hasil<>'00:00:00' THEN 1 ELSE 0 END),0) FROM permintaan_lab pl LEFT JOIN permintaan_pemeriksaan_lab ppl ON ppl.noorder=pl.noorder WHERE pl.noorder=? AND pl.no_rawat=?`, nomor, noRawat).Scan(&dibayar, &diterima); err != nil {
		return err
	}
	if dibayar > 0 {
		return ErrSudahDiproses
	}
	if diterima > 0 {
		return ErrSudahDiterima
	}
	hasil, err := tx.ExecContext(ctx, `DELETE FROM permintaan_lab WHERE noorder=? AND no_rawat=?`, nomor, noRawat)
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

func (r *Repositori) simpanDetail(ctx context.Context, tx *sql.Tx, nomor, kode string, idDetail []int64) error {
	for _, id := range idDetail {
		hasil, err := tx.ExecContext(ctx, `INSERT INTO permintaan_detail_permintaan_lab (noorder,kd_jenis_prw,id_template,stts_bayar) SELECT ?,kd_jenis_prw,id_template,'Belum' FROM template_laboratorium WHERE kd_jenis_prw=? AND id_template=?`, nomor, kode, id)
		if err != nil {
			return err
		}
		jumlah, err := hasil.RowsAffected()
		if err != nil {
			return err
		}
		if jumlah == 0 {
			return fmt.Errorf("detail laboratorium %d tidak tersedia untuk tindakan %s", id, kode)
		}
	}
	return nil
}

func (r *Repositori) tindakan(ctx context.Context, q queryer, lingkup lingkupTarif, kode string) (Tindakan, error) {
	var item Tindakan
	err := q.QueryRowContext(ctx, `SELECT kd_jenis_prw,COALESCE(nm_perawatan,''),kd_pj,kelas,COALESCE(total_byr,0) FROM jns_perawatan_lab WHERE kd_jenis_prw=? AND status='1' AND kategori='PK' AND kd_pj=? LIMIT 1`, kode, lingkup.KodeCaraBayar).Scan(&item.Kode, &item.Nama, &item.KodeCaraBayar, &item.Kelas, &item.Total)
	if errors.Is(err, sql.ErrNoRows) {
		return item, fmt.Errorf("tindakan laboratorium %s tidak tersedia untuk pasien", kode)
	}
	return item, err
}

func (r *Repositori) lingkup(ctx context.Context, q queryer, noRawat string) (lingkupTarif, error) {
	var hasil lingkupTarif
	var kodeBayar, status string
	if err := q.QueryRowContext(ctx, `SELECT kd_pj,status_lanjut FROM reg_periksa WHERE no_rawat=? LIMIT 1`, noRawat).Scan(&kodeBayar, &status); err != nil {
		return hasil, err
	}
	hasil.KodeCaraBayar = kodeCaraBayarTarif(kodeBayar)
	hasil.StatusRawat = strings.ToLower(status)
	if strings.EqualFold(status, "Ralan") {
		hasil.Kelas = "Rawat Jalan"
	} else {
		noRawatKamar := noRawat
		var induk string
		if err := q.QueryRowContext(ctx, `SELECT no_rawat FROM ranap_gabung WHERE no_rawat2=? LIMIT 1`, noRawat).Scan(&induk); err == nil && induk != "" {
			noRawatKamar = induk
		}
		_ = q.QueryRowContext(ctx, `SELECT k.kelas FROM kamar_inap ki INNER JOIN kamar k ON k.kd_kamar=ki.kd_kamar WHERE ki.no_rawat=? ORDER BY STR_TO_DATE(CONCAT(ki.tgl_masuk,' ',ki.jam_masuk),'%Y-%m-%d %H:%i:%s') DESC LIMIT 1`, noRawatKamar).Scan(&hasil.Kelas)
	}
	return hasil, nil
}

func kodeCaraBayarTarif(kode string) string {
	if strings.EqualFold(strings.TrimSpace(kode), "BPJ") || strings.TrimSpace(kode) == "36" {
		return "BPJ"
	}
	return "A09"
}

func (r *Repositori) nomorBerikutnya(ctx context.Context, q queryer, tanggal string) (string, error) {
	var terakhir int
	if err := q.QueryRowContext(ctx, `SELECT COALESCE(MAX(CONVERT(RIGHT(noorder,4),UNSIGNED)),0) FROM permintaan_lab WHERE tgl_permintaan=?`, tanggal).Scan(&terakhir); err != nil {
		return "", err
	}
	return fmt.Sprintf("PK%s%04d", strings.ReplaceAll(tanggal, "-", ""), terakhir+1), nil
}

func (r *Repositori) billingTerkunci(ctx context.Context, q queryer, noRawat string) (bool, error) {
	var jumlah int
	err := q.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM billing WHERE no_rawat=?) + (SELECT COUNT(*) FROM reg_periksa WHERE no_rawat=? AND stts='Batal')`, noRawat, noRawat).Scan(&jumlah)
	return jumlah > 0, err
}

func perbaruiStatusPermintaan(p *Permintaan) {
	p.StatusPemeriksaan = "Menunggu Laboratorium"
	if p.TanggalSampel != "" || p.JamSampel != "" {
		p.StatusPemeriksaan = "Sampel Diterima"
	}
	if p.TanggalHasil != "" || p.JamHasil != "" {
		p.StatusPemeriksaan = "Selesai"
	}
	dibayar := 0
	for _, item := range p.Pemeriksaan {
		if strings.EqualFold(item.StatusBayar, "Sudah") {
			dibayar++
		}
	}
	p.StatusBayar = "Belum Bayar"
	if dibayar > 0 && dibayar < len(p.Pemeriksaan) {
		p.StatusBayar = "Sebagian Dibayar"
	} else if dibayar > 0 && dibayar == len(p.Pemeriksaan) {
		p.StatusBayar = "Sudah Bayar"
	}
	terkunci := dibayar > 0 || p.StatusPemeriksaan != "Menunggu Laboratorium"
	p.DapatDiubah, p.DapatDihapus = !terkunci, !terkunci
}
