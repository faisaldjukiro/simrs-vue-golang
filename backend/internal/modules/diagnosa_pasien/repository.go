package diagnosa_pasien

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInputTidakValid  = errors.New("input diagnosa pasien tidak valid")
	ErrBillingTerkunci  = errors.New("billing sudah terverifikasi atau kunjungan dibatalkan")
	ErrKunjunganInvalid = errors.New("kunjungan pasien tidak ditemukan")
	ErrCodingDuplikat   = errors.New("diagnosa atau prosedur pasien sudah tersedia")
	ErrCodingTidakAda   = errors.New("diagnosa atau prosedur pasien tidak ditemukan")
)

type Coding struct {
	Kode           string `json:"kode"`
	Nama           string `json:"nama"`
	Prioritas      int    `json:"prioritas"`
	StatusPenyakit string `json:"status_penyakit,omitempty"`
	Valid          bool   `json:"valid"`
	BolehUtama     bool   `json:"boleh_utama"`
	ValidCode      string `json:"validcode"`
	AccPDX         string `json:"accpdx"`
	Asterisk       string `json:"asterisk,omitempty"`
	IndonesiaM     string `json:"im"`
	Penanda        string `json:"penanda,omitempty"`
	Pesan          string `json:"pesan"`
	Jumlah         int    `json:"jumlah,omitempty"`
}

type Input struct {
	NoRawat  string   `json:"no_rawat"`
	Status   string   `json:"status"`
	Diagnosa []Coding `json:"diagnosa"`
	Prosedur []Coding `json:"prosedur"`
}

type Data struct {
	NoRawat         string   `json:"no_rawat"`
	Status          string   `json:"status"`
	BillingTerkunci bool     `json:"billing_terkunci"`
	Diagnosa        []Coding `json:"diagnosa"`
	Prosedur        []Coding `json:"prosedur"`
}

type HapusInput struct {
	NoRawat string
	Status  string
	Jenis   string
	Kode    string
}

type Repositori struct{ simrsDB *sql.DB }

func NewRepositori(simrsDB *sql.DB) *Repositori { return &Repositori{simrsDB: simrsDB} }

func (r *Repositori) Data(ctx context.Context, noRawat, status string) (Data, error) {
	data := Data{
		NoRawat:  noRawat,
		Status:   status,
		Diagnosa: make([]Coding, 0),
		Prosedur: make([]Coding, 0),
	}
	var ada int
	if err := r.simrsDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM reg_periksa WHERE no_rawat=?`, noRawat).Scan(&ada); err != nil {
		return Data{}, fmt.Errorf("periksa kunjungan pasien: %w", err)
	}
	if ada == 0 {
		return Data{}, ErrKunjunganInvalid
	}
	var err error
	data.BillingTerkunci, err = r.billingTerkunci(ctx, r.simrsDB, noRawat)
	if err != nil {
		return Data{}, fmt.Errorf("periksa status billing pasien: %w", err)
	}
	data.Diagnosa, err = r.ambilDiagnosa(ctx, noRawat, status)
	if err != nil {
		return Data{}, err
	}
	data.Prosedur, err = r.ambilProsedur(ctx, noRawat, status)
	return data, err
}

func (r *Repositori) ambilDiagnosa(ctx context.Context, noRawat, status string) ([]Coding, error) {
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT dp.kd_penyakit, COALESCE(p.nm_penyakit,''), dp.prioritas,
			COALESCE(dp.status_penyakit,''), COALESCE(p.validcode,'1'),
			COALESCE(p.accpdx,'Y'), COALESCE(p.asterisk,'0'), COALESCE(p.im,'0')
		FROM diagnosa_pasien dp
		LEFT JOIN penyakit p ON p.kd_penyakit=dp.kd_penyakit
		WHERE dp.no_rawat=? AND dp.status=?
		ORDER BY dp.prioritas, dp.kd_penyakit
	`, noRawat, status)
	if err != nil {
		return nil, fmt.Errorf("baca diagnosa pasien: %w", err)
	}
	defer rows.Close()
	hasil := make([]Coding, 0)
	for rows.Next() {
		item := Coding{}
		if err := rows.Scan(&item.Kode, &item.Nama, &item.Prioritas, &item.StatusPenyakit, &item.ValidCode, &item.AccPDX, &item.Asterisk, &item.IndonesiaM); err != nil {
			return nil, fmt.Errorf("baca baris diagnosa pasien: %w", err)
		}
		lengkapiCoding(&item, item.Prioritas == 1)
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *Repositori) ambilProsedur(ctx context.Context, noRawat, status string) ([]Coding, error) {
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT pp.kode, COALESCE(i.deskripsi_panjang,''), pp.prioritas,
			COALESCE(i.validcode,'1'), COALESCE(i.accpdx,'Y'), COALESCE(i.im,'0'),
			CAST(COALESCE(NULLIF(pp.jumlah,''), '1') AS UNSIGNED)
		FROM prosedur_pasien pp
		LEFT JOIN icd9 i ON i.kode=pp.kode
		WHERE pp.no_rawat=? AND pp.status=?
		ORDER BY pp.prioritas, pp.kode
	`, noRawat, status)
	if err != nil {
		return nil, fmt.Errorf("baca prosedur pasien: %w", err)
	}
	defer rows.Close()
	hasil := make([]Coding, 0)
	for rows.Next() {
		item := Coding{}
		if err := rows.Scan(&item.Kode, &item.Nama, &item.Prioritas, &item.ValidCode, &item.AccPDX, &item.IndonesiaM, &item.Jumlah); err != nil {
			return nil, fmt.Errorf("baca baris prosedur pasien: %w", err)
		}
		lengkapiCoding(&item, item.Prioritas == 1)
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *Repositori) CariCoding(ctx context.Context, jenis, kataKunci string, utama bool) ([]Coding, error) {
	pola := "%" + strings.TrimSpace(kataKunci) + "%"
	tepat := normalisasiKode(kataKunci)
	var rows *sql.Rows
	var err error
	if jenis == "diagnosa" {
		rows, err = r.simrsDB.QueryContext(ctx, `
			SELECT p.kd_penyakit, COALESCE(p.nm_penyakit,''), COALESCE(p.validcode,'1'),
				COALESCE(p.accpdx,'Y'), COALESCE(p.asterisk,'0'), COALESCE(p.im,'0')
			FROM penyakit p
			LEFT JOIN kategori_penyakit kp ON kp.kd_ktg=p.kd_ktg
			WHERE p.kd_penyakit LIKE ? OR p.nm_penyakit LIKE ? OR p.ciri_ciri LIKE ?
				OR p.keterangan LIKE ? OR kp.nm_kategori LIKE ? OR kp.ciri_umum LIKE ?
			ORDER BY p.kd_penyakit=? DESC, p.validcode DESC, p.kd_penyakit LIMIT 20
		`, pola, pola, pola, pola, pola, pola, tepat)
	} else {
		rows, err = r.simrsDB.QueryContext(ctx, `
			SELECT kode, COALESCE(deskripsi_panjang,''), COALESCE(validcode,'1'),
				COALESCE(accpdx,'Y'), COALESCE(im,'0')
			FROM icd9 WHERE kode LIKE ? OR deskripsi_panjang LIKE ? OR deskripsi_pendek LIKE ?
			ORDER BY kode=? DESC, validcode DESC, kode LIMIT 20
		`, pola, pola, pola, tepat)
	}
	if err != nil {
		return nil, fmt.Errorf("cari master %s: %w", jenis, err)
	}
	defer rows.Close()
	hasil := make([]Coding, 0, 20)
	for rows.Next() {
		item := Coding{}
		if jenis == "diagnosa" {
			err = rows.Scan(&item.Kode, &item.Nama, &item.ValidCode, &item.AccPDX, &item.Asterisk, &item.IndonesiaM)
		} else {
			err = rows.Scan(&item.Kode, &item.Nama, &item.ValidCode, &item.AccPDX, &item.IndonesiaM)
			item.Jumlah = 1
		}
		if err != nil {
			return nil, fmt.Errorf("baca hasil pencarian %s: %w", jenis, err)
		}
		lengkapiCoding(&item, utama)
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *Repositori) Simpan(ctx context.Context, input Input) error {
	tx, err := r.simrsDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mulai transaksi diagnosa pasien: %w", err)
	}
	defer tx.Rollback()

	terkunci, err := r.billingTerkunci(ctx, tx, input.NoRawat)
	if err != nil {
		return fmt.Errorf("periksa status billing pasien: %w", err)
	}
	if terkunci {
		return ErrBillingTerkunci
	}

	prioritasDiagnosa, err := r.prioritasTerakhir(ctx, tx, "diagnosa_pasien", input.NoRawat, input.Status)
	if err != nil {
		return err
	}
	prioritasProsedur, err := r.prioritasTerakhir(ctx, tx, "prosedur_pasien", input.NoRawat, input.Status)
	if err != nil {
		return err
	}
	diagnosa, err := r.validasiDaftar(ctx, tx, "diagnosa", input.Diagnosa, prioritasDiagnosa)
	if err != nil {
		return err
	}
	prosedur, err := r.validasiDaftar(ctx, tx, "prosedur", input.Prosedur, prioritasProsedur)
	if err != nil {
		return err
	}
	if len(diagnosa) == 0 && len(prosedur) == 0 {
		return inputTidakValid("pilih minimal satu diagnosa atau prosedur")
	}

	for _, item := range diagnosa {
		ada, err := r.codingSudahAda(ctx, tx, "diagnosa_pasien", "kd_penyakit", input.NoRawat, input.Status, item.Kode)
		if err != nil {
			return err
		}
		if ada {
			return fmt.Errorf("%w: diagnosa %s sudah tersimpan", ErrCodingDuplikat, item.Kode)
		}
		statusPenyakit, err := r.statusPenyakit(ctx, tx, input.NoRawat, item.Kode)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO diagnosa_pasien (no_rawat,kd_penyakit,status,prioritas,status_penyakit)
			VALUES (?,?,?,?,?)
		`, input.NoRawat, item.Kode, input.Status, item.Prioritas, statusPenyakit); err != nil {
			return fmt.Errorf("simpan diagnosa %s: %w", item.Kode, err)
		}
	}
	for _, item := range prosedur {
		ada, err := r.codingSudahAda(ctx, tx, "prosedur_pasien", "kode", input.NoRawat, input.Status, item.Kode)
		if err != nil {
			return err
		}
		if ada {
			return fmt.Errorf("%w: prosedur %s sudah tersimpan", ErrCodingDuplikat, item.Kode)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO prosedur_pasien (no_rawat,kode,status,prioritas,jumlah)
			VALUES (?,?,?,?,?)
		`, input.NoRawat, item.Kode, input.Status, item.Prioritas, item.Jumlah); err != nil {
			return fmt.Errorf("simpan prosedur %s: %w", item.Kode, err)
		}
	}
	if err := r.sinkronkanResume(ctx, tx, input.NoRawat, input.Status); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("simpan transaksi diagnosa pasien: %w", err)
	}
	return nil
}

func (r *Repositori) validasiDaftar(ctx context.Context, tx *sql.Tx, jenis string, daftar []Coding, prioritasAwal int) ([]Coding, error) {
	hasil := make([]Coding, 0, len(daftar))
	terpakai := make(map[string]struct{}, len(daftar))
	for _, kiriman := range daftar {
		kode := normalisasiKode(kiriman.Kode)
		if kode == "" {
			continue
		}
		if _, ada := terpakai[kode]; ada {
			return nil, inputTidakValid("kode %s %s tidak boleh duplikat", jenis, kode)
		}
		item := Coding{Kode: kode, Prioritas: prioritasAwal + len(hasil) + 1, Jumlah: kiriman.Jumlah}
		if jenis == "diagnosa" {
			err := tx.QueryRowContext(ctx, `
				SELECT COALESCE(nm_penyakit,''), COALESCE(validcode,'1'),
					COALESCE(accpdx,'Y'), COALESCE(asterisk,'0'), COALESCE(im,'0')
				FROM penyakit WHERE kd_penyakit=? LIMIT 1
			`, kode).Scan(&item.Nama, &item.ValidCode, &item.AccPDX, &item.Asterisk, &item.IndonesiaM)
			if errors.Is(err, sql.ErrNoRows) {
				return nil, inputTidakValid("kode diagnosa %s tidak ditemukan pada master ICD-10", kode)
			}
			if err != nil {
				return nil, fmt.Errorf("validasi diagnosa %s: %w", kode, err)
			}
		} else {
			if item.Jumlah == 0 {
				item.Jumlah = 1
			}
			if item.Jumlah < 1 || item.Jumlah > 999 {
				return nil, inputTidakValid("jumlah prosedur %s harus antara 1 sampai 999", kode)
			}
			err := tx.QueryRowContext(ctx, `
				SELECT COALESCE(deskripsi_panjang,''), COALESCE(validcode,'1'),
					COALESCE(accpdx,'Y'), COALESCE(im,'0')
				FROM icd9 WHERE kode=? LIMIT 1
			`, kode).Scan(&item.Nama, &item.ValidCode, &item.AccPDX, &item.IndonesiaM)
			if errors.Is(err, sql.ErrNoRows) {
				return nil, inputTidakValid("kode prosedur %s tidak ditemukan pada master ICD-9-CM", kode)
			}
			if err != nil {
				return nil, fmt.Errorf("validasi prosedur %s: %w", kode, err)
			}
		}
		lengkapiCoding(&item, item.Prioritas == 1)
		if !item.Valid {
			return nil, inputTidakValid("%s", item.Pesan)
		}
		terpakai[kode] = struct{}{}
		hasil = append(hasil, item)
	}
	return hasil, nil
}

func (r *Repositori) Hapus(ctx context.Context, input HapusInput) error {
	tx, err := r.simrsDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mulai transaksi hapus coding pasien: %w", err)
	}
	defer tx.Rollback()
	terkunci, err := r.billingTerkunci(ctx, tx, input.NoRawat)
	if err != nil {
		return fmt.Errorf("periksa status billing pasien: %w", err)
	}
	if terkunci {
		return ErrBillingTerkunci
	}
	tabel, kolom := "diagnosa_pasien", "kd_penyakit"
	if input.Jenis == "prosedur" {
		tabel, kolom = "prosedur_pasien", "kode"
	}
	hasil, err := tx.ExecContext(ctx, `DELETE FROM `+tabel+` WHERE no_rawat=? AND status=? AND `+kolom+`=?`, input.NoRawat, input.Status, input.Kode)
	if err != nil {
		return fmt.Errorf("hapus %s pasien %s: %w", input.Jenis, input.Kode, err)
	}
	jumlah, err := hasil.RowsAffected()
	if err != nil {
		return err
	}
	if jumlah == 0 {
		return fmt.Errorf("%w: %s %s", ErrCodingTidakAda, input.Jenis, input.Kode)
	}
	if err := r.rapikanPrioritas(ctx, tx, tabel, kolom, input.NoRawat, input.Status); err != nil {
		return err
	}
	if err := r.sinkronkanResume(ctx, tx, input.NoRawat, input.Status); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("simpan penghapusan coding pasien: %w", err)
	}
	return nil
}

func (r *Repositori) prioritasTerakhir(ctx context.Context, tx *sql.Tx, tabel, noRawat, status string) (int, error) {
	var prioritas int
	err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(prioritas),0) FROM `+tabel+` WHERE no_rawat=? AND status=?`, noRawat, status).Scan(&prioritas)
	if err != nil {
		return 0, fmt.Errorf("baca prioritas %s: %w", tabel, err)
	}
	return prioritas, nil
}

func (r *Repositori) codingSudahAda(ctx context.Context, tx *sql.Tx, tabel, kolom, noRawat, status, kode string) (bool, error) {
	var jumlah int
	err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+tabel+` WHERE no_rawat=? AND status=? AND `+kolom+`=?`, noRawat, status, kode).Scan(&jumlah)
	if err != nil {
		return false, fmt.Errorf("periksa duplikat %s %s: %w", tabel, kode, err)
	}
	return jumlah > 0, nil
}

func (r *Repositori) rapikanPrioritas(ctx context.Context, tx *sql.Tx, tabel, kolom, noRawat, status string) error {
	rows, err := tx.QueryContext(ctx, `SELECT `+kolom+` FROM `+tabel+` WHERE no_rawat=? AND status=? ORDER BY prioritas, `+kolom, noRawat, status)
	if err != nil {
		return fmt.Errorf("baca urutan %s: %w", tabel, err)
	}
	var kode []string
	for rows.Next() {
		var item string
		if err := rows.Scan(&item); err != nil {
			rows.Close()
			return err
		}
		kode = append(kode, item)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for i, item := range kode {
		if _, err := tx.ExecContext(ctx, `UPDATE `+tabel+` SET prioritas=? WHERE no_rawat=? AND status=? AND `+kolom+`=?`, i+1, noRawat, status, item); err != nil {
			return fmt.Errorf("rapikan urutan %s: %w", tabel, err)
		}
	}
	return nil
}

func (r *Repositori) sinkronkanResume(ctx context.Context, tx *sql.Tx, noRawat, status string) error {
	diagnosa, err := r.daftarKode(ctx, tx, "diagnosa_pasien", "kd_penyakit", noRawat, status, 5)
	if err != nil {
		return err
	}
	prosedur, err := r.daftarKode(ctx, tx, "prosedur_pasien", "kode", noRawat, status, 4)
	if err != nil {
		return err
	}
	for len(diagnosa) < 5 {
		diagnosa = append(diagnosa, "")
	}
	for len(prosedur) < 4 {
		prosedur = append(prosedur, "")
	}
	tabelResume := "resume_pasien"
	if status == "Ranap" {
		tabelResume = "resume_pasien_ranap"
	}
	_, err = tx.ExecContext(ctx, `UPDATE `+tabelResume+` SET
		kd_diagnosa_utama=?, kd_diagnosa_sekunder=?, kd_diagnosa_sekunder2=?,
		kd_diagnosa_sekunder3=?, kd_diagnosa_sekunder4=?, kd_prosedur_utama=?,
		kd_prosedur_sekunder=?, kd_prosedur_sekunder2=?, kd_prosedur_sekunder3=?
		WHERE no_rawat=?`, diagnosa[0], diagnosa[1], diagnosa[2], diagnosa[3], diagnosa[4],
		prosedur[0], prosedur[1], prosedur[2], prosedur[3], noRawat)
	if err != nil {
		return fmt.Errorf("sinkronkan coding ke %s: %w", tabelResume, err)
	}
	return nil
}

func (r *Repositori) daftarKode(ctx context.Context, tx *sql.Tx, tabel, kolom, noRawat, status string, batas int) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT `+kolom+` FROM `+tabel+` WHERE no_rawat=? AND status=? ORDER BY prioritas, `+kolom+` LIMIT ?`, noRawat, status, batas)
	if err != nil {
		return nil, fmt.Errorf("baca coding untuk resume: %w", err)
	}
	defer rows.Close()
	hasil := make([]string, 0, batas)
	for rows.Next() {
		var kode string
		if err := rows.Scan(&kode); err != nil {
			return nil, err
		}
		hasil = append(hasil, kode)
	}
	return hasil, rows.Err()
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (r *Repositori) billingTerkunci(ctx context.Context, q queryer, noRawat string) (bool, error) {
	var jumlah int
	err := q.QueryRowContext(ctx, `
		SELECT (SELECT COUNT(*) FROM billing WHERE no_rawat=?) +
			(SELECT COUNT(*) FROM reg_periksa WHERE no_rawat=? AND stts='Batal')
	`, noRawat, noRawat).Scan(&jumlah)
	return jumlah > 0, err
}

func (r *Repositori) statusPenyakit(ctx context.Context, tx *sql.Tx, noRawat, kode string) (string, error) {
	var pernah int
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM diagnosa_pasien dp
		INNER JOIN reg_periksa lama ON lama.no_rawat=dp.no_rawat
		INNER JOIN reg_periksa sekarang ON sekarang.no_rawat=?
		WHERE lama.no_rkm_medis=sekarang.no_rkm_medis
			AND dp.kd_penyakit=?
	`, noRawat, kode).Scan(&pernah)
	if err != nil {
		return "", fmt.Errorf("periksa riwayat diagnosa %s: %w", kode, err)
	}
	if pernah > 0 {
		return "Lama", nil
	}
	return "Baru", nil
}

func lengkapiCoding(item *Coding, utama bool) {
	item.Kode = normalisasiKode(item.Kode)
	item.Nama = strings.TrimSpace(item.Nama)
	item.ValidCode = strings.TrimSpace(item.ValidCode)
	item.AccPDX = strings.ToUpper(strings.TrimSpace(item.AccPDX))
	item.Asterisk = strings.TrimSpace(item.Asterisk)
	item.IndonesiaM = strings.TrimSpace(item.IndonesiaM)
	item.BolehUtama = item.AccPDX != "N"
	item.Valid = false
	if item.ValidCode == "0" {
		item.Penanda = "Header"
		item.Pesan = fmt.Sprintf("Kode %s hanya berfungsi sebagai header (validcode 0)", item.Kode)
		return
	}
	if utama && !item.BolehUtama {
		item.Penanda = "Tidak bisa utama"
		item.Pesan = fmt.Sprintf("Kode %s tidak dapat menjadi coding utama (accpdx N)", item.Kode)
		return
	}
	item.Valid = true
	item.Pesan = "Kode valid"
	if item.IndonesiaM == "1" {
		item.Penanda = "IM"
		item.Pesan = "Kode valid untuk IDRG (Indonesia Modification)"
	} else if !item.BolehUtama {
		item.Penanda = "Sekunder"
	}
}

func normalisasiKode(kode string) string { return strings.ToUpper(strings.TrimSpace(kode)) }

func inputTidakValid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInputTidakValid, fmt.Sprintf(format, args...))
}
