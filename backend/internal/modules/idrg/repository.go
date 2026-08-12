package idrg

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
)

type Repositori struct {
	simrsDB *sql.DB
}

func NewRepositori(simrsDB *sql.DB) *Repositori {
	return &Repositori{simrsDB: simrsDB}
}

type FilterPasien struct {
	Tanggal    string `json:"tanggal"`
	JenisRawat string `json:"jenis_rawat"`
	Pencarian  string `json:"pencarian"`
	Halaman    int    `json:"halaman"`
	Batas      int    `json:"batas"`
}

type Paginasi struct {
	Halaman      int   `json:"halaman"`
	Batas        int   `json:"batas"`
	Total        int64 `json:"total"`
	TotalHalaman int   `json:"total_halaman"`
}

type DaftarPasien struct {
	Data     []PasienKlaim `json:"data"`
	Paginasi Paginasi      `json:"paginasi"`
	Filter   FilterPasien  `json:"filter"`
}

type PasienKlaim struct {
	JenisRawatData string         `json:"jenis_rawat_data"`
	NoRawat        string         `json:"no_rawat"`
	NoRekamMedis   string         `json:"no_rkm_medis"`
	NamaPasien     string         `json:"nm_pasien"`
	TanggalLahir   string         `json:"tgl_lahir"`
	JenisKelamin   string         `json:"jk"`
	NoKartu        string         `json:"no_kartu"`
	NoSEP          string         `json:"no_sep"`
	StatusLanjut   string         `json:"status_lanjut"`
	KelasRawat     string         `json:"kelas_rawat"`
	TanggalMasuk   string         `json:"tgl_masuk"`
	TanggalKeluar  string         `json:"tgl_keluar"`
	DiagnosaAwal   string         `json:"diagnosa_awal"`
	DiagnosaAkhir  string         `json:"diagnosa_akhir"`
	NamaDokter     string         `json:"nm_dokter"`
	Kamar          string         `json:"kamar"`
	StatusPulang   string         `json:"status_pulang"`
	LamaRawat      string         `json:"lama_rawat"`
	Penanggung     string         `json:"penanggung"`
	StatusKlaim    string         `json:"status_klaim"`
	TarifRS        map[string]any `json:"tarif_rs"`
}

type Diagnosa struct {
	Kode   string `json:"kd_diag"`
	Nama   string `json:"nm_diag"`
	Status string `json:"status"`
}

type Prosedur struct {
	Kode         string `json:"kd_prosedur"`
	Nama         string `json:"nm_prosedur"`
	Multiplicity int    `json:"multiplicity"`
}

type HasilCoding struct {
	Sukses   bool           `json:"sukses"`
	Kode     int            `json:"kode"`
	Pesan    string         `json:"pesan"`
	Metadata map[string]any `json:"metadata"`
	Data     map[string]any `json:"data"`
	Response map[string]any `json:"response"`
}

func (r *Repositori) DaftarPasien(ctx context.Context, filter FilterPasien) (DaftarPasien, error) {
	rapikanFilter(&filter)

	if filter.JenisRawat == "ralan" {
		return r.daftarPasienRawatJalan(ctx, filter)
	}
	return r.daftarPasienRawatInap(ctx, filter)
}

func (r *Repositori) AmbilDiagnosa(ctx context.Context, noRawat string) (HasilCoding, error) {
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT
			COALESCE(dp.kd_penyakit, ''),
			COALESCE(p.nm_penyakit, ''),
			COALESCE(dp.status, '')
		FROM diagnosa_pasien dp
		LEFT JOIN penyakit p ON p.kd_penyakit = dp.kd_penyakit
		WHERE dp.no_rawat = ?
		ORDER BY dp.prioritas ASC
	`, noRawat)
	if err != nil {
		return HasilCoding{}, fmt.Errorf("read SIMRS diagnosa pasien: %w", err)
	}
	defer rows.Close()

	daftar := make([]Diagnosa, 0)
	for rows.Next() {
		var data Diagnosa
		if err := rows.Scan(&data.Kode, &data.Nama, &data.Status); err != nil {
			return HasilCoding{}, fmt.Errorf("scan SIMRS diagnosa pasien: %w", err)
		}
		if len(daftar) == 0 {
			data.Status = "primer"
		} else {
			data.Status = "sekunder"
		}
		daftar = append(daftar, data)
	}
	if err := rows.Err(); err != nil {
		return HasilCoding{}, fmt.Errorf("iterate SIMRS diagnosa pasien: %w", err)
	}
	if len(daftar) == 0 {
		var err error
		daftar, err = r.ambilDiagnosaResume(ctx, noRawat)
		if err != nil {
			return HasilCoding{}, err
		}
	}

	pesan := "Ok"
	if len(daftar) == 0 {
		pesan = "Diagnosa belum ditemukan di SIMRS."
	}
	return hasilDiagnosa(daftar, pesan), nil
}

func (r *Repositori) AmbilProsedur(ctx context.Context, noRawat string) (HasilCoding, error) {
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT
			COALESCE(pp.kode, ''),
			COALESCE(i.deskripsi_panjang, ''),
			COALESCE(NULLIF(pp.jumlah, 0), 1)
		FROM prosedur_pasien pp
		LEFT JOIN icd9 i ON i.kode = pp.kode
		WHERE pp.no_rawat = ?
		ORDER BY pp.prioritas ASC
	`, noRawat)
	if err != nil {
		return HasilCoding{}, fmt.Errorf("read SIMRS prosedur pasien: %w", err)
	}
	defer rows.Close()

	daftar := make([]Prosedur, 0)
	for rows.Next() {
		var data Prosedur
		if err := rows.Scan(&data.Kode, &data.Nama, &data.Multiplicity); err != nil {
			return HasilCoding{}, fmt.Errorf("scan SIMRS prosedur pasien: %w", err)
		}
		if data.Multiplicity < 1 {
			data.Multiplicity = 1
		}
		daftar = append(daftar, data)
	}
	if err := rows.Err(); err != nil {
		return HasilCoding{}, fmt.Errorf("iterate SIMRS prosedur pasien: %w", err)
	}

	pesan := "Ok"
	if len(daftar) == 0 {
		pesan = "Prosedur belum ditemukan di SIMRS."
	}
	return hasilProsedur(daftar, pesan), nil
}

func (r *Repositori) NomorSEP(ctx context.Context, noRawat string) (string, error) {
	noRawat = strings.TrimSpace(noRawat)
	if noRawat == "" {
		return "", nil
	}

	var nomorSEP string
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT COALESCE(no_sep, '')
		FROM bridging_sep
		WHERE no_rawat = ?
		ORDER BY jnspelayanan = '1' DESC, tglsep DESC, no_sep DESC
		LIMIT 1
	`, noRawat).Scan(&nomorSEP)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("read SIMRS nomor SEP: %w", err)
	}
	return strings.TrimSpace(nomorSEP), nil
}

func (r *Repositori) AmbilCoderNIK(ctx context.Context, username string) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return "", nil
	}

	var coderNIK string
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT COALESCE(no_ik, '')
		FROM inacbg_coder_nik
		WHERE nik = ?
		LIMIT 1
	`, username).Scan(&coderNIK)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("read SIMRS coder NIK: %w", err)
	}
	return strings.TrimSpace(coderNIK), nil
}

func (r *Repositori) CariDiagnosa(ctx context.Context, kataKunci string) (HasilCoding, error) {
	kataKunci = strings.TrimSpace(kataKunci)
	if kataKunci == "" {
		return hasilDiagnosa([]Diagnosa{}, "Diagnosa tidak ditemukan."), nil
	}

	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT
			COALESCE(kd_penyakit, ''),
			COALESCE(nm_penyakit, '')
		FROM penyakit
		WHERE kd_penyakit LIKE ? OR nm_penyakit LIKE ?
		ORDER BY kd_penyakit = ? DESC, kd_penyakit
		LIMIT 20
	`, "%"+kataKunci+"%", "%"+kataKunci+"%", kataKunci)
	if err != nil {
		return HasilCoding{}, fmt.Errorf("search SIMRS diagnosa: %w", err)
	}
	defer rows.Close()

	daftar := make([]Diagnosa, 0)
	for rows.Next() {
		var data Diagnosa
		if err := rows.Scan(&data.Kode, &data.Nama); err != nil {
			return HasilCoding{}, fmt.Errorf("scan SIMRS diagnosa: %w", err)
		}
		daftar = append(daftar, data)
	}
	if err := rows.Err(); err != nil {
		return HasilCoding{}, fmt.Errorf("iterate SIMRS diagnosa: %w", err)
	}

	pesan := "Ok"
	if len(daftar) == 0 {
		pesan = "Diagnosa tidak ditemukan."
	}
	return hasilDiagnosa(daftar, pesan), nil
}

func (r *Repositori) CariProsedur(ctx context.Context, kataKunci string) (HasilCoding, error) {
	kataKunci = strings.TrimSpace(kataKunci)
	if kataKunci == "" {
		return hasilProsedur([]Prosedur{}, "Prosedur tidak ditemukan."), nil
	}

	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT
			COALESCE(kode, ''),
			COALESCE(deskripsi_panjang, '')
		FROM icd9
		WHERE kode LIKE ? OR deskripsi_panjang LIKE ?
		ORDER BY kode = ? DESC, kode
		LIMIT 20
	`, "%"+kataKunci+"%", "%"+kataKunci+"%", kataKunci)
	if err != nil {
		return HasilCoding{}, fmt.Errorf("search SIMRS prosedur: %w", err)
	}
	defer rows.Close()

	daftar := make([]Prosedur, 0)
	for rows.Next() {
		var data Prosedur
		if err := rows.Scan(&data.Kode, &data.Nama); err != nil {
			return HasilCoding{}, fmt.Errorf("scan SIMRS prosedur: %w", err)
		}
		data.Multiplicity = 1
		daftar = append(daftar, data)
	}
	if err := rows.Err(); err != nil {
		return HasilCoding{}, fmt.Errorf("iterate SIMRS prosedur: %w", err)
	}

	pesan := "Ok"
	if len(daftar) == 0 {
		pesan = "Prosedur tidak ditemukan."
	}
	return hasilProsedur(daftar, pesan), nil
}

func (r *Repositori) daftarPasienRawatInap(ctx context.Context, filter FilterPasien) (DaftarPasien, error) {
	kondisi := []string{"r.tgl_registrasi = ?", "r.kd_pj = 'BPJ'"}
	argumen := []any{filter.Tanggal}
	if filter.Pencarian != "" {
		kondisi = append(kondisi, `(
			ki.no_rawat LIKE ? OR r.no_rkm_medis LIKE ? OR p.nm_pasien LIKE ?
			OR COALESCE(sep.no_sep, '') LIKE ? OR COALESCE(ki.kd_kamar, '') LIKE ?
			OR COALESCE(ki.diagnosa_awal, '') LIKE ? OR COALESCE(ki.diagnosa_akhir, '') LIKE ?
		)`)
		kataKunci := "%" + filter.Pencarian + "%"
		argumen = append(argumen, kataKunci, kataKunci, kataKunci, kataKunci, kataKunci, kataKunci, kataKunci)
	}
	kondisi = append(kondisi, `NOT EXISTS (
		SELECT 1
		FROM kamar_inap ki_lain
		WHERE ki_lain.no_rawat = ki.no_rawat
		  AND CONCAT(ki_lain.tgl_masuk, ' ', ki_lain.jam_masuk, ' ', ki_lain.kd_kamar) > CONCAT(ki.tgl_masuk, ' ', ki.jam_masuk, ' ', ki.kd_kamar)
	)`)

	total, err := r.hitungRawatInap(ctx, kondisi, argumen)
	if err != nil {
		return DaftarPasien{}, err
	}
	paginasi := paginasiDariTotal(filter.Halaman, filter.Batas, total)
	argumenData := append([]any{}, argumen...)
	argumenData = append(argumenData, paginasi.Batas, offsetPaginasi(paginasi.Halaman, paginasi.Batas))

	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT
			'ranap',
			COALESCE(ki.no_rawat, ''),
			COALESCE(r.no_rkm_medis, ''),
			COALESCE(NULLIF(p.nm_pasien, ''), CONCAT('RM ', r.no_rkm_medis), ''),
			COALESCE(DATE_FORMAT(p.tgl_lahir, '%Y-%m-%d'), ''),
			COALESCE(p.jk, ''),
			COALESCE(sep.no_kartu, p.no_peserta, ''),
			COALESCE(sep.no_sep, ''),
			'Ranap',
			COALESCE(sep.klsrawat, k.kelas, '3'),
			COALESCE(DATE_FORMAT(r.tgl_registrasi, '%Y-%m-%d'), ''),
			CASE
				WHEN ki.tgl_keluar = '0000-00-00' THEN COALESCE(DATE_FORMAT(r.tgl_registrasi, '%Y-%m-%d'), '')
				ELSE COALESCE(DATE_FORMAT(ki.tgl_keluar, '%Y-%m-%d'), '')
			END,
			COALESCE(ki.diagnosa_awal, ''),
			COALESCE(ki.diagnosa_akhir, ''),
			COALESCE((
				SELECT GROUP_CONCAT(DISTINCT d_dpjp.nm_dokter ORDER BY d_dpjp.nm_dokter SEPARATOR '#')
				FROM dpjp_ranap dp
				JOIN dokter d_dpjp ON d_dpjp.kd_dokter = dp.kd_dokter
				WHERE dp.no_rawat = ki.no_rawat
			), COALESCE(d.nm_dokter, '')),
			TRIM(CONCAT(COALESCE(b.nm_bangsal, ''), ' ', COALESCE(ki.kd_kamar, ''))),
			COALESCE(NULLIF(ki.stts_pulang, ''), 'Dirawat'),
			COALESCE(ki.lama, ''),
			COALESCE(pj.png_jawab, ''),
			CASE
				WHEN COALESCE(id.status, '') <> '' THEN id.status
				WHEN EXISTS (
					SELECT 1
					FROM inacbg_klaim_baru2 kb
					WHERE kb.no_rawat = ki.no_rawat
					   OR (COALESCE(sep.no_sep, '') <> '' AND kb.no_sep = sep.no_sep)
				) THEN 'new_claim'
				ELSE 'belum'
			END
		FROM kamar_inap ki
		JOIN reg_periksa r ON r.no_rawat = ki.no_rawat
		LEFT JOIN pasien p ON p.no_rkm_medis = r.no_rkm_medis
		LEFT JOIN dokter d ON d.kd_dokter = r.kd_dokter
		LEFT JOIN penjab pj ON pj.kd_pj = r.kd_pj
		LEFT JOIN kamar k ON k.kd_kamar = ki.kd_kamar
		LEFT JOIN bangsal b ON b.kd_bangsal = k.kd_bangsal
		LEFT JOIN bridging_sep sep ON sep.no_rawat = ki.no_rawat AND sep.jnspelayanan = '1'
		LEFT JOIN inacbg_data id ON id.no_sep = sep.no_sep
		WHERE `+strings.Join(kondisi, " AND ")+`
		ORDER BY ki.tgl_masuk, ki.no_rawat
		LIMIT ? OFFSET ?
	`, argumenData...)
	if err != nil {
		return DaftarPasien{}, fmt.Errorf("read SIMRS IDRG inpatients: %w", err)
	}
	defer rows.Close()

	daftar, err := scanPasienKlaim(rows)
	if err != nil {
		return DaftarPasien{}, err
	}
	if err := r.isiTarifKlaim(ctx, daftar); err != nil {
		return DaftarPasien{}, err
	}
	return DaftarPasien{Data: daftar, Paginasi: paginasi, Filter: filter}, nil
}

func (r *Repositori) daftarPasienRawatJalan(ctx context.Context, filter FilterPasien) (DaftarPasien, error) {
	kondisi := []string{"r.tgl_registrasi = ?", "r.status_lanjut = 'Ralan'", "r.kd_poli NOT LIKE '%IGD%'", "r.kd_pj = 'BPJ'"}
	argumen := []any{filter.Tanggal}
	if filter.Pencarian != "" {
		kondisi = append(kondisi, "(r.no_rawat LIKE ? OR r.no_rkm_medis LIKE ? OR p.nm_pasien LIKE ? OR COALESCE(sep.no_sep, '') LIKE ?)")
		kataKunci := "%" + filter.Pencarian + "%"
		argumen = append(argumen, kataKunci, kataKunci, kataKunci, kataKunci)
	}

	total, err := r.hitungRawatJalan(ctx, kondisi, argumen)
	if err != nil {
		return DaftarPasien{}, err
	}
	paginasi := paginasiDariTotal(filter.Halaman, filter.Batas, total)
	argumenData := append([]any{}, argumen...)
	argumenData = append(argumenData, paginasi.Batas, offsetPaginasi(paginasi.Halaman, paginasi.Batas))

	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT
			'ralan',
			COALESCE(r.no_rawat, ''),
			COALESCE(r.no_rkm_medis, ''),
			COALESCE(NULLIF(p.nm_pasien, ''), CONCAT('RM ', r.no_rkm_medis), ''),
			COALESCE(DATE_FORMAT(p.tgl_lahir, '%Y-%m-%d'), ''),
			COALESCE(p.jk, ''),
			COALESCE(sep.no_kartu, p.no_peserta, ''),
			COALESCE(sep.no_sep, ''),
			'Ralan',
			'3',
			COALESCE(DATE_FORMAT(r.tgl_registrasi, '%Y-%m-%d'), ''),
			COALESCE(DATE_FORMAT(r.tgl_registrasi, '%Y-%m-%d'), ''),
			'',
			'',
			COALESCE(d.nm_dokter, ''),
			COALESCE(poly.nm_poli, r.kd_poli, ''),
			'',
			'1',
			COALESCE(pj.png_jawab, ''),
			CASE
				WHEN COALESCE(id.status, '') <> '' THEN id.status
				WHEN EXISTS (
					SELECT 1
					FROM inacbg_klaim_baru2 kb
					WHERE kb.no_rawat = r.no_rawat
					   OR (COALESCE(sep.no_sep, '') <> '' AND kb.no_sep = sep.no_sep)
				) THEN 'new_claim'
				ELSE 'belum'
			END
		FROM reg_periksa r
		LEFT JOIN pasien p ON p.no_rkm_medis = r.no_rkm_medis
		LEFT JOIN dokter d ON d.kd_dokter = r.kd_dokter
		LEFT JOIN penjab pj ON pj.kd_pj = r.kd_pj
		LEFT JOIN poliklinik poly ON poly.kd_poli = r.kd_poli
		LEFT JOIN bridging_sep sep ON sep.no_rawat = r.no_rawat AND sep.jnspelayanan = '2'
		LEFT JOIN inacbg_data id ON id.no_sep = sep.no_sep
		WHERE `+strings.Join(kondisi, " AND ")+`
		ORDER BY r.jam_reg DESC
		LIMIT ? OFFSET ?
	`, argumenData...)
	if err != nil {
		return DaftarPasien{}, fmt.Errorf("read SIMRS IDRG outpatients: %w", err)
	}
	defer rows.Close()

	daftar, err := scanPasienKlaim(rows)
	if err != nil {
		return DaftarPasien{}, err
	}
	if err := r.isiTarifKlaim(ctx, daftar); err != nil {
		return DaftarPasien{}, err
	}
	return DaftarPasien{Data: daftar, Paginasi: paginasi, Filter: filter}, nil
}

func (r *Repositori) isiTarifKlaim(ctx context.Context, daftar []PasienKlaim) error {
	for index := range daftar {
		tarif, err := r.HitungTarifKlaim(ctx, daftar[index].NoRawat)
		if err != nil {
			return err
		}
		daftar[index].TarifRS = tarif
	}
	return nil
}

func (r *Repositori) HitungTarifKlaim(ctx context.Context, noRawat string) (map[string]any, error) {
	noRawat = strings.TrimSpace(noRawat)
	tarif := tarifKosongFloat()
	if noRawat == "" {
		return tarifString(tarif), nil
	}

	var biayaRegistrasi float64
	if err := r.simrsDB.QueryRowContext(ctx, `
		SELECT COALESCE(biaya_reg, 0)
		FROM reg_periksa
		WHERE no_rawat = ?
		LIMIT 1
	`, noRawat).Scan(&biayaRegistrasi); err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("read SIMRS biaya registrasi: %w", err)
	}
	tarif["kamar"] = biayaRegistrasi

	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT
			COALESCE(status, ''),
			COALESCE(nm_perawatan, ''),
			COALESCE(SUM(totalbiaya), 0)
		FROM billing
		WHERE no_rawat = ?
		  AND status IN (
			'Ralan Dokter Paramedis',
			'Ranap Dokter Paramedis',
			'Operasi',
			'Ranap Dokter',
			'Ralan Dokter',
			'Ranap Paramedis',
			'Ralan Paramedis',
			'Radiologi',
			'Laborat',
			'Kamar',
			'Obat',
			'Retur Obat',
			'Resep Pulang',
			'Tambahan',
			'Harian',
			'Service'
		  )
		GROUP BY status, nm_perawatan
	`, noRawat)
	if err != nil {
		return nil, fmt.Errorf("read SIMRS billing klaim: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var namaPerawatan string
		var total float64
		if err := rows.Scan(&status, &namaPerawatan, &total); err != nil {
			return nil, fmt.Errorf("scan SIMRS billing klaim: %w", err)
		}

		status = strings.TrimSpace(status)
		namaLower := strings.ToLower(namaPerawatan)
		switch status {
		case "Ralan Dokter Paramedis", "Ranap Dokter Paramedis":
			if strings.Contains(namaLower, "terapi") {
				tarif["rehabilitasi"] += total
			} else {
				tarif["prosedur_non_bedah"] += total
			}
		case "Operasi":
			tarif["prosedur_bedah"] += total
		case "Radiologi":
			tarif["radiologi"] += total
		case "Laborat":
			tarif["laboratorium"] += total
		case "Tambahan":
			tarif["bmhp"] += total
		case "Ranap Dokter", "Ralan Dokter":
			tarif["konsultasi"] += total
		case "Ranap Paramedis", "Ralan Paramedis":
			tarif["keperawatan"] += total
		case "Kamar":
			tarif["kamar"] += total
		case "Harian", "Service":
			tarif["sewa_alat"] += total
		case "Obat", "Retur Obat", "Resep Pulang":
			switch {
			case strings.Contains(namaLower, "kronis"):
				tarif["obat_kronis"] += total
			case strings.Contains(namaLower, "kemo"):
				tarif["obat_kemoterapi"] += total
			default:
				tarif["obat"] += total
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate SIMRS billing klaim: %w", err)
	}

	return tarifString(tarif), nil
}

func (r *Repositori) hitungRawatInap(ctx context.Context, kondisi []string, argumen []any) (int64, error) {
	var total int64
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM kamar_inap ki
		JOIN reg_periksa r ON r.no_rawat = ki.no_rawat
		LEFT JOIN pasien p ON p.no_rkm_medis = r.no_rkm_medis
		LEFT JOIN bridging_sep sep ON sep.no_rawat = ki.no_rawat AND sep.jnspelayanan = '1'
		WHERE `+strings.Join(kondisi, " AND "), argumen...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count SIMRS IDRG inpatients: %w", err)
	}
	return total, nil
}

func (r *Repositori) hitungRawatJalan(ctx context.Context, kondisi []string, argumen []any) (int64, error) {
	var total int64
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM reg_periksa r
		LEFT JOIN pasien p ON p.no_rkm_medis = r.no_rkm_medis
		LEFT JOIN bridging_sep sep ON sep.no_rawat = r.no_rawat AND sep.jnspelayanan = '2'
		WHERE `+strings.Join(kondisi, " AND "), argumen...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count SIMRS IDRG outpatients: %w", err)
	}
	return total, nil
}

func (r *Repositori) ambilDiagnosaResume(ctx context.Context, noRawat string) ([]Diagnosa, error) {
	return r.ambilDiagnosaResumeDariTabel(ctx, "resume_pasien_ranap", noRawat)
}

func (r *Repositori) ambilDiagnosaResumeDariTabel(ctx context.Context, namaTabel string, noRawat string) ([]Diagnosa, error) {
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT
			COALESCE(kd_diagnosa_utama, ''), COALESCE(diagnosa_utama, ''),
			COALESCE(kd_diagnosa_sekunder, ''), COALESCE(diagnosa_sekunder, ''),
			COALESCE(kd_diagnosa_sekunder2, ''), COALESCE(diagnosa_sekunder2, ''),
			COALESCE(kd_diagnosa_sekunder3, ''), COALESCE(diagnosa_sekunder3, ''),
			COALESCE(kd_diagnosa_sekunder4, ''), COALESCE(diagnosa_sekunder4, '')
		FROM `+namaTabel+`
		WHERE no_rawat = ?
		LIMIT 1
	`, noRawat)
	if err != nil {
		if namaTabel == "resume_pasien_ranap" {
			return r.ambilDiagnosaResumeDariTabel(ctx, "resume_pasien", noRawat)
		}
		return nil, fmt.Errorf("read SIMRS resume diagnosa: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if namaTabel == "resume_pasien_ranap" {
			return r.ambilDiagnosaResumeDariTabel(ctx, "resume_pasien", noRawat)
		}
		return []Diagnosa{}, nil
	}

	var nilai [10]string
	args := make([]any, 0, len(nilai))
	for i := range nilai {
		args = append(args, &nilai[i])
	}
	if err := rows.Scan(args...); err != nil {
		return nil, fmt.Errorf("scan SIMRS resume diagnosa: %w", err)
	}

	daftar := make([]Diagnosa, 0)
	for i := 0; i < len(nilai); i += 2 {
		kode := strings.TrimSpace(nilai[i])
		nama := strings.TrimSpace(nilai[i+1])
		if kode == "" || kode == "-" {
			continue
		}
		status := "sekunder"
		if len(daftar) == 0 {
			status = "primer"
		}
		if nama == "" || nama == "-" {
			nama = kode
		}
		daftar = append(daftar, Diagnosa{Kode: kode, Nama: nama, Status: status})
	}
	return daftar, nil
}

func scanPasienKlaim(rows *sql.Rows) ([]PasienKlaim, error) {
	daftar := make([]PasienKlaim, 0)
	for rows.Next() {
		data := PasienKlaim{TarifRS: tarifKosong()}
		if err := rows.Scan(
			&data.JenisRawatData, &data.NoRawat, &data.NoRekamMedis, &data.NamaPasien,
			&data.TanggalLahir, &data.JenisKelamin, &data.NoKartu, &data.NoSEP,
			&data.StatusLanjut, &data.KelasRawat, &data.TanggalMasuk, &data.TanggalKeluar,
			&data.DiagnosaAwal, &data.DiagnosaAkhir, &data.NamaDokter, &data.Kamar,
			&data.StatusPulang, &data.LamaRawat, &data.Penanggung, &data.StatusKlaim,
		); err != nil {
			return nil, fmt.Errorf("scan SIMRS IDRG patient: %w", err)
		}
		if data.StatusKlaim == "" {
			data.StatusKlaim = "belum"
		}
		daftar = append(daftar, data)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate SIMRS IDRG patient: %w", err)
	}
	return daftar, nil
}

func rapikanFilter(filter *FilterPasien) {
	filter.JenisRawat = strings.ToLower(strings.TrimSpace(filter.JenisRawat))
	if filter.JenisRawat != "ralan" {
		filter.JenisRawat = "ranap"
	}
	filter.Pencarian = strings.TrimSpace(filter.Pencarian)
	if filter.Halaman < 1 {
		filter.Halaman = 1
	}
	if filter.Batas < 1 {
		filter.Batas = 50
	}
	if filter.Batas > 200 {
		filter.Batas = 200
	}
}

func paginasiDariTotal(halaman int, batas int, total int64) Paginasi {
	totalHalaman := 0
	if batas > 0 && total > 0 {
		totalHalaman = int((total + int64(batas) - 1) / int64(batas))
	}
	return Paginasi{Halaman: halaman, Batas: batas, Total: total, TotalHalaman: totalHalaman}
}

func offsetPaginasi(halaman int, batas int) int {
	if halaman < 1 {
		halaman = 1
	}
	return (halaman - 1) * batas
}

func hasilDiagnosa(daftar []Diagnosa, pesan string) HasilCoding {
	return HasilCoding{
		Sukses:   true,
		Kode:     200,
		Pesan:    pesan,
		Metadata: map[string]any{"code": 200, "message": pesan},
		Data:     map[string]any{"diagnosa": daftar},
		Response: map[string]any{"diagnosa": daftar},
	}
}

func hasilProsedur(daftar []Prosedur, pesan string) HasilCoding {
	return HasilCoding{
		Sukses:   true,
		Kode:     200,
		Pesan:    pesan,
		Metadata: map[string]any{"code": 200, "message": pesan},
		Data:     map[string]any{"prosedur": daftar, "procedure": daftar},
		Response: map[string]any{"prosedur": daftar, "procedure": daftar},
	}
}

func tarifKosong() map[string]any {
	return map[string]any{
		"prosedur_non_bedah": "0",
		"prosedur_bedah":     "0",
		"konsultasi":         "0",
		"tenaga_ahli":        "0",
		"keperawatan":        "0",
		"penunjang":          "0",
		"radiologi":          "0",
		"laboratorium":       "0",
		"pelayanan_darah":    "0",
		"rehabilitasi":       "0",
		"kamar":              "0",
		"rawat_intensif":     "0",
		"obat":               "0",
		"obat_kronis":        "0",
		"obat_kemoterapi":    "0",
		"alkes":              "0",
		"bmhp":               "0",
		"sewa_alat":          "0",
	}
}

func tarifKosongFloat() map[string]float64 {
	return map[string]float64{
		"prosedur_non_bedah": 0,
		"prosedur_bedah":     0,
		"konsultasi":         0,
		"tenaga_ahli":        0,
		"keperawatan":        0,
		"penunjang":          0,
		"radiologi":          0,
		"laboratorium":       0,
		"pelayanan_darah":    0,
		"rehabilitasi":       0,
		"kamar":              0,
		"rawat_intensif":     0,
		"obat":               0,
		"obat_kronis":        0,
		"obat_kemoterapi":    0,
		"alkes":              0,
		"bmhp":               0,
		"sewa_alat":          0,
	}
}

func tarifString(tarif map[string]float64) map[string]any {
	hasil := make(map[string]any, len(tarif))
	for key, value := range tarif {
		angka := int64(math.Round(value))
		if angka < 0 {
			angka = 0
		}
		hasil[key] = fmt.Sprintf("%d", angka)
	}
	return hasil
}
