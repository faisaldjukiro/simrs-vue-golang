package resume_pasien_ranap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrTidakDitemukan  = errors.New("resume pasien rawat inap tidak ditemukan")
	ErrSudahAda        = errors.New("resume pasien rawat inap sudah tersedia")
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
	DiagnosaAwal          string `json:"diagnosa_awal"`
	Alasan                string `json:"alasan"`
	KeluhanUtama          string `json:"keluhan_utama"`
	PemeriksaanFisik      string `json:"pemeriksaan_fisik"`
	JalannyaPenyakit      string `json:"jalannya_penyakit"`
	PemeriksaanPenunjang  string `json:"pemeriksaan_penunjang"`
	HasilLaborat          string `json:"hasil_laborat"`
	TindakanDanOperasi    string `json:"tindakan_dan_operasi"`
	ObatDiRS              string `json:"obat_di_rs"`
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
	Alergi                string `json:"alergi"`
	Diet                  string `json:"diet"`
	LabBelum              string `json:"lab_belum"`
	Edukasi               string `json:"edukasi"`
	CaraKeluar            string `json:"cara_keluar"`
	KeteranganKeluar      string `json:"ket_keluar"`
	Keadaan               string `json:"keadaan"`
	KeteranganKeadaan     string `json:"ket_keadaan"`
	Dilanjutkan           string `json:"dilanjutkan"`
	KeteranganDilanjutkan string `json:"ket_dilanjutkan"`
	Kontrol               string `json:"kontrol"`
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
	CaraKeluar  []string `json:"cara_keluar"`
	Keadaan     []string `json:"keadaan"`
	Dilanjutkan []string `json:"dilanjutkan"`
}

var kolom = []string{
	"no_rawat", "kd_dokter", "diagnosa_awal", "alasan", "keluhan_utama", "pemeriksaan_fisik",
	"jalannya_penyakit", "pemeriksaan_penunjang", "hasil_laborat", "tindakan_dan_operasi", "obat_di_rs",
	"diagnosa_utama", "kd_diagnosa_utama", "diagnosa_sekunder", "kd_diagnosa_sekunder",
	"diagnosa_sekunder2", "kd_diagnosa_sekunder2", "diagnosa_sekunder3", "kd_diagnosa_sekunder3",
	"diagnosa_sekunder4", "kd_diagnosa_sekunder4", "prosedur_utama", "kd_prosedur_utama",
	"prosedur_sekunder", "kd_prosedur_sekunder", "prosedur_sekunder2", "kd_prosedur_sekunder2",
	"prosedur_sekunder3", "kd_prosedur_sekunder3", "alergi", "diet", "lab_belum", "edukasi",
	"cara_keluar", "ket_keluar", "keadaan", "ket_keadaan", "dilanjutkan", "ket_dilanjutkan", "kontrol", "obat_pulang",
}

type Repositori struct{ simrsDB *sql.DB }

func NewRepositori(simrsDB *sql.DB) *Repositori { return &Repositori{simrsDB: simrsDB} }

func (r *Repositori) Data(ctx context.Context, noRawat string) (Data, error) {
	data := Data{
		Resume: Input{
			NoRawat: noRawat, Kontrol: time.Now().Format("2006-01-02 15:04:05"),
			CaraKeluar: "Atas Izin Dokter", Keadaan: "Membaik", Dilanjutkan: "Kembali Ke RS",
		},
		Pilihan: Pilihan{
			CaraKeluar:  []string{"Atas Izin Dokter", "Pindah RS", "Pulang Atas Permintaan Sendiri", "Lainnya"},
			Keadaan:     []string{"Membaik", "Sembuh", "Keadaan Khusus", "Meninggal"},
			Dilanjutkan: []string{"Kembali Ke RS", "RS Lain", "Dokter Luar", "Puskesmes", "Lainnya"},
		},
	}
	terkunci, err := r.billingTerkunci(ctx, r.simrsDB, noRawat)
	if err != nil {
		return Data{}, fmt.Errorf("periksa status billing pasien: %w", err)
	}
	data.BillingTerkunci = terkunci

	resume, dokter, err := r.ambil(ctx, noRawat)
	if err == nil {
		// Coding diagnosa dan prosedur mengikuti sumber utama SIMRS Khanza.
		// Resume lama bisa saja tersimpan sebelum coding selesai sehingga nama
		// atau diagnosa/prosedur sekundernya masih kosong.
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
	query := `SELECT ` + prefiksKolom("r") + `,
		COALESCE(d.nm_dokter,''), COALESCE(s.nm_sps,'')
		FROM resume_pasien_ranap r
		LEFT JOIN dokter d ON d.kd_dokter=r.kd_dokter
		LEFT JOIN spesialis s ON s.kd_sps=d.kd_sps
		WHERE r.no_rawat=? LIMIT 1`
	row := r.simrsDB.QueryRowContext(ctx, query, noRawat)
	err := scanResume(row, &item, &dokter)
	dokter.Kode = item.KodeDokter
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Input{}, Dokter{}, fmt.Errorf("baca resume pasien rawat inap: %w", err)
	}
	return item, dokter, err
}

func (r *Repositori) isiDefault(ctx context.Context, item *Input, dokter *Dokter) error {
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT COALESCE(ki.diagnosa_awal,'')
		FROM kamar_inap ki WHERE ki.no_rawat=?
		ORDER BY ki.tgl_masuk DESC, ki.jam_masuk DESC LIMIT 1
	`, item.NoRawat).Scan(&item.DiagnosaAwal)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("baca diagnosa awal pasien: %w", err)
	}

	err = r.simrsDB.QueryRowContext(ctx, `
		SELECT d.kd_dokter, COALESCE(d.nm_dokter,''), COALESCE(s.nm_sps,'')
		FROM dokter d
		LEFT JOIN spesialis s ON s.kd_sps=d.kd_sps
		WHERE d.kd_dokter=COALESCE(
			(SELECT dr.kd_dokter FROM dpjp_ranap dr WHERE dr.no_rawat=? ORDER BY dr.urutan LIMIT 1),
			(SELECT rp.kd_dokter FROM reg_periksa rp WHERE rp.no_rawat=? LIMIT 1)
		) LIMIT 1
	`, item.NoRawat, item.NoRawat).Scan(&dokter.Kode, &dokter.Nama, &dokter.Spesialis)
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
		query := `INSERT INTO resume_pasien_ranap (` + strings.Join(kolom, ",") + `) VALUES (` + placeholder(len(args)) + `)`
		_, err := tx.ExecContext(ctx, query, args...)
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrSudahAda
		}
		if err != nil {
			return fmt.Errorf("simpan resume pasien rawat inap: %w", err)
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
		hasil, err := tx.ExecContext(ctx, `UPDATE resume_pasien_ranap SET `+strings.Join(set, ",")+` WHERE no_rawat=?`, args...)
		if err != nil {
			return fmt.Errorf("ubah resume pasien rawat inap: %w", err)
		}
		jumlah, err := hasil.RowsAffected()
		if err != nil {
			return err
		}
		if jumlah == 0 {
			var ada int
			if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM resume_pasien_ranap WHERE no_rawat=?`, input.NoRawat).Scan(&ada); err != nil {
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
		hasil, err := tx.ExecContext(ctx, `DELETE FROM resume_pasien_ranap WHERE no_rawat=?`, noRawat)
		if err != nil {
			return fmt.Errorf("hapus resume pasien rawat inap: %w", err)
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
		if nama == "kontrol" {
			hasil[i] = `COALESCE(DATE_FORMAT(` + alias + `.kontrol,'%Y-%m-%d %H:%i:%s'),'')`
		} else {
			hasil[i] = `COALESCE(` + alias + `.` + nama + `,'')`
		}
	}
	return strings.Join(hasil, ",")
}

func placeholder(jumlah int) string { return strings.TrimSuffix(strings.Repeat("?,", jumlah), ",") }

func nilai(i Input) []any {
	return []any{i.NoRawat, i.KodeDokter, i.DiagnosaAwal, i.Alasan, i.KeluhanUtama, i.PemeriksaanFisik,
		i.JalannyaPenyakit, i.PemeriksaanPenunjang, i.HasilLaborat, i.TindakanDanOperasi, i.ObatDiRS,
		i.DiagnosaUtama, i.KodeDiagnosaUtama, i.DiagnosaSekunder, i.KodeDiagnosaSekunder,
		i.DiagnosaSekunder2, i.KodeDiagnosaSekunder2, i.DiagnosaSekunder3, i.KodeDiagnosaSekunder3,
		i.DiagnosaSekunder4, i.KodeDiagnosaSekunder4, i.ProsedurUtama, i.KodeProsedurUtama,
		i.ProsedurSekunder, i.KodeProsedurSekunder, i.ProsedurSekunder2, i.KodeProsedurSekunder2,
		i.ProsedurSekunder3, i.KodeProsedurSekunder3, i.Alergi, i.Diet, i.LabBelum, i.Edukasi,
		i.CaraKeluar, i.KeteranganKeluar, i.Keadaan, i.KeteranganKeadaan, i.Dilanjutkan,
		i.KeteranganDilanjutkan, i.Kontrol, i.ObatPulang}
}

type pemindai interface{ Scan(...any) error }

func scanResume(row pemindai, i *Input, dokter *Dokter) error {
	return row.Scan(&i.NoRawat, &i.KodeDokter, &i.DiagnosaAwal, &i.Alasan, &i.KeluhanUtama, &i.PemeriksaanFisik,
		&i.JalannyaPenyakit, &i.PemeriksaanPenunjang, &i.HasilLaborat, &i.TindakanDanOperasi, &i.ObatDiRS,
		&i.DiagnosaUtama, &i.KodeDiagnosaUtama, &i.DiagnosaSekunder, &i.KodeDiagnosaSekunder,
		&i.DiagnosaSekunder2, &i.KodeDiagnosaSekunder2, &i.DiagnosaSekunder3, &i.KodeDiagnosaSekunder3,
		&i.DiagnosaSekunder4, &i.KodeDiagnosaSekunder4, &i.ProsedurUtama, &i.KodeProsedurUtama,
		&i.ProsedurSekunder, &i.KodeProsedurSekunder, &i.ProsedurSekunder2, &i.KodeProsedurSekunder2,
		&i.ProsedurSekunder3, &i.KodeProsedurSekunder3, &i.Alergi, &i.Diet, &i.LabBelum, &i.Edukasi,
		&i.CaraKeluar, &i.KeteranganKeluar, &i.Keadaan, &i.KeteranganKeadaan, &i.Dilanjutkan,
		&i.KeteranganDilanjutkan, &i.Kontrol, &i.ObatPulang, &dokter.Nama, &dokter.Spesialis)
}
