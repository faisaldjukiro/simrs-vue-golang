package resume_pasien

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type pilihanCoding struct {
	kode       *string
	nama       *string
	jenis      string
	prioritas  int
	validCode  string
	accPDX     string
	asterisk   string
	indonesiaM string
}

type HasilValidasiCoding struct {
	Jenis      string `json:"jenis"`
	Kode       string `json:"kode"`
	Nama       string `json:"nama"`
	Valid      bool   `json:"valid"`
	BolehUtama bool   `json:"boleh_utama"`
	ValidCode  string `json:"validcode"`
	AccPDX     string `json:"accpdx"`
	Asterisk   string `json:"asterisk,omitempty"`
	IndonesiaM string `json:"im"`
	Penanda    string `json:"penanda,omitempty"`
	Pesan      string `json:"pesan"`
}

func (r *Repositori) ValidasiCoding(ctx context.Context, jenis, kode string, utama bool) (HasilValidasiCoding, error) {
	hasil := HasilValidasiCoding{Jenis: jenis, Kode: normalisasiKode(kode), BolehUtama: true}
	var err error
	if jenis == "diagnosa" {
		err = r.simrsDB.QueryRowContext(ctx, `
			SELECT COALESCE(nm_penyakit,''), COALESCE(validcode,'1'),
				COALESCE(accpdx,'Y'), COALESCE(asterisk,'0'), COALESCE(im,'0')
			FROM penyakit WHERE kd_penyakit=? LIMIT 1
		`, hasil.Kode).Scan(&hasil.Nama, &hasil.ValidCode, &hasil.AccPDX, &hasil.Asterisk, &hasil.IndonesiaM)
	} else {
		err = r.simrsDB.QueryRowContext(ctx, `
			SELECT COALESCE(deskripsi_panjang,''), COALESCE(validcode,'1'),
				COALESCE(accpdx,'Y'), COALESCE(im,'0')
			FROM icd9 WHERE kode=? LIMIT 1
		`, hasil.Kode).Scan(&hasil.Nama, &hasil.ValidCode, &hasil.AccPDX, &hasil.IndonesiaM)
	}
	if errors.Is(err, sql.ErrNoRows) {
		hasil.Pesan = fmt.Sprintf("Kode %s %s tidak ditemukan pada master", jenis, hasil.Kode)
		return hasil, nil
	}
	if err != nil {
		return HasilValidasiCoding{}, fmt.Errorf("validasi kode %s %s: %w", jenis, hasil.Kode, err)
	}
	lengkapiHasilValidasi(&hasil, utama)
	return hasil, nil
}

func (r *Repositori) CariCoding(ctx context.Context, jenis, kataKunci string, utama bool) ([]HasilValidasiCoding, error) {
	pola := "%" + strings.TrimSpace(kataKunci) + "%"
	tepat := normalisasiKode(kataKunci)
	var (
		rows *sql.Rows
		err  error
	)
	if jenis == "diagnosa" {
		rows, err = r.simrsDB.QueryContext(ctx, `
			SELECT COALESCE(kd_penyakit,''), COALESCE(nm_penyakit,''),
				COALESCE(validcode,'1'), COALESCE(accpdx,'Y'),
				COALESCE(asterisk,'0'), COALESCE(im,'0')
			FROM penyakit
			WHERE kd_penyakit LIKE ? OR nm_penyakit LIKE ?
			ORDER BY kd_penyakit=? DESC, validcode DESC, kd_penyakit
			LIMIT 20
		`, pola, pola, tepat)
	} else {
		rows, err = r.simrsDB.QueryContext(ctx, `
			SELECT COALESCE(kode,''), COALESCE(deskripsi_panjang,''),
				COALESCE(validcode,'1'), COALESCE(accpdx,'Y'), COALESCE(im,'0')
			FROM icd9
			WHERE kode LIKE ? OR deskripsi_panjang LIKE ?
			ORDER BY kode=? DESC, validcode DESC, kode
			LIMIT 20
		`, pola, pola, tepat)
	}
	if err != nil {
		return nil, fmt.Errorf("cari master %s: %w", jenis, err)
	}
	defer rows.Close()

	hasil := make([]HasilValidasiCoding, 0, 20)
	for rows.Next() {
		item := HasilValidasiCoding{Jenis: jenis}
		if jenis == "diagnosa" {
			err = rows.Scan(&item.Kode, &item.Nama, &item.ValidCode, &item.AccPDX, &item.Asterisk, &item.IndonesiaM)
		} else {
			err = rows.Scan(&item.Kode, &item.Nama, &item.ValidCode, &item.AccPDX, &item.IndonesiaM)
		}
		if err != nil {
			return nil, fmt.Errorf("baca hasil pencarian %s: %w", jenis, err)
		}
		lengkapiHasilValidasi(&item, utama)
		hasil = append(hasil, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar master %s: %w", jenis, err)
	}
	return hasil, nil
}

func lengkapiHasilValidasi(hasil *HasilValidasiCoding, utama bool) {
	hasil.Kode = normalisasiKode(hasil.Kode)
	hasil.Nama = strings.TrimSpace(hasil.Nama)
	hasil.ValidCode = strings.TrimSpace(hasil.ValidCode)
	hasil.AccPDX = strings.ToUpper(strings.TrimSpace(hasil.AccPDX))
	hasil.Asterisk = strings.TrimSpace(hasil.Asterisk)
	hasil.IndonesiaM = strings.TrimSpace(hasil.IndonesiaM)
	hasil.BolehUtama = hasil.AccPDX != "N"
	hasil.Valid = false
	if hasil.ValidCode == "0" {
		hasil.Penanda = "Header"
		hasil.Pesan = fmt.Sprintf("Kode %s hanya berfungsi sebagai header (validcode 0)", hasil.Kode)
		return
	}
	if utama && !hasil.BolehUtama {
		hasil.Penanda = "Tidak bisa utama"
		hasil.Pesan = fmt.Sprintf("Kode %s tidak dapat menjadi %s utama (accpdx N)", hasil.Kode, hasil.Jenis)
		return
	}
	hasil.Valid = true
	hasil.Pesan = "Kode valid"
	if hasil.IndonesiaM == "1" {
		hasil.Penanda = "IM"
		hasil.Pesan = "Kode valid untuk IDRG (Indonesia Modification)"
	} else if !hasil.BolehUtama {
		hasil.Penanda = "Sekunder"
	}
}

func (r *Repositori) siapkanDanSinkronkanCoding(ctx context.Context, tx *sql.Tx, input *Input) error {
	diagnosa, err := r.validasiDiagnosa(ctx, tx, input)
	if err != nil {
		return err
	}
	prosedur, err := r.validasiProsedur(ctx, tx, input)
	if err != nil {
		return err
	}
	if err := r.sinkronkanDiagnosaJikaKosong(ctx, tx, input.NoRawat, diagnosa); err != nil {
		return err
	}
	return r.sinkronkanProsedurJikaKosong(ctx, tx, input.NoRawat, prosedur)
}

func (r *Repositori) validasiDiagnosa(ctx context.Context, tx *sql.Tx, input *Input) ([]pilihanCoding, error) {
	semua := []pilihanCoding{
		{kode: &input.KodeDiagnosaUtama, nama: &input.DiagnosaUtama, jenis: "diagnosa", prioritas: 1},
		{kode: &input.KodeDiagnosaSekunder, nama: &input.DiagnosaSekunder, jenis: "diagnosa", prioritas: 2},
		{kode: &input.KodeDiagnosaSekunder2, nama: &input.DiagnosaSekunder2, jenis: "diagnosa", prioritas: 3},
		{kode: &input.KodeDiagnosaSekunder3, nama: &input.DiagnosaSekunder3, jenis: "diagnosa", prioritas: 4},
		{kode: &input.KodeDiagnosaSekunder4, nama: &input.DiagnosaSekunder4, jenis: "diagnosa", prioritas: 5},
	}
	terpilih := make([]pilihanCoding, 0, len(semua))
	terpakai := make(map[string]struct{}, len(semua))
	for _, item := range semua {
		kode := normalisasiKode(*item.kode)
		nama := strings.TrimSpace(*item.nama)
		if kode == "" {
			if nama != "" {
				return nil, inputCodingTidakValid("kode %s untuk %q wajib diisi", item.jenis, nama)
			}
			*item.kode, *item.nama = "", ""
			continue
		}
		if _, ada := terpakai[kode]; ada {
			return nil, inputCodingTidakValid("kode diagnosa %s tidak boleh duplikat", kode)
		}
		err := tx.QueryRowContext(ctx, `
			SELECT COALESCE(nm_penyakit,''), COALESCE(validcode,'1'),
				COALESCE(accpdx,'Y'), COALESCE(asterisk,'0'), COALESCE(im,'0')
			FROM penyakit WHERE kd_penyakit=? LIMIT 1
		`, kode).Scan(item.nama, &item.validCode, &item.accPDX, &item.asterisk, &item.indonesiaM)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, inputCodingTidakValid("kode diagnosa %s tidak ditemukan pada master ICD-10", kode)
		}
		if err != nil {
			return nil, fmt.Errorf("validasi master diagnosa %s: %w", kode, err)
		}
		*item.kode = kode
		item.prioritas = len(terpilih) + 1
		item.validCode = strings.TrimSpace(item.validCode)
		item.accPDX = strings.ToUpper(strings.TrimSpace(item.accPDX))
		if item.validCode == "0" {
			return nil, inputCodingTidakValid("kode diagnosa %s hanya berfungsi sebagai header (validcode 0)", kode)
		}
		if item.prioritas == 1 && item.accPDX == "N" {
			return nil, inputCodingTidakValid("kode diagnosa %s tidak dapat menjadi diagnosa utama (accpdx N)", kode)
		}
		terpakai[kode] = struct{}{}
		terpilih = append(terpilih, item)
	}
	return terpilih, nil
}

func (r *Repositori) validasiProsedur(ctx context.Context, tx *sql.Tx, input *Input) ([]pilihanCoding, error) {
	semua := []pilihanCoding{
		{kode: &input.KodeProsedurUtama, nama: &input.ProsedurUtama, jenis: "prosedur", prioritas: 1},
		{kode: &input.KodeProsedurSekunder, nama: &input.ProsedurSekunder, jenis: "prosedur", prioritas: 2},
		{kode: &input.KodeProsedurSekunder2, nama: &input.ProsedurSekunder2, jenis: "prosedur", prioritas: 3},
		{kode: &input.KodeProsedurSekunder3, nama: &input.ProsedurSekunder3, jenis: "prosedur", prioritas: 4},
	}
	terpilih := make([]pilihanCoding, 0, len(semua))
	terpakai := make(map[string]struct{}, len(semua))
	for _, item := range semua {
		kode := normalisasiKode(*item.kode)
		nama := strings.TrimSpace(*item.nama)
		if kode == "" {
			if nama != "" {
				return nil, inputCodingTidakValid("kode %s untuk %q wajib diisi", item.jenis, nama)
			}
			*item.kode, *item.nama = "", ""
			continue
		}
		if _, ada := terpakai[kode]; ada {
			return nil, inputCodingTidakValid("kode prosedur %s tidak boleh duplikat", kode)
		}
		err := tx.QueryRowContext(ctx, `
			SELECT COALESCE(deskripsi_panjang,''), COALESCE(validcode,'1'),
				COALESCE(accpdx,'Y'), COALESCE(im,'0')
			FROM icd9 WHERE kode=? LIMIT 1
		`, kode).Scan(item.nama, &item.validCode, &item.accPDX, &item.indonesiaM)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, inputCodingTidakValid("kode prosedur %s tidak ditemukan pada master ICD-9-CM", kode)
		}
		if err != nil {
			return nil, fmt.Errorf("validasi master prosedur %s: %w", kode, err)
		}
		*item.kode = kode
		item.prioritas = len(terpilih) + 1
		item.validCode = strings.TrimSpace(item.validCode)
		item.accPDX = strings.ToUpper(strings.TrimSpace(item.accPDX))
		if item.validCode == "0" {
			return nil, inputCodingTidakValid("kode prosedur %s hanya berfungsi sebagai header (validcode 0)", kode)
		}
		if item.prioritas == 1 && item.accPDX == "N" {
			return nil, inputCodingTidakValid("kode prosedur %s tidak dapat menjadi prosedur utama (accpdx N)", kode)
		}
		terpakai[kode] = struct{}{}
		terpilih = append(terpilih, item)
	}
	return terpilih, nil
}

func (r *Repositori) sinkronkanDiagnosaJikaKosong(ctx context.Context, tx *sql.Tx, noRawat string, daftar []pilihanCoding) error {
	var jumlah int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM diagnosa_pasien WHERE no_rawat=?`, noRawat).Scan(&jumlah); err != nil {
		return fmt.Errorf("periksa diagnosa pasien: %w", err)
	}
	if jumlah > 0 {
		return nil
	}
	for urutan, item := range daftar {
		statusPenyakit, err := r.statusPenyakit(ctx, tx, noRawat, *item.kode)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO diagnosa_pasien
				(no_rawat, kd_penyakit, status, prioritas, status_penyakit)
			VALUES (?, ?, 'Ralan', ?, ?)
		`, noRawat, *item.kode, urutan+1, statusPenyakit)
		if err != nil {
			return fmt.Errorf("simpan diagnosa pasien %s: %w", *item.kode, err)
		}
	}
	return nil
}

func (r *Repositori) sinkronkanProsedurJikaKosong(ctx context.Context, tx *sql.Tx, noRawat string, daftar []pilihanCoding) error {
	var jumlah int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM prosedur_pasien WHERE no_rawat=?`, noRawat).Scan(&jumlah); err != nil {
		return fmt.Errorf("periksa prosedur pasien: %w", err)
	}
	if jumlah > 0 {
		return nil
	}
	for urutan, item := range daftar {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO prosedur_pasien (no_rawat, kode, status, prioritas, jumlah)
			VALUES (?, ?, 'Ralan', ?, 1)
		`, noRawat, *item.kode, urutan+1)
		if err != nil {
			return fmt.Errorf("simpan prosedur pasien %s: %w", *item.kode, err)
		}
	}
	return nil
}

func (r *Repositori) statusPenyakit(ctx context.Context, tx *sql.Tx, noRawat, kode string) (string, error) {
	var pernah int
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM diagnosa_pasien dp
		INNER JOIN reg_periksa lama ON lama.no_rawat=dp.no_rawat
		INNER JOIN reg_periksa sekarang ON sekarang.no_rawat=?
		WHERE lama.no_rkm_medis=sekarang.no_rkm_medis
			AND dp.kd_penyakit=? AND dp.no_rawat<>?
	`, noRawat, kode, noRawat).Scan(&pernah)
	if err != nil {
		return "", fmt.Errorf("periksa riwayat diagnosa %s: %w", kode, err)
	}
	if pernah > 0 {
		return "Lama", nil
	}
	return "Baru", nil
}

func normalisasiKode(kode string) string {
	return strings.ToUpper(strings.TrimSpace(kode))
}

func inputCodingTidakValid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInputTidakValid, fmt.Sprintf(format, args...))
}
