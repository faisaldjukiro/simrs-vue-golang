// Package beranda membaca data yang ditampilkan di halaman beranda.
package beranda

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

type Ringkasan struct {
	JumlahRegistrasi int64 `json:"jumlah_registrasi"`
	JumlahRawatJalan int64 `json:"jumlah_rawat_jalan"`
	JumlahIGD        int64 `json:"jumlah_igd"`
	JumlahRawatInap  int64 `json:"jumlah_rawat_inap"`
}

type Pasien struct {
	NoRawat             string `json:"no_rawat"`
	NamaPasien          string `json:"nama_pasien"`
	NoRekamMedis        string `json:"no_rekam_medis"`
	NoRegistrasi        string `json:"no_registrasi"`
	JenisKelamin        string `json:"jenis_kelamin"`
	TanggalLahir        string `json:"tanggal_lahir"`
	Umur                string `json:"umur"`
	Alamat              string `json:"alamat"`
	NoTelepon           string `json:"no_telepon"`
	Poliklinik          string `json:"poliklinik"`
	KodePoliklinik      string `json:"kode_poliklinik"`
	Dokter              string `json:"dokter"`
	Penjamin            string `json:"penjamin"`
	Status              string `json:"status"`
	StatusBayar         string `json:"status_bayar"`
	TanggalRegistrasi   string `json:"tanggal_registrasi"`
	JamRegistrasi       string `json:"jam_registrasi"`
	Kamar               string `json:"kamar,omitempty"`
	DiagnosaAwal        string `json:"diagnosa_awal,omitempty"`
	LamaRawat           string `json:"lama_rawat,omitempty"`
	NoSEP               string `json:"no_sep,omitempty"`
	KelasSEP            string `json:"kelas_sep,omitempty"`
	TanggalSEP          string `json:"tanggal_sep,omitempty"`
	JenisPerawatanSIMRS string `json:"-"`
}

type Poliklinik struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}

type Dokter struct {
	Kode          string `json:"kode"`
	Nama          string `json:"nama"`
	KodeSpesialis string `json:"kode_spesialis"`
	Status        string `json:"status"`
}

type SidebarPasien struct {
	Kode        string   `json:"kode"`
	Nama        string   `json:"nama"`
	Ikon        string   `json:"ikon"`
	DaftarModul []string `json:"daftar_modul"`
}

type FilterPasien struct {
	TanggalMulai         string `json:"date_from"`
	TanggalSelesai       string `json:"date_to"`
	Status               string `json:"status"`
	StatusBayar          string `json:"status_bayar"`
	Poliklinik           string `json:"poly"`
	Dokter               string `json:"dokter"`
	Pencarian            string `json:"search"`
	BelumPulang          bool   `json:"belum_pulang"`
	TanggalDipilihManual bool   `json:"-"`
	Halaman              int    `json:"halaman"`
	Batas                int    `json:"batas"`
}

type FilterBeranda struct {
	RawatJalan FilterPasien
	IGD        FilterPasien
	RawatInap  FilterPasien
}

type PilihanStatus struct {
	Periksa     []string `json:"periksa"`
	RawatInap   []string `json:"rawat_inap"`
	StatusBayar []string `json:"status_bayar"`
}

type Paginasi struct {
	Halaman      int   `json:"halaman"`
	Batas        int   `json:"batas"`
	Total        int64 `json:"total"`
	TotalHalaman int   `json:"total_halaman"`
}

type PaginasiBeranda struct {
	Registrasi Paginasi `json:"registrasi"`
	RawatJalan Paginasi `json:"rawat_jalan"`
	IGD        Paginasi `json:"igd"`
	RawatInap  Paginasi `json:"rawat_inap"`
}

type Beranda struct {
	Ringkasan     Ringkasan       `json:"ringkasan"`
	Registrasi    []Pasien        `json:"registrasi"`
	RawatJalan    []Pasien        `json:"rawat_jalan"`
	IGD           []Pasien        `json:"igd"`
	RawatInap     []Pasien        `json:"rawat_inap"`
	Paginasi      PaginasiBeranda `json:"paginasi"`
	Poliklinik    []Poliklinik    `json:"poliklinik"`
	Dokter        []Dokter        `json:"dokter"`
	PilihanStatus PilihanStatus   `json:"pilihan_status"`
	SidebarPasien []SidebarPasien `json:"sidebar_pasien"`
}

type Repositori struct {
	simrsDB    *sql.DB
	aplikasiDB *sql.DB
}

type hasilPasien struct {
	Data     []Pasien
	Paginasi Paginasi
}

func NewRepositori(simrsDB *sql.DB, aplikasiDB *sql.DB) *Repositori {
	return &Repositori{simrsDB: simrsDB, aplikasiDB: aplikasiDB}
}

func (r *Repositori) BacaBeranda(ctx context.Context, filter FilterBeranda, userID uint64) (Beranda, error) {
	tx, err := r.simrsDB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return Beranda{}, fmt.Errorf("begin SIMRS read-only transaction: %w", err)
	}
	defer tx.Rollback()

	hasil := Beranda{
		Registrasi:    make([]Pasien, 0),
		RawatJalan:    make([]Pasien, 0),
		IGD:           make([]Pasien, 0),
		RawatInap:     make([]Pasien, 0),
		Poliklinik:    make([]Poliklinik, 0),
		Dokter:        make([]Dokter, 0),
		SidebarPasien: make([]SidebarPasien, 0),
	}

	registrasi, err := bacaRegistrasiHariIni(ctx, tx)
	if err != nil {
		return Beranda{}, err
	}
	hasil.Registrasi = registrasi

	rawatJalan, err := bacaPasienRawatJalan(ctx, tx, filter.RawatJalan, false)
	if err != nil {
		return Beranda{}, err
	}
	hasil.RawatJalan = rawatJalan.Data
	hasil.Paginasi.RawatJalan = rawatJalan.Paginasi

	igd, err := bacaPasienRawatJalan(ctx, tx, filter.IGD, true)
	if err != nil {
		return Beranda{}, err
	}
	hasil.IGD = igd.Data
	hasil.Paginasi.IGD = igd.Paginasi

	rawatInap, err := bacaRawatInap(ctx, tx, filter.RawatInap)
	if err != nil {
		return Beranda{}, err
	}
	hasil.RawatInap = rawatInap.Data
	hasil.Paginasi.RawatInap = rawatInap.Paginasi
	hasil.Poliklinik, err = bacaPoliklinik(ctx, tx)
	if err != nil {
		return Beranda{}, err
	}
	hasil.Dokter, err = bacaDokter(ctx, tx)
	if err != nil {
		return Beranda{}, err
	}
	hasil.PilihanStatus, err = bacaPilihanStatus(ctx, tx)
	if err != nil {
		return Beranda{}, err
	}
	hasil.Ringkasan.JumlahRegistrasi = int64(len(hasil.Registrasi))
	hasil.Paginasi.Registrasi = paginasiDariTotal(1, len(hasil.Registrasi), int64(len(hasil.Registrasi)))
	hasil.Ringkasan.JumlahRawatJalan = hasil.Paginasi.RawatJalan.Total
	hasil.Ringkasan.JumlahIGD = hasil.Paginasi.IGD.Total
	hasil.Ringkasan.JumlahRawatInap = hasil.Paginasi.RawatInap.Total

	if err := tx.Commit(); err != nil {
		return Beranda{}, fmt.Errorf("commit SIMRS read-only transaction: %w", err)
	}

	hasil.SidebarPasien, err = r.bacaSidebarPasien(ctx, userID)
	if err != nil {
		return Beranda{}, err
	}
	return hasil, nil
}

func (r *Repositori) bacaSidebarPasien(ctx context.Context, userID uint64) ([]SidebarPasien, error) {
	rows, err := r.aplikasiDB.QueryContext(ctx, `
		SELECT sidebar.kode_sidebar,sidebar.nama_sidebar,sidebar.ikon,sidebar.daftar_modul
		FROM sidebar_pasien sidebar
		WHERE sidebar.aktif = TRUE
		  AND sidebar.kode_sidebar IS NOT NULL
		  AND EXISTS (
			SELECT 1
			FROM user_permissions akses
			INNER JOIN permissions izin ON izin.id=akses.permission_id
			WHERE akses.user_id=?
			  AND (izin.code='*' OR izin.code=sidebar.permission_code)
		  )
		ORDER BY sidebar.urutan,sidebar.nama_sidebar
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("baca sidebar pasien aplikasi: %w", err)
	}
	defer rows.Close()

	daftar := make([]SidebarPasien, 0)
	sudahAda := make(map[string]bool)
	for rows.Next() {
		var sidebar SidebarPasien
		var daftarModulJSON []byte
		if err := rows.Scan(&sidebar.Kode, &sidebar.Nama, &sidebar.Ikon, &daftarModulJSON); err != nil {
			return nil, fmt.Errorf("baca baris sidebar pasien: %w", err)
		}
		if sudahAda[sidebar.Kode] {
			continue
		}
		if len(daftarModulJSON) > 0 {
			if err := json.Unmarshal(daftarModulJSON, &sidebar.DaftarModul); err != nil {
				return nil, fmt.Errorf("baca daftar modul sidebar pasien: %w", err)
			}
		}
		sudahAda[sidebar.Kode] = true
		daftar = append(daftar, sidebar)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi sidebar pasien: %w", err)
	}
	return daftar, nil
}

func paginasiDariTotal(halaman int, batas int, total int64) Paginasi {
	halaman = normalisasiHalaman(halaman)
	batas = normalisasiBatas(batas)

	totalHalaman := 0
	if total > 0 {
		totalHalaman = int((total + int64(batas) - 1) / int64(batas))
	}

	return Paginasi{
		Halaman:      halaman,
		Batas:        batas,
		Total:        total,
		TotalHalaman: totalHalaman,
	}
}

func normalisasiHalaman(halaman int) int {
	if halaman < 1 {
		return 1
	}
	return halaman
}

func normalisasiBatas(batas int) int {
	if batas < 1 {
		return 100
	}
	if batas > 500 {
		return 500
	}
	return batas
}

func offsetPaginasi(halaman int, batas int) int {
	halaman = normalisasiHalaman(halaman)
	batas = normalisasiBatas(batas)
	return (halaman - 1) * batas
}

func bacaRegistrasiHariIni(ctx context.Context, tx *sql.Tx) ([]Pasien, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT
			COALESCE(r.no_rawat, ''), COALESCE(p.nm_pasien, ''),
			COALESCE(r.no_rkm_medis, ''), COALESCE(r.no_reg, ''),
			COALESCE(p.jk, ''), COALESCE(DATE_FORMAT(p.tgl_lahir, '%Y-%m-%d'), ''),
			TRIM(CONCAT(COALESCE(r.umurdaftar, ''), ' ', COALESCE(r.sttsumur, ''))),
			COALESCE(p.alamat, ''), COALESCE(p.no_tlp, ''),
			COALESCE(poly.nm_poli, r.kd_poli, ''), COALESCE(r.kd_poli, ''),
			COALESCE(d.nm_dokter, ''), COALESCE(pj.png_jawab, ''),
			COALESCE(r.stts, ''), COALESCE(DATE_FORMAT(r.tgl_registrasi, '%Y-%m-%d'), ''),
			COALESCE(TIME_FORMAT(r.jam_reg, '%H:%i'), ''), COALESCE(r.status_lanjut, ''),
			COALESCE(r.status_bayar, '')
		FROM reg_periksa r
		LEFT JOIN pasien p ON p.no_rkm_medis = r.no_rkm_medis
		LEFT JOIN poliklinik poly ON poly.kd_poli = r.kd_poli
		LEFT JOIN dokter d ON d.kd_dokter = r.kd_dokter
		LEFT JOIN penjab pj ON pj.kd_pj = r.kd_pj
		WHERE r.tgl_registrasi = CURDATE()
		ORDER BY r.jam_reg DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("read SIMRS registrations: %w", err)
	}
	defer rows.Close()

	daftarPasien := make([]Pasien, 0)
	for rows.Next() {
		var data Pasien
		if err := rows.Scan(
			&data.NoRawat, &data.NamaPasien, &data.NoRekamMedis, &data.NoRegistrasi,
			&data.JenisKelamin, &data.TanggalLahir, &data.Umur, &data.Alamat, &data.NoTelepon,
			&data.Poliklinik, &data.KodePoliklinik, &data.Dokter, &data.Penjamin,
			&data.Status, &data.TanggalRegistrasi, &data.JamRegistrasi, &data.JenisPerawatanSIMRS,
			&data.StatusBayar,
		); err != nil {
			return nil, fmt.Errorf("scan SIMRS registration: %w", err)
		}
		daftarPasien = append(daftarPasien, data)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate SIMRS registrations: %w", err)
	}
	return daftarPasien, nil
}

func bacaPasienRawatJalan(ctx context.Context, tx *sql.Tx, filter FilterPasien, igd bool) (hasilPasien, error) {
	kondisi := []string{
		"r.tgl_registrasi BETWEEN ? AND ?",
		"r.status_lanjut = 'Ralan'",
	}
	argumen := []any{filter.TanggalMulai, filter.TanggalSelesai}

	if igd {
		kondisi = append(kondisi, "r.kd_poli LIKE '%IGD%'")
	} else {
		kondisi = append(kondisi, "r.kd_poli NOT LIKE '%IGD%'")
	}
	if filter.Status != "" {
		kondisi = append(kondisi, "r.stts = ?")
		argumen = append(argumen, filter.Status)
	}
	if filter.StatusBayar != "" {
		kondisi = append(kondisi, "r.status_bayar = ?")
		argumen = append(argumen, filter.StatusBayar)
	}
	if !igd && filter.Poliklinik != "" {
		kondisi = append(kondisi, "r.kd_poli = ?")
		argumen = append(argumen, filter.Poliklinik)
	}
	if filter.Dokter != "" {
		kondisi = append(kondisi, "r.kd_dokter = ?")
		argumen = append(argumen, filter.Dokter)
	}
	if filter.Pencarian != "" {
		kondisi = append(kondisi, "(r.no_rkm_medis LIKE ? OR r.no_rawat LIKE ? OR p.nm_pasien LIKE ?)")
		kataKunci := "%" + filter.Pencarian + "%"
		argumen = append(argumen, kataKunci, kataKunci, kataKunci)
	}

	total, err := hitungPasienRawatJalan(ctx, tx, kondisi, argumen)
	if err != nil {
		return hasilPasien{}, err
	}
	paginasi := paginasiDariTotal(filter.Halaman, filter.Batas, total)
	argumenData := append([]any{}, argumen...)
	argumenData = append(argumenData, paginasi.Batas, offsetPaginasi(paginasi.Halaman, paginasi.Batas))

	rows, err := tx.QueryContext(ctx, `
		SELECT
			COALESCE(r.no_rawat, ''), COALESCE(p.nm_pasien, ''),
			COALESCE(r.no_rkm_medis, ''), COALESCE(r.no_reg, ''),
			COALESCE(p.jk, ''), COALESCE(DATE_FORMAT(p.tgl_lahir, '%Y-%m-%d'), ''),
			TRIM(CONCAT(COALESCE(r.umurdaftar, ''), ' ', COALESCE(r.sttsumur, ''))),
			COALESCE(p.alamat, ''), COALESCE(p.no_tlp, ''),
			COALESCE(poly.nm_poli, r.kd_poli, ''), COALESCE(r.kd_poli, ''),
			COALESCE(d.nm_dokter, ''), COALESCE(pj.png_jawab, ''),
			COALESCE(r.stts, ''), COALESCE(DATE_FORMAT(r.tgl_registrasi, '%Y-%m-%d'), ''),
			COALESCE(TIME_FORMAT(r.jam_reg, '%H:%i'), ''), COALESCE(r.status_lanjut, ''),
			COALESCE(r.status_bayar, ''), COALESCE(sep.no_sep, '')
		FROM reg_periksa r
		LEFT JOIN pasien p ON p.no_rkm_medis = r.no_rkm_medis
		LEFT JOIN poliklinik poly ON poly.kd_poli = r.kd_poli
		LEFT JOIN dokter d ON d.kd_dokter = r.kd_dokter
		LEFT JOIN penjab pj ON pj.kd_pj = r.kd_pj
		LEFT JOIN bridging_sep sep ON sep.no_rawat = r.no_rawat AND sep.jnspelayanan = '2'
		WHERE `+strings.Join(kondisi, " AND ")+`
		ORDER BY r.jam_reg DESC
		LIMIT ? OFFSET ?
	`, argumenData...)
	if err != nil {
		return hasilPasien{}, fmt.Errorf("read filtered SIMRS outpatients: %w", err)
	}
	defer rows.Close()

	daftarPasien := make([]Pasien, 0)
	for rows.Next() {
		var data Pasien
		if err := rows.Scan(
			&data.NoRawat, &data.NamaPasien, &data.NoRekamMedis, &data.NoRegistrasi,
			&data.JenisKelamin, &data.TanggalLahir, &data.Umur, &data.Alamat, &data.NoTelepon,
			&data.Poliklinik, &data.KodePoliklinik, &data.Dokter, &data.Penjamin,
			&data.Status, &data.TanggalRegistrasi, &data.JamRegistrasi, &data.JenisPerawatanSIMRS,
			&data.StatusBayar, &data.NoSEP,
		); err != nil {
			return hasilPasien{}, fmt.Errorf("scan filtered SIMRS outpatient: %w", err)
		}
		daftarPasien = append(daftarPasien, data)
	}
	if err := rows.Err(); err != nil {
		return hasilPasien{}, fmt.Errorf("iterate filtered SIMRS outpatients: %w", err)
	}
	return hasilPasien{Data: daftarPasien, Paginasi: paginasi}, nil
}

func hitungPasienRawatJalan(ctx context.Context, tx *sql.Tx, kondisi []string, argumen []any) (int64, error) {
	var total int64
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM reg_periksa r
		LEFT JOIN pasien p ON p.no_rkm_medis = r.no_rkm_medis
		WHERE `+strings.Join(kondisi, " AND "), argumen...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count filtered SIMRS outpatients: %w", err)
	}
	return total, nil
}

func bacaRawatInap(ctx context.Context, tx *sql.Tx, filter FilterPasien) (hasilPasien, error) {
	kondisi := make([]string, 0)
	argumen := make([]any, 0)

	if filter.BelumPulang {
		kondisi = append(kondisi,
			"ki.tgl_keluar = '0000-00-00'",
			"ki.jam_keluar = '00:00:00'",
			"ki.stts_pulang = '-'",
		)
	} else if filter.TanggalDipilihManual {
		kondisi = append(kondisi, "ki.tgl_masuk BETWEEN ? AND ?")
		argumen = append(argumen, filter.TanggalMulai, filter.TanggalSelesai)
	} else {
		kondisi = append(kondisi, "(ki.tgl_masuk BETWEEN ? AND ? OR ki.tgl_keluar = '0000-00-00')")
		argumen = append(argumen, filter.TanggalMulai, filter.TanggalSelesai)
	}
	if filter.Status != "" {
		kondisi = append(kondisi, "ki.stts_pulang = ?")
		argumen = append(argumen, filter.Status)
	}
	if filter.StatusBayar != "" {
		kondisi = append(kondisi, "r.status_bayar = ?")
		argumen = append(argumen, filter.StatusBayar)
	}
	if filter.Dokter != "" {
		kondisi = append(kondisi, "(r.kd_dokter = ? OR EXISTS (SELECT 1 FROM dpjp_ranap dp WHERE dp.no_rawat = ki.no_rawat AND dp.kd_dokter = ?))")
		argumen = append(argumen, filter.Dokter, filter.Dokter)
	}
	if filter.Pencarian != "" {
		kondisi = append(kondisi, `(
			ki.no_rawat LIKE ? OR ki.kd_kamar LIKE ? OR ki.diagnosa_awal LIKE ? OR ki.diagnosa_akhir LIKE ?
			OR r.no_rkm_medis LIKE ? OR r.p_jawab LIKE ? OR r.status_bayar LIKE ?
			OR p.nm_pasien LIKE ? OR p.alamat LIKE ? OR COALESCE(d.nm_dokter, '') LIKE ?
			OR EXISTS (
				SELECT 1
				FROM dpjp_ranap dp_cari
				JOIN dokter d_cari ON d_cari.kd_dokter = dp_cari.kd_dokter
				WHERE dp_cari.no_rawat = ki.no_rawat
				  AND d_cari.nm_dokter LIKE ?
			)
			OR COALESCE(pj.png_jawab, '') LIKE ? OR COALESCE(b.nm_bangsal, '') LIKE ?
		)`)
		kataKunci := "%" + filter.Pencarian + "%"
		argumen = append(argumen, kataKunci, kataKunci, kataKunci, kataKunci, kataKunci, kataKunci, kataKunci, kataKunci, kataKunci, kataKunci, kataKunci, kataKunci, kataKunci)
	}

	total, err := hitungRawatInap(ctx, tx, kondisi, argumen)
	if err != nil {
		return hasilPasien{}, err
	}
	paginasi := paginasiDariTotal(filter.Halaman, filter.Batas, total)
	argumenData := append([]any{}, argumen...)
	argumenData = append(argumenData, paginasi.Batas, offsetPaginasi(paginasi.Halaman, paginasi.Batas))

	rows, err := tx.QueryContext(ctx, `
		SELECT
			COALESCE(ki.no_rawat, ''), COALESCE(p.nm_pasien, ''),
			COALESCE(r.no_rkm_medis, ''), '', COALESCE(p.jk, ''),
			COALESCE(DATE_FORMAT(p.tgl_lahir, '%Y-%m-%d'), ''),
			TRIM(CONCAT(COALESCE(r.umurdaftar, ''), ' ', COALESCE(r.sttsumur, ''))),
			COALESCE(p.alamat, ''), COALESCE(p.no_tlp, ''),
			TRIM(CONCAT(COALESCE(ki.kd_kamar, ''), ' ', COALESCE(b.nm_bangsal, ''))),
			COALESCE(ki.kd_kamar, ''),
			COALESCE((
				SELECT GROUP_CONCAT(DISTINCT d_dpjp.nm_dokter ORDER BY d_dpjp.nm_dokter SEPARATOR ', ')
				FROM dpjp_ranap dp_tampil
				JOIN dokter d_dpjp ON d_dpjp.kd_dokter = dp_tampil.kd_dokter
				WHERE dp_tampil.no_rawat = ki.no_rawat
			), COALESCE(d.nm_dokter, '')),
			COALESCE(pj.png_jawab, ''), COALESCE(NULLIF(ki.stts_pulang, ''), 'Dirawat'),
			COALESCE(DATE_FORMAT(ki.tgl_masuk, '%Y-%m-%d'), ''),
			COALESCE(TIME_FORMAT(ki.jam_masuk, '%H:%i'), ''),
			COALESCE(ki.diagnosa_awal, ''), COALESCE(ki.lama, ''),
			COALESCE(r.status_bayar, ''), COALESCE(sep.no_sep, ''),
			COALESCE(sep.klsrawat, ''), COALESCE(DATE_FORMAT(sep.tglsep, '%Y-%m-%d'), '')
		FROM kamar_inap ki
		JOIN reg_periksa r ON r.no_rawat = ki.no_rawat
		LEFT JOIN pasien p ON p.no_rkm_medis = r.no_rkm_medis
		LEFT JOIN dokter d ON d.kd_dokter = r.kd_dokter
		LEFT JOIN penjab pj ON pj.kd_pj = r.kd_pj
		LEFT JOIN kamar k ON k.kd_kamar = ki.kd_kamar
		LEFT JOIN bangsal b ON b.kd_bangsal = k.kd_bangsal
		LEFT JOIN bridging_sep sep ON sep.no_rawat = ki.no_rawat AND sep.jnspelayanan = '1'
		WHERE `+strings.Join(kondisi, " AND ")+`
		ORDER BY ki.tgl_masuk, ki.jam_masuk
		LIMIT ? OFFSET ?
	`, argumenData...)
	if err != nil {
		return hasilPasien{}, fmt.Errorf("read active SIMRS inpatients: %w", err)
	}
	defer rows.Close()

	daftarPasien := make([]Pasien, 0)
	for rows.Next() {
		var data Pasien
		if err := rows.Scan(
			&data.NoRawat, &data.NamaPasien, &data.NoRekamMedis, &data.NoRegistrasi,
			&data.JenisKelamin, &data.TanggalLahir, &data.Umur, &data.Alamat, &data.NoTelepon,
			&data.Kamar, &data.KodePoliklinik, &data.Dokter, &data.Penjamin,
			&data.Status, &data.TanggalRegistrasi, &data.JamRegistrasi, &data.DiagnosaAwal,
			&data.LamaRawat, &data.StatusBayar, &data.NoSEP, &data.KelasSEP, &data.TanggalSEP,
		); err != nil {
			return hasilPasien{}, fmt.Errorf("scan active SIMRS inpatient: %w", err)
		}
		data.Poliklinik = data.Kamar
		daftarPasien = append(daftarPasien, data)
	}
	if err := rows.Err(); err != nil {
		return hasilPasien{}, fmt.Errorf("iterate active SIMRS inpatients: %w", err)
	}
	return hasilPasien{Data: daftarPasien, Paginasi: paginasi}, nil
}

func hitungRawatInap(ctx context.Context, tx *sql.Tx, kondisi []string, argumen []any) (int64, error) {
	var total int64
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM kamar_inap ki
		JOIN reg_periksa r ON r.no_rawat = ki.no_rawat
		LEFT JOIN pasien p ON p.no_rkm_medis = r.no_rkm_medis
		LEFT JOIN dokter d ON d.kd_dokter = r.kd_dokter
		LEFT JOIN penjab pj ON pj.kd_pj = r.kd_pj
		LEFT JOIN kamar k ON k.kd_kamar = ki.kd_kamar
		LEFT JOIN bangsal b ON b.kd_bangsal = k.kd_bangsal
		WHERE `+strings.Join(kondisi, " AND "), argumen...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count active SIMRS inpatients: %w", err)
	}
	return total, nil
}

func bacaPoliklinik(ctx context.Context, tx *sql.Tx) ([]Poliklinik, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT COALESCE(kd_poli, ''), COALESCE(nm_poli, '')
		FROM poliklinik
		ORDER BY nm_poli
	`)
	if err != nil {
		return nil, fmt.Errorf("read SIMRS polyclinics: %w", err)
	}
	defer rows.Close()

	daftarPoliklinik := make([]Poliklinik, 0)
	for rows.Next() {
		var data Poliklinik
		if err := rows.Scan(&data.Kode, &data.Nama); err != nil {
			return nil, fmt.Errorf("scan SIMRS polyclinic: %w", err)
		}
		daftarPoliklinik = append(daftarPoliklinik, data)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate SIMRS polyclinics: %w", err)
	}
	return daftarPoliklinik, nil
}

func bacaDokter(ctx context.Context, tx *sql.Tx) ([]Dokter, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT
			COALESCE(kd_dokter, ''), COALESCE(nm_dokter, ''),
			COALESCE(kd_sps, ''), COALESCE(status, '')
		FROM dokter
		WHERE COALESCE(kd_dokter, '') <> ''
		  AND COALESCE(nm_dokter, '') <> ''
		ORDER BY nm_dokter
	`)
	if err != nil {
		return nil, fmt.Errorf("read SIMRS doctors: %w", err)
	}
	defer rows.Close()

	daftarDokter := make([]Dokter, 0)
	for rows.Next() {
		var data Dokter
		if err := rows.Scan(&data.Kode, &data.Nama, &data.KodeSpesialis, &data.Status); err != nil {
			return nil, fmt.Errorf("scan SIMRS doctor: %w", err)
		}
		daftarDokter = append(daftarDokter, data)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate SIMRS doctors: %w", err)
	}
	return daftarDokter, nil
}

func bacaPilihanStatus(ctx context.Context, tx *sql.Tx) (PilihanStatus, error) {
	statusPeriksa, err := bacaEnumKolom(ctx, tx, "reg_periksa", "stts")
	if err != nil {
		return PilihanStatus{}, err
	}
	statusRawatInap, err := bacaEnumKolom(ctx, tx, "kamar_inap", "stts_pulang")
	if err != nil {
		return PilihanStatus{}, err
	}
	statusBayar, err := bacaEnumKolom(ctx, tx, "reg_periksa", "status_bayar")
	if err != nil {
		return PilihanStatus{}, err
	}

	return PilihanStatus{
		Periksa:     statusPeriksa,
		RawatInap:   statusRawatInap,
		StatusBayar: statusBayar,
	}, nil
}

func bacaEnumKolom(ctx context.Context, tx *sql.Tx, namaTabel string, namaKolom string) ([]string, error) {
	var tipeKolom string
	err := tx.QueryRowContext(ctx, `
		SELECT COLUMN_TYPE
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		  AND COLUMN_NAME = ?
		LIMIT 1
	`, namaTabel, namaKolom).Scan(&tipeKolom)
	if err != nil {
		return nil, fmt.Errorf("read SIMRS enum %s.%s: %w", namaTabel, namaKolom, err)
	}

	return parseEnumMySQL(tipeKolom), nil
}

func parseEnumMySQL(tipeKolom string) []string {
	if !strings.HasPrefix(tipeKolom, "enum(") || !strings.HasSuffix(tipeKolom, ")") {
		return []string{}
	}

	isi := strings.TrimSuffix(strings.TrimPrefix(tipeKolom, "enum("), ")")
	hasil := make([]string, 0)
	var builder strings.Builder
	dalamKutip := false
	escape := false

	for _, karakter := range isi {
		if escape {
			builder.WriteRune(karakter)
			escape = false
			continue
		}
		if karakter == '\\' {
			escape = true
			continue
		}
		if karakter == '\'' {
			if dalamKutip {
				hasil = append(hasil, builder.String())
				builder.Reset()
			}
			dalamKutip = !dalamKutip
			continue
		}
		if dalamKutip {
			builder.WriteRune(karakter)
		}
	}

	return hasil
}
