package resume_pasien

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrTidakDitemukan  = errors.New("resume pasien rawat jalan tidak ditemukan")
	ErrSudahAda        = errors.New("resume pasien rawat jalan sudah tersedia")
	ErrBillingTerkunci = errors.New("billing sudah terverifikasi atau kunjungan dibatalkan")
)

type Dokter struct {
	Kode      string `json:"kode"`
	Nama      string `json:"nama"`
	Spesialis string `json:"spesialis"`
}

type Input struct {
	NoRawat               string `json:"no_rawat"`
	KodeDokter            string `json:"kode_dokter"`
	KondisiPulang         string `json:"kondisi_pulang"`
	KeluhanUtama          string `json:"keluhan_utama"`
	JalannyaPenyakit      string `json:"jalannya_penyakit"`
	PemeriksaanPenunjang  string `json:"pemeriksaan_penunjang"`
	HasilLaborat          string `json:"hasil_laborat"`
	DiagnosaUtama         string `json:"diagnosa_utama"`
	KodeDiagnosaUtama     string `json:"kd_diagnosa_utama"`
	DiagnosaSekunder      string `json:"diagnosa_sekunder"`
	KodeDiagnosaSekunder  string `json:"kd_diagnosa_sekunder"`
	DiagnosaSekunder2     string `json:"diagnosa_sekunder2"`
	KodeDiagnosaSekunder2 string `json:"kd_diagnosa_sekunder2"`
	DiagnosaSekunder3     string `json:"diagnosa_sekunder3"`
	KodeDiagnosaSekunder3 string `json:"kd_diagnosa_sekunder3"`
	DiagnosaSekunder4     string `json:"diagnosa_sekunder4"`
	KodeDiagnosaSekunder4 string `json:"kd_diagnosa_sekunder4"`
	ProsedurUtama         string `json:"prosedur_utama"`
	KodeProsedurUtama     string `json:"kd_prosedur_utama"`
	ProsedurSekunder      string `json:"prosedur_sekunder"`
	KodeProsedurSekunder  string `json:"kd_prosedur_sekunder"`
	ProsedurSekunder2     string `json:"prosedur_sekunder2"`
	KodeProsedurSekunder2 string `json:"kd_prosedur_sekunder2"`
	ProsedurSekunder3     string `json:"prosedur_sekunder3"`
	KodeProsedurSekunder3 string `json:"kd_prosedur_sekunder3"`
	ObatPulang            string `json:"obat_pulang"`
}

type Data struct {
	Tersedia        bool    `json:"tersedia"`
	BillingTerkunci bool    `json:"billing_terkunci"`
	Resume          Input   `json:"resume"`
	Dokter          *Dokter `json:"dokter,omitempty"`
	Pilihan         Pilihan `json:"pilihan"`
}

type Pilihan struct {
	KondisiPulang []string `json:"kondisi_pulang"`
}

type ReferensiResume struct {
	Tanggal string `json:"tanggal"`
	Jam     string `json:"jam"`
	Isi     string `json:"isi"`
	Sumber  string `json:"sumber"`
}

var kolom = []string{
	"no_rawat", "kd_dokter", "kondisi_pulang", "keluhan_utama", "jalannya_penyakit",
	"pemeriksaan_penunjang", "hasil_laborat", "diagnosa_utama", "kd_diagnosa_utama",
	"diagnosa_sekunder", "kd_diagnosa_sekunder", "diagnosa_sekunder2", "kd_diagnosa_sekunder2",
	"diagnosa_sekunder3", "kd_diagnosa_sekunder3", "diagnosa_sekunder4", "kd_diagnosa_sekunder4",
	"prosedur_utama", "kd_prosedur_utama", "prosedur_sekunder", "kd_prosedur_sekunder",
	"prosedur_sekunder2", "kd_prosedur_sekunder2", "prosedur_sekunder3", "kd_prosedur_sekunder3",
	"obat_pulang",
}

type Repositori struct{ simrsDB *sql.DB }

func NewRepositori(simrsDB *sql.DB) *Repositori { return &Repositori{simrsDB: simrsDB} }

func (r *Repositori) Data(ctx context.Context, noRawat string) (Data, error) {
	data := Data{
		Resume:  Input{NoRawat: noRawat, KondisiPulang: "Hidup"},
		Pilihan: Pilihan{KondisiPulang: []string{"Hidup", "Meninggal"}},
	}
	terkunci, err := r.billingTerkunci(ctx, r.simrsDB, noRawat)
	if err != nil {
		return Data{}, fmt.Errorf("periksa status billing pasien: %w", err)
	}
	data.BillingTerkunci = terkunci

	resume, dokter, err := r.ambil(ctx, noRawat)
	if err == nil {
		if err := r.isiCoding(ctx, &resume); err != nil {
			return Data{}, err
		}
		data.Tersedia, data.Resume, data.Dokter = true, resume, &dokter
		return data, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Data{}, err
	}
	if err := r.isiDefault(ctx, &data.Resume, &dokter); err != nil {
		return Data{}, err
	}
	if dokter.Kode != "" {
		data.Dokter = &dokter
	}
	return data, nil
}

func (r *Repositori) ambil(ctx context.Context, noRawat string) (Input, Dokter, error) {
	var item Input
	var dokter Dokter
	row := r.simrsDB.QueryRowContext(ctx, `SELECT `+prefiksKolom("r")+`,
		COALESCE(d.nm_dokter,''), COALESCE(s.nm_sps,'')
		FROM resume_pasien r
		LEFT JOIN dokter d ON d.kd_dokter=r.kd_dokter
		LEFT JOIN spesialis s ON s.kd_sps=d.kd_sps
		WHERE r.no_rawat=? LIMIT 1`, noRawat)
	err := scanResume(row, &item, &dokter)
	dokter.Kode = item.KodeDokter
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Input{}, Dokter{}, fmt.Errorf("baca resume pasien rawat jalan: %w", err)
	}
	return item, dokter, err
}

func (r *Repositori) isiDefault(ctx context.Context, item *Input, dokter *Dokter) error {
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT d.kd_dokter, COALESCE(d.nm_dokter,''), COALESCE(s.nm_sps,'')
		FROM reg_periksa rp
		INNER JOIN dokter d ON d.kd_dokter=rp.kd_dokter
		LEFT JOIN spesialis s ON s.kd_sps=d.kd_sps
		WHERE rp.no_rawat=? LIMIT 1
	`, item.NoRawat).Scan(&dokter.Kode, &dokter.Nama, &dokter.Spesialis)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("baca dokter penanggung jawab: %w", err)
	}
	item.KodeDokter = dokter.Kode
	return r.isiCoding(ctx, item)
}

func (r *Repositori) isiCoding(ctx context.Context, item *Input) error {
	if err := r.isiDiagnosa(ctx, item); err != nil {
		return err
	}
	return r.isiProsedur(ctx, item)
}

func (r *Repositori) isiDiagnosa(ctx context.Context, item *Input) error {
	item.KodeDiagnosaUtama, item.DiagnosaUtama = "", ""
	item.KodeDiagnosaSekunder, item.DiagnosaSekunder = "", ""
	item.KodeDiagnosaSekunder2, item.DiagnosaSekunder2 = "", ""
	item.KodeDiagnosaSekunder3, item.DiagnosaSekunder3 = "", ""
	item.KodeDiagnosaSekunder4, item.DiagnosaSekunder4 = "", ""
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT dp.prioritas, dp.kd_penyakit, COALESCE(p.nm_penyakit,'')
		FROM diagnosa_pasien dp
		LEFT JOIN penyakit p ON p.kd_penyakit=dp.kd_penyakit
		WHERE dp.no_rawat=?
		ORDER BY dp.prioritas, dp.kd_penyakit LIMIT 5
	`, item.NoRawat)
	if err != nil {
		return fmt.Errorf("baca diagnosa pasien: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var prioritas int
		var kode, nama string
		if err := rows.Scan(&prioritas, &kode, &nama); err != nil {
			return err
		}
		switch prioritas {
		case 1:
			item.KodeDiagnosaUtama, item.DiagnosaUtama = kode, nama
		case 2:
			item.KodeDiagnosaSekunder, item.DiagnosaSekunder = kode, nama
		case 3:
			item.KodeDiagnosaSekunder2, item.DiagnosaSekunder2 = kode, nama
		case 4:
			item.KodeDiagnosaSekunder3, item.DiagnosaSekunder3 = kode, nama
		case 5:
			item.KodeDiagnosaSekunder4, item.DiagnosaSekunder4 = kode, nama
		}
	}
	return rows.Err()
}

func (r *Repositori) isiProsedur(ctx context.Context, item *Input) error {
	item.KodeProsedurUtama, item.ProsedurUtama = "", ""
	item.KodeProsedurSekunder, item.ProsedurSekunder = "", ""
	item.KodeProsedurSekunder2, item.ProsedurSekunder2 = "", ""
	item.KodeProsedurSekunder3, item.ProsedurSekunder3 = "", ""
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT pp.prioritas, pp.kode, COALESCE(i.deskripsi_panjang,'')
		FROM prosedur_pasien pp
		LEFT JOIN icd9 i ON i.kode=pp.kode
		WHERE pp.no_rawat=?
		ORDER BY pp.prioritas, pp.kode LIMIT 4
	`, item.NoRawat)
	if err != nil {
		return fmt.Errorf("baca prosedur pasien: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var prioritas int
		var kode, nama string
		if err := rows.Scan(&prioritas, &kode, &nama); err != nil {
			return err
		}
		switch prioritas {
		case 1:
			item.KodeProsedurUtama, item.ProsedurUtama = kode, nama
		case 2:
			item.KodeProsedurSekunder, item.ProsedurSekunder = kode, nama
		case 3:
			item.KodeProsedurSekunder2, item.ProsedurSekunder2 = kode, nama
		case 4:
			item.KodeProsedurSekunder3, item.ProsedurSekunder3 = kode, nama
		}
	}
	return rows.Err()
}

func (r *Repositori) Simpan(ctx context.Context, input Input) error {
	return r.dalamTransaksi(ctx, input.NoRawat, func(tx *sql.Tx) error {
		if err := r.siapkanDanSinkronkanCoding(ctx, tx, &input); err != nil {
			return err
		}
		args := nilai(input)
		_, err := tx.ExecContext(ctx, `INSERT INTO resume_pasien (`+strings.Join(kolom, ",")+`) VALUES (`+placeholder(len(args))+`)`, args...)
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrSudahAda
		}
		if err != nil {
			return fmt.Errorf("simpan resume pasien rawat jalan: %w", err)
		}
		return nil
	})
}

func (r *Repositori) Ubah(ctx context.Context, input Input) error {
	return r.dalamTransaksi(ctx, input.NoRawat, func(tx *sql.Tx) error {
		if err := r.siapkanDanSinkronkanCoding(ctx, tx, &input); err != nil {
			return err
		}
		set := make([]string, 0, len(kolom)-1)
		for _, nama := range kolom[1:] {
			set = append(set, nama+"=?")
		}
		args := append([]any{}, nilai(input)[1:]...)
		args = append(args, input.NoRawat)
		hasil, err := tx.ExecContext(ctx, `UPDATE resume_pasien SET `+strings.Join(set, ",")+` WHERE no_rawat=?`, args...)
		if err != nil {
			return fmt.Errorf("ubah resume pasien rawat jalan: %w", err)
		}
		jumlah, err := hasil.RowsAffected()
		if err != nil {
			return err
		}
		if jumlah == 0 {
			var ada int
			if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM resume_pasien WHERE no_rawat=?`, input.NoRawat).Scan(&ada); err != nil {
				return err
			}
			if ada == 0 {
				return ErrTidakDitemukan
			}
		}
		return nil
	})
}

func (r *Repositori) Hapus(ctx context.Context, noRawat string) error {
	return r.dalamTransaksi(ctx, noRawat, func(tx *sql.Tx) error {
		hasil, err := tx.ExecContext(ctx, `DELETE FROM resume_pasien WHERE no_rawat=?`, noRawat)
		if err != nil {
			return fmt.Errorf("hapus resume pasien rawat jalan: %w", err)
		}
		jumlah, err := hasil.RowsAffected()
		if err != nil {
			return err
		}
		if jumlah == 0 {
			return ErrTidakDitemukan
		}
		return nil
	})
}

func (r *Repositori) Referensi(ctx context.Context, noRawat, jenis, kataKunci string) ([]ReferensiResume, error) {
	pola := "%" + strings.TrimSpace(kataKunci) + "%"
	switch jenis {
	case "keluhan":
		return r.referensiKeluhan(ctx, noRawat, pola)
	case "pemeriksaan":
		return r.referensiPemeriksaan(ctx, noRawat, pola)
	case "radiologi":
		return r.referensiRadiologi(ctx, noRawat, pola)
	case "laboratorium":
		return r.referensiLaboratorium(ctx, noRawat, pola)
	case "obat":
		return r.referensiObat(ctx, noRawat, pola)
	default:
		return nil, fmt.Errorf("%w: jenis referensi resume tidak valid", ErrInputTidakValid)
	}
}

func (r *Repositori) referensiKeluhan(ctx context.Context, noRawat, pola string) ([]ReferensiResume, error) {
	ralan, err := r.bacaReferensi(ctx, `
		SELECT tgl_perawatan, jam_rawat, keluhan, 'Pemeriksaan Ralan'
		FROM pemeriksaan_ralan
		WHERE no_rawat=? AND (tgl_perawatan LIKE ? OR keluhan LIKE ?) AND keluhan<>''
		ORDER BY tgl_perawatan, jam_rawat
	`, noRawat, pola, pola)
	if err != nil {
		return nil, err
	}
	ranap, err := r.bacaReferensi(ctx, `
		SELECT tgl_perawatan, jam_rawat, keluhan, 'Pemeriksaan Ranap'
		FROM pemeriksaan_ranap
		WHERE no_rawat=? AND (tgl_perawatan LIKE ? OR keluhan LIKE ?) AND keluhan<>''
		ORDER BY tgl_perawatan, jam_rawat
	`, noRawat, pola, pola)
	if err != nil {
		return nil, err
	}
	return append(ralan, ranap...), nil
}

func (r *Repositori) referensiPemeriksaan(ctx context.Context, noRawat, pola string) ([]ReferensiResume, error) {
	ralan, err := r.bacaReferensi(ctx, `
		SELECT tgl_perawatan, jam_rawat, pemeriksaan, 'Pemeriksaan Ralan'
		FROM pemeriksaan_ralan
		WHERE no_rawat=? AND (tgl_perawatan LIKE ? OR pemeriksaan LIKE ?) AND pemeriksaan<>''
		ORDER BY tgl_perawatan, jam_rawat
	`, noRawat, pola, pola)
	if err != nil {
		return nil, err
	}
	ranap, err := r.bacaReferensi(ctx, `
		SELECT tgl_perawatan, jam_rawat, pemeriksaan, 'Pemeriksaan Ranap'
		FROM pemeriksaan_ranap
		WHERE no_rawat=? AND (tgl_perawatan LIKE ? OR pemeriksaan LIKE ?) AND pemeriksaan<>''
		ORDER BY tgl_perawatan, jam_rawat
	`, noRawat, pola, pola)
	if err != nil {
		return nil, err
	}
	return append(ralan, ranap...), nil
}

func (r *Repositori) referensiRadiologi(ctx context.Context, noRawat, pola string) ([]ReferensiResume, error) {
	return r.bacaReferensi(ctx, `
		SELECT tgl_periksa, jam, hasil, 'Radiologi'
		FROM hasil_radiologi
		WHERE no_rawat=? AND (tgl_periksa LIKE ? OR hasil LIKE ?) AND hasil<>''
		ORDER BY tgl_periksa, jam
	`, noRawat, pola, pola)
}

func (r *Repositori) referensiLaboratorium(ctx context.Context, noRawat, pola string) ([]ReferensiResume, error) {
	return r.bacaReferensi(ctx, `
		SELECT detail_periksa_lab.tgl_periksa, detail_periksa_lab.jam,
			CONCAT(COALESCE(template_laboratorium.Pemeriksaan, ''), ' : ', COALESCE(detail_periksa_lab.nilai, '')), 'Laboratorium'
		FROM detail_periksa_lab
		INNER JOIN template_laboratorium ON detail_periksa_lab.id_template=template_laboratorium.id_template
		WHERE detail_periksa_lab.no_rawat=?
			AND (detail_periksa_lab.tgl_periksa LIKE ? OR template_laboratorium.Pemeriksaan LIKE ?)
		ORDER BY detail_periksa_lab.tgl_periksa, detail_periksa_lab.jam
	`, noRawat, pola, pola)
}

func (r *Repositori) referensiObat(ctx context.Context, noRawat, pola string) ([]ReferensiResume, error) {
	return r.bacaReferensi(ctx, `
		SELECT detail_pemberian_obat.tgl_perawatan, detail_pemberian_obat.jam,
			CONCAT(COALESCE(databarang.nama_brng, ''), ' : ', COALESCE(detail_pemberian_obat.jml, ''), ' ', COALESCE(databarang.kode_sat, '')), 'Obat'
		FROM detail_pemberian_obat
		INNER JOIN databarang ON detail_pemberian_obat.kode_brng=databarang.kode_brng
		WHERE detail_pemberian_obat.no_rawat=?
			AND (detail_pemberian_obat.tgl_perawatan LIKE ? OR databarang.nama_brng LIKE ?)
		ORDER BY detail_pemberian_obat.tgl_perawatan, detail_pemberian_obat.jam
	`, noRawat, pola, pola)
}

func (r *Repositori) bacaReferensi(ctx context.Context, query string, args ...any) ([]ReferensiResume, error) {
	rows, err := r.simrsDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("baca referensi resume pasien: %w", err)
	}
	defer rows.Close()
	daftar := make([]ReferensiResume, 0, 20)
	for rows.Next() {
		var item ReferensiResume
		if err := rows.Scan(&item.Tanggal, &item.Jam, &item.Isi, &item.Sumber); err != nil {
			return nil, err
		}
		daftar = append(daftar, item)
	}
	return daftar, rows.Err()
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (r *Repositori) billingTerkunci(ctx context.Context, q queryer, noRawat string) (bool, error) {
	var jumlah int
	err := q.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM billing WHERE no_rawat=?) + (SELECT COUNT(*) FROM reg_periksa WHERE no_rawat=? AND stts='Batal')`, noRawat, noRawat).Scan(&jumlah)
	return jumlah > 0, err
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

func prefiksKolom(alias string) string {
	hasil := make([]string, len(kolom))
	for i, nama := range kolom {
		hasil[i] = `COALESCE(` + alias + `.` + nama + `,'')`
	}
	return strings.Join(hasil, ",")
}

func placeholder(jumlah int) string { return strings.TrimSuffix(strings.Repeat("?,", jumlah), ",") }

func nilai(i Input) []any {
	return []any{i.NoRawat, i.KodeDokter, i.KondisiPulang, i.KeluhanUtama, i.JalannyaPenyakit,
		i.PemeriksaanPenunjang, i.HasilLaborat, i.DiagnosaUtama, i.KodeDiagnosaUtama,
		i.DiagnosaSekunder, i.KodeDiagnosaSekunder, i.DiagnosaSekunder2, i.KodeDiagnosaSekunder2,
		i.DiagnosaSekunder3, i.KodeDiagnosaSekunder3, i.DiagnosaSekunder4, i.KodeDiagnosaSekunder4,
		i.ProsedurUtama, i.KodeProsedurUtama, i.ProsedurSekunder, i.KodeProsedurSekunder,
		i.ProsedurSekunder2, i.KodeProsedurSekunder2, i.ProsedurSekunder3, i.KodeProsedurSekunder3,
		i.ObatPulang}
}

type pemindai interface{ Scan(...any) error }

func scanResume(row pemindai, i *Input, dokter *Dokter) error {
	return row.Scan(&i.NoRawat, &i.KodeDokter, &i.KondisiPulang, &i.KeluhanUtama, &i.JalannyaPenyakit,
		&i.PemeriksaanPenunjang, &i.HasilLaborat, &i.DiagnosaUtama, &i.KodeDiagnosaUtama,
		&i.DiagnosaSekunder, &i.KodeDiagnosaSekunder, &i.DiagnosaSekunder2, &i.KodeDiagnosaSekunder2,
		&i.DiagnosaSekunder3, &i.KodeDiagnosaSekunder3, &i.DiagnosaSekunder4, &i.KodeDiagnosaSekunder4,
		&i.ProsedurUtama, &i.KodeProsedurUtama, &i.ProsedurSekunder, &i.KodeProsedurSekunder,
		&i.ProsedurSekunder2, &i.KodeProsedurSekunder2, &i.ProsedurSekunder3, &i.KodeProsedurSekunder3,
		&i.ObatPulang, &dokter.Nama, &dokter.Spesialis)
}
