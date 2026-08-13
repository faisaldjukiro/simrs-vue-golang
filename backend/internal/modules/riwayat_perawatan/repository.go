package riwayat_perawatan

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"path"
	"strings"
)

type Filter struct {
	NoRawat        string
	Mode           string
	TanggalMulai   string
	TanggalSelesai string
	NomorRawat     string
}

type Diagnosis struct {
	Kode      string `json:"kode"`
	Nama      string `json:"nama"`
	Prioritas int    `json:"prioritas"`
	Status    string `json:"status"`
}

type Prosedur struct {
	Kode      string `json:"kode"`
	Nama      string `json:"nama"`
	Prioritas int    `json:"prioritas"`
}

type Catatan struct {
	Jenis     string `json:"jenis"`
	Tanggal   string `json:"tanggal"`
	Jam       string `json:"jam"`
	Petugas   string `json:"petugas"`
	Jabatan   string `json:"jabatan"`
	Kesadaran string `json:"kesadaran"`
	Tensi     string `json:"tensi"`
	Suhu      string `json:"suhu"`
	Nadi      string `json:"nadi"`
	Respirasi string `json:"respirasi"`
	SpO2      string `json:"spo2"`
	Subjek    string `json:"subjek"`
	Objek     string `json:"objek"`
	Asesmen   string `json:"asesmen"`
	Plan      string `json:"plan"`
	Instruksi string `json:"instruksi"`
	Evaluasi  string `json:"evaluasi"`
}

type PemeriksaanPenunjang struct {
	Kode    string               `json:"kode"`
	Nama    string               `json:"nama"`
	Tanggal string               `json:"tanggal"`
	Jam     string               `json:"jam"`
	Petugas string               `json:"petugas"`
	Dokter  string               `json:"dokter"`
	Biaya   float64              `json:"biaya"`
	Hasil   string               `json:"hasil,omitempty"`
	Detail  []DetailLaboratorium `json:"detail,omitempty"`
}

type DetailLaboratorium struct {
	Pemeriksaan  string  `json:"pemeriksaan"`
	Nilai        string  `json:"nilai"`
	Satuan       string  `json:"satuan"`
	NilaiRujukan string  `json:"nilai_rujukan"`
	Keterangan   string  `json:"keterangan"`
	Biaya        float64 `json:"biaya"`
}

type CatatanDokter struct {
	Tanggal string `json:"tanggal"`
	Jam     string `json:"jam"`
	Dokter  string `json:"dokter"`
	Catatan string `json:"catatan"`
}

type DokumenKlinis struct {
	Jenis string              `json:"jenis"`
	Tabel string              `json:"tabel"`
	Data  []map[string]string `json:"data"`
}

type BerkasDigital struct {
	Jenis      string `json:"jenis"`
	Nama       string `json:"nama"`
	LokasiFile string `json:"lokasi_file"`
	URL        string `json:"url"`
}

type TandaTangan struct {
	KodeDokter   string `json:"kode_dokter"`
	Dokter       string `json:"dokter"`
	Peran        string `json:"peran"`
	Urutan       int    `json:"urutan"`
	URLQRCode    string `json:"url_qr_code"`
	URLGenerator string `json:"url_generator"`
}

type Obat struct {
	Jenis   string  `json:"jenis"`
	Kode    string  `json:"kode"`
	Nama    string  `json:"nama"`
	Tanggal string  `json:"tanggal"`
	Jam     string  `json:"jam"`
	Jumlah  float64 `json:"jumlah"`
	Satuan  string  `json:"satuan"`
	Total   float64 `json:"total"`
}

type Tindakan struct {
	Jenis     string  `json:"jenis"`
	Kode      string  `json:"kode"`
	Nama      string  `json:"nama"`
	Tanggal   string  `json:"tanggal"`
	Jam       string  `json:"jam"`
	Pelaksana string  `json:"pelaksana"`
	Biaya     float64 `json:"biaya"`
}

type PenggunaanKamar struct {
	KodeKamar     string  `json:"kode_kamar"`
	Bangsal       string  `json:"bangsal"`
	TanggalMasuk  string  `json:"tanggal_masuk"`
	JamMasuk      string  `json:"jam_masuk"`
	TanggalKeluar string  `json:"tanggal_keluar"`
	JamKeluar     string  `json:"jam_keluar"`
	Lama          float64 `json:"lama"`
	StatusPulang  string  `json:"status_pulang"`
	Total         float64 `json:"total"`
}

type Operasi struct {
	Kode     string  `json:"kode"`
	Nama     string  `json:"nama"`
	Tanggal  string  `json:"tanggal"`
	Anestesi string  `json:"anestesi"`
	Operator string  `json:"operator"`
	Total    float64 `json:"total"`
}

type BiayaLain struct {
	Jenis string  `json:"jenis"`
	Nama  string  `json:"nama"`
	Total float64 `json:"total"`
}

type ResepPulang struct {
	Kode   string  `json:"kode"`
	Nama   string  `json:"nama"`
	Dosis  string  `json:"dosis"`
	Jumlah float64 `json:"jumlah"`
	Satuan string  `json:"satuan"`
	Total  float64 `json:"total"`
}

type Resume struct {
	Dokter               string `json:"dokter"`
	KondisiPulang        string `json:"kondisi_pulang"`
	KeluhanUtama         string `json:"keluhan_utama"`
	JalannyaPenyakit     string `json:"jalannya_penyakit"`
	PemeriksaanPenunjang string `json:"pemeriksaan_penunjang"`
	HasilLaboratorium    string `json:"hasil_laboratorium"`
	DiagnosisUtama       string `json:"diagnosis_utama"`
	DiagnosisSekunder    string `json:"diagnosis_sekunder"`
	ProsedurUtama        string `json:"prosedur_utama"`
	ProsedurSekunder     string `json:"prosedur_sekunder"`
	ObatPulang           string `json:"obat_pulang"`
}

type Kunjungan struct {
	NoRawat         string                 `json:"no_rawat"`
	NoRegistrasi    string                 `json:"no_registrasi"`
	Tanggal         string                 `json:"tanggal"`
	Jam             string                 `json:"jam"`
	StatusRawat     string                 `json:"status_rawat"`
	StatusPeriksa   string                 `json:"status_periksa"`
	StatusBayar     string                 `json:"status_bayar"`
	Dokter          string                 `json:"dokter"`
	KodeDokter      string                 `json:"-"`
	Poliklinik      string                 `json:"poliklinik"`
	Ruangan         string                 `json:"ruangan"`
	Penjamin        string                 `json:"penjamin"`
	NoSEP           string                 `json:"no_sep"`
	KelasRawat      string                 `json:"kelas_rawat"`
	PenanggungJawab string                 `json:"penanggung_jawab"`
	AlamatPJ        string                 `json:"alamat_penanggung_jawab"`
	HubunganPJ      string                 `json:"hubungan_penanggung_jawab"`
	DPJP            []string               `json:"dpjp"`
	Diagnosis       []Diagnosis            `json:"diagnosis"`
	Prosedur        []Prosedur             `json:"prosedur"`
	Catatan         []Catatan              `json:"catatan"`
	Radiologi       []PemeriksaanPenunjang `json:"radiologi"`
	Laboratorium    []PemeriksaanPenunjang `json:"laboratorium"`
	Obat            []Obat                 `json:"obat"`
	Tindakan        []Tindakan             `json:"tindakan"`
	Kamar           []PenggunaanKamar      `json:"kamar"`
	Operasi         []Operasi              `json:"operasi"`
	BiayaLain       []BiayaLain            `json:"biaya_lain"`
	ResepPulang     []ResepPulang          `json:"resep_pulang"`
	Resume          *Resume                `json:"resume,omitempty"`
	CatatanDokter   []CatatanDokter        `json:"catatan_dokter"`
	DokumenKlinis   []DokumenKlinis        `json:"dokumen_klinis"`
	BerkasDigital   []BerkasDigital        `json:"berkas_digital"`
	TandaTangan     []TandaTangan          `json:"tanda_tangan"`
}

type Data struct {
	NoRekamMedis string      `json:"no_rekam_medis"`
	NamaPasien   string      `json:"nama_pasien"`
	Mode         string      `json:"mode"`
	Jumlah       int         `json:"jumlah"`
	Kunjungan    []Kunjungan `json:"kunjungan"`
}

type Repositori struct {
	simrsDB    *sql.DB
	webBaseURL string
}

func NewRepositori(simrsDB *sql.DB, webBaseURL ...string) *Repositori {
	alamat := ""
	if len(webBaseURL) > 0 {
		alamat = strings.TrimRight(strings.TrimSpace(webBaseURL[0]), "/")
	}
	return &Repositori{simrsDB: simrsDB, webBaseURL: alamat}
}

func (r *Repositori) Data(ctx context.Context, filter Filter) (Data, error) {
	data := Data{Mode: filter.Mode, Kunjungan: make([]Kunjungan, 0)}
	if err := r.simrsDB.QueryRowContext(ctx, `
		SELECT rp.no_rkm_medis, COALESCE(p.nm_pasien,'')
		FROM reg_periksa rp INNER JOIN pasien p ON p.no_rkm_medis=rp.no_rkm_medis
		WHERE rp.no_rawat=? LIMIT 1
	`, filter.NoRawat).Scan(&data.NoRekamMedis, &data.NamaPasien); err != nil {
		return data, fmt.Errorf("baca identitas riwayat perawatan: %w", err)
	}

	query, args := queryKunjungan(data.NoRekamMedis, filter)
	rows, err := r.simrsDB.QueryContext(ctx, query, args...)
	if err != nil {
		return data, fmt.Errorf("baca kunjungan pasien: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Kunjungan
		if err := rows.Scan(&item.NoRegistrasi, &item.NoRawat, &item.Tanggal, &item.Jam,
			&item.KodeDokter, &item.Dokter, &item.Poliklinik, &item.StatusRawat, &item.StatusPeriksa,
			&item.StatusBayar, &item.Penjamin, &item.Ruangan, &item.PenanggungJawab,
			&item.AlamatPJ, &item.HubunganPJ, &item.NoSEP, &item.KelasRawat); err != nil {
			return data, fmt.Errorf("scan kunjungan pasien: %w", err)
		}
		item.Diagnosis = make([]Diagnosis, 0)
		item.Prosedur = make([]Prosedur, 0)
		item.Catatan = make([]Catatan, 0)
		item.Radiologi = make([]PemeriksaanPenunjang, 0)
		item.Laboratorium = make([]PemeriksaanPenunjang, 0)
		item.Obat = make([]Obat, 0)
		item.Tindakan = make([]Tindakan, 0)
		item.Kamar = make([]PenggunaanKamar, 0)
		item.Operasi = make([]Operasi, 0)
		item.BiayaLain = make([]BiayaLain, 0)
		item.ResepPulang = make([]ResepPulang, 0)
		item.DPJP = make([]string, 0)
		item.CatatanDokter = make([]CatatanDokter, 0)
		item.DokumenKlinis = make([]DokumenKlinis, 0)
		item.BerkasDigital = make([]BerkasDigital, 0)
		item.TandaTangan = make([]TandaTangan, 0)
		data.Kunjungan = append(data.Kunjungan, item)
	}
	if err := rows.Err(); err != nil {
		return data, fmt.Errorf("iterasi kunjungan pasien: %w", err)
	}
	data.Jumlah = len(data.Kunjungan)
	if data.Jumlah == 0 {
		return data, nil
	}

	indeks, nomorRawat, argsRawat := indeksKunjungan(data.Kunjungan)
	placeholder := strings.TrimSuffix(strings.Repeat("?,", len(nomorRawat)), ",")
	if err := r.isiDiagnosis(ctx, &data, indeks, placeholder, argsRawat); err != nil {
		return data, err
	}
	if err := r.isiProsedur(ctx, &data, indeks, placeholder, argsRawat); err != nil {
		return data, err
	}
	if err := r.isiCatatan(ctx, &data, indeks, placeholder, argsRawat); err != nil {
		return data, err
	}
	if err := r.isiPenunjang(ctx, &data, indeks, placeholder, argsRawat, true); err != nil {
		return data, err
	}
	if err := r.isiPenunjang(ctx, &data, indeks, placeholder, argsRawat, false); err != nil {
		return data, err
	}
	if err := r.isiObat(ctx, &data, indeks, placeholder, argsRawat); err != nil {
		return data, err
	}
	if err := r.isiSemuaTindakan(ctx, &data, indeks, placeholder, argsRawat); err != nil {
		return data, err
	}
	if err := r.isiKamar(ctx, &data, indeks, placeholder, argsRawat); err != nil {
		return data, err
	}
	if err := r.isiOperasi(ctx, &data, indeks, placeholder, argsRawat); err != nil {
		return data, err
	}
	if err := r.isiBiayaLain(ctx, &data, indeks, placeholder, argsRawat); err != nil {
		return data, err
	}
	if err := r.isiResepPulang(ctx, &data, indeks, placeholder, argsRawat); err != nil {
		return data, err
	}
	if err := r.isiResume(ctx, &data, indeks, placeholder, argsRawat); err != nil {
		return data, err
	}
	if err := r.isiDPJP(ctx, &data, indeks, placeholder, argsRawat); err != nil {
		return data, err
	}
	if err := r.isiCatatanDokter(ctx, &data, indeks, placeholder, argsRawat); err != nil {
		return data, err
	}
	if err := r.isiDetailPenunjang(ctx, &data, indeks, placeholder, argsRawat); err != nil {
		return data, err
	}
	if err := r.isiDokumenKlinis(ctx, &data, indeks, placeholder, argsRawat); err != nil {
		return data, err
	}
	if err := r.isiBerkasDigital(ctx, &data, indeks, placeholder, argsRawat); err != nil {
		return data, err
	}
	return data, nil
}

func (r *Repositori) urlBerkas(folder, lokasi string) string {
	lokasi = strings.TrimSpace(lokasi)
	if lokasi == "" || lokasi == "-" || r.webBaseURL == "" {
		return ""
	}
	base, err := url.Parse(r.webBaseURL)
	if err != nil {
		return ""
	}
	base.Path = path.Join(base.Path, folder, strings.TrimLeft(lokasi, "/"))
	return base.String()
}

func queryKunjungan(noRM string, filter Filter) (string, []any) {
	query := `
		SELECT COALESCE(rp.no_reg,''), rp.no_rawat, DATE_FORMAT(rp.tgl_registrasi,'%Y-%m-%d'),
			TIME_FORMAT(rp.jam_reg,'%H:%i:%s'), COALESCE(rp.kd_dokter,''), COALESCE(d.nm_dokter,''), COALESCE(po.nm_poli,''),
			rp.status_lanjut, COALESCE(rp.stts,''), COALESCE(rp.status_bayar,''), COALESCE(pj.png_jawab,''),
			COALESCE((SELECT GROUP_CONCAT(DISTINCT b.nm_bangsal ORDER BY ki.tgl_masuk,ki.jam_masuk SEPARATOR ', ')
				FROM kamar_inap ki INNER JOIN kamar k ON k.kd_kamar=ki.kd_kamar
				INNER JOIN bangsal b ON b.kd_bangsal=k.kd_bangsal WHERE ki.no_rawat=rp.no_rawat),''),
			COALESCE(rp.p_jawab,''),COALESCE(rp.almt_pj,''),COALESCE(rp.hubunganpj,''),
			COALESCE((SELECT bs.no_sep FROM bridging_sep bs WHERE bs.no_rawat=rp.no_rawat
				ORDER BY bs.jnspelayanan='1' DESC,bs.tglsep DESC,bs.no_sep DESC LIMIT 1),''),
			COALESCE((SELECT bs.klsrawat FROM bridging_sep bs WHERE bs.no_rawat=rp.no_rawat
				ORDER BY bs.jnspelayanan='1' DESC,bs.tglsep DESC,bs.no_sep DESC LIMIT 1),'')
		FROM reg_periksa rp
		LEFT JOIN dokter d ON d.kd_dokter=rp.kd_dokter
		LEFT JOIN poliklinik po ON po.kd_poli=rp.kd_poli
		LEFT JOIN penjab pj ON pj.kd_pj=rp.kd_pj
		WHERE rp.stts<>'Batal' AND rp.no_rkm_medis=?`
	args := []any{noRM}
	switch filter.Mode {
	case "tanggal":
		query += " AND rp.tgl_registrasi BETWEEN ? AND ? ORDER BY rp.tgl_registrasi DESC,rp.jam_reg DESC"
		args = append(args, filter.TanggalMulai, filter.TanggalSelesai)
	case "nomor":
		query += " AND rp.no_rawat=? ORDER BY rp.tgl_registrasi DESC,rp.jam_reg DESC"
		args = append(args, filter.NomorRawat)
	case "semua":
		query += " ORDER BY rp.tgl_registrasi DESC,rp.jam_reg DESC"
	default:
		query += " ORDER BY rp.tgl_registrasi DESC,rp.jam_reg DESC LIMIT 5"
	}
	return query, args
}

func indeksKunjungan(kunjungan []Kunjungan) (map[string]int, []string, []any) {
	indeks := make(map[string]int, len(kunjungan))
	nomor := make([]string, 0, len(kunjungan))
	args := make([]any, 0, len(kunjungan))
	for i, item := range kunjungan {
		indeks[item.NoRawat] = i
		nomor = append(nomor, item.NoRawat)
		args = append(args, item.NoRawat)
	}
	return indeks, nomor, args
}

func (r *Repositori) isiDiagnosis(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	rows, err := r.simrsDB.QueryContext(ctx, `SELECT dp.no_rawat,dp.kd_penyakit,COALESCE(p.nm_penyakit,''),dp.prioritas,COALESCE(dp.status_penyakit,'') FROM diagnosa_pasien dp LEFT JOIN penyakit p ON p.kd_penyakit=dp.kd_penyakit WHERE dp.no_rawat IN (`+ph+`) ORDER BY dp.no_rawat,dp.prioritas`, args...)
	if err != nil {
		return fmt.Errorf("baca diagnosis riwayat: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var no string
		var item Diagnosis
		if err := rows.Scan(&no, &item.Kode, &item.Nama, &item.Prioritas, &item.Status); err != nil {
			return err
		}
		data.Kunjungan[indeks[no]].Diagnosis = append(data.Kunjungan[indeks[no]].Diagnosis, item)
	}
	return rows.Err()
}

func (r *Repositori) isiProsedur(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	rows, err := r.simrsDB.QueryContext(ctx, `SELECT pp.no_rawat,pp.kode,COALESCE(i.deskripsi_panjang,''),pp.prioritas FROM prosedur_pasien pp LEFT JOIN icd9 i ON i.kode=pp.kode WHERE pp.no_rawat IN (`+ph+`) ORDER BY pp.no_rawat,pp.prioritas`, args...)
	if err != nil {
		return fmt.Errorf("baca prosedur riwayat: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var no string
		var item Prosedur
		if err := rows.Scan(&no, &item.Kode, &item.Nama, &item.Prioritas); err != nil {
			return err
		}
		data.Kunjungan[indeks[no]].Prosedur = append(data.Kunjungan[indeks[no]].Prosedur, item)
	}
	return rows.Err()
}

func (r *Repositori) isiCatatan(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	bagian := func(tabel, jenis string) string {
		return `SELECT pr.no_rawat,'` + jenis + `',DATE_FORMAT(pr.tgl_perawatan,'%Y-%m-%d') tanggal,TIME_FORMAT(pr.jam_rawat,'%H:%i:%s') jam,COALESCE(pg.nama,''),COALESCE(pg.jbtn,''),COALESCE(pr.kesadaran,''),COALESCE(pr.tensi,''),COALESCE(pr.suhu_tubuh,''),COALESCE(pr.nadi,''),COALESCE(pr.respirasi,''),COALESCE(pr.spo2,''),COALESCE(pr.keluhan,''),COALESCE(pr.pemeriksaan,''),COALESCE(pr.penilaian,''),COALESCE(pr.rtl,''),COALESCE(pr.instruksi,''),COALESCE(pr.evaluasi,'') FROM ` + tabel + ` pr LEFT JOIN pegawai pg ON pg.nik=pr.nip WHERE pr.no_rawat IN (` + ph + `)`
	}
	query := bagian("pemeriksaan_ranap", "Rawat Inap") + ` UNION ALL ` + bagian("pemeriksaan_ralan", "Rawat Jalan") + ` ORDER BY tanggal DESC,jam DESC`
	duaArgs := append(append([]any{}, args...), args...)
	rows, err := r.simrsDB.QueryContext(ctx, query, duaArgs...)
	if err != nil {
		return fmt.Errorf("baca CPPT riwayat: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var no string
		var item Catatan
		if err := rows.Scan(&no, &item.Jenis, &item.Tanggal, &item.Jam, &item.Petugas, &item.Jabatan, &item.Kesadaran, &item.Tensi, &item.Suhu, &item.Nadi, &item.Respirasi, &item.SpO2, &item.Subjek, &item.Objek, &item.Asesmen, &item.Plan, &item.Instruksi, &item.Evaluasi); err != nil {
			return err
		}
		data.Kunjungan[indeks[no]].Catatan = append(data.Kunjungan[indeks[no]].Catatan, item)
	}
	return rows.Err()
}

func (r *Repositori) isiPenunjang(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any, radiologi bool) error {
	tabel, master, tanggal := "periksa_lab", "jns_perawatan_lab", "tgl_periksa"
	if radiologi {
		tabel, master = "periksa_radiologi", "jns_perawatan_radiologi"
	}
	query := `SELECT x.no_rawat,x.kd_jenis_prw,COALESCE(j.nm_perawatan,''),DATE_FORMAT(x.` + tanggal + `,'%Y-%m-%d'),TIME_FORMAT(x.jam,'%H:%i:%s'),COALESCE(pt.nama,''),COALESCE(d.nm_dokter,''),COALESCE(x.biaya,0) FROM ` + tabel + ` x LEFT JOIN ` + master + ` j ON j.kd_jenis_prw=x.kd_jenis_prw LEFT JOIN petugas pt ON pt.nip=x.nip LEFT JOIN dokter d ON d.kd_dokter=x.kd_dokter WHERE x.no_rawat IN (` + ph + `) ORDER BY x.` + tanggal + ` DESC,x.jam DESC`
	rows, err := r.simrsDB.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("baca pemeriksaan penunjang riwayat: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var no string
		var item PemeriksaanPenunjang
		if err := rows.Scan(&no, &item.Kode, &item.Nama, &item.Tanggal, &item.Jam, &item.Petugas, &item.Dokter, &item.Biaya); err != nil {
			return err
		}
		if radiologi {
			data.Kunjungan[indeks[no]].Radiologi = append(data.Kunjungan[indeks[no]].Radiologi, item)
		} else {
			data.Kunjungan[indeks[no]].Laboratorium = append(data.Kunjungan[indeks[no]].Laboratorium, item)
		}
	}
	return rows.Err()
}

func (r *Repositori) isiObat(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	rows, err := r.simrsDB.QueryContext(ctx, `SELECT o.no_rawat,o.kode_brng,COALESCE(d.nama_brng,''),DATE_FORMAT(o.tgl_perawatan,'%Y-%m-%d'),TIME_FORMAT(o.jam,'%H:%i:%s'),COALESCE(o.jml,0),COALESCE(d.kode_sat,''),COALESCE(o.total,0) FROM detail_pemberian_obat o LEFT JOIN databarang d ON d.kode_brng=o.kode_brng WHERE o.no_rawat IN (`+ph+`) ORDER BY o.tgl_perawatan DESC,o.jam DESC`, args...)
	if err != nil {
		return fmt.Errorf("baca pemberian obat riwayat: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var no string
		item := Obat{Jenis: "Pemberian Obat"}
		if err := rows.Scan(&no, &item.Kode, &item.Nama, &item.Tanggal, &item.Jam, &item.Jumlah, &item.Satuan, &item.Total); err != nil {
			return err
		}
		data.Kunjungan[indeks[no]].Obat = append(data.Kunjungan[indeks[no]].Obat, item)
	}
	return rows.Err()
}

type sumberTindakan struct{ tabel, master, jenis, pelaksana, joinPelaksana string }

func (r *Repositori) isiSemuaTindakan(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	sumber := []sumberTindakan{
		{"rawat_jl_dr", "jns_perawatan", "Rawat Jalan Dokter", "COALESCE(d.nm_dokter,'')", "LEFT JOIN dokter d ON d.kd_dokter=x.kd_dokter"},
		{"rawat_jl_pr", "jns_perawatan", "Rawat Jalan Petugas", "COALESCE(pt.nama,'')", "LEFT JOIN petugas pt ON pt.nip=x.nip"},
		{"rawat_jl_drpr", "jns_perawatan", "Rawat Jalan Dokter & Petugas", "CONCAT_WS(' / ',NULLIF(d.nm_dokter,''),NULLIF(pt.nama,''))", "LEFT JOIN dokter d ON d.kd_dokter=x.kd_dokter LEFT JOIN petugas pt ON pt.nip=x.nip"},
		{"rawat_inap_dr", "jns_perawatan_inap", "Rawat Inap Dokter", "COALESCE(d.nm_dokter,'')", "LEFT JOIN dokter d ON d.kd_dokter=x.kd_dokter"},
		{"rawat_inap_pr", "jns_perawatan_inap", "Rawat Inap Petugas", "COALESCE(pt.nama,'')", "LEFT JOIN petugas pt ON pt.nip=x.nip"},
		{"rawat_inap_drpr", "jns_perawatan_inap", "Rawat Inap Dokter & Petugas", "CONCAT_WS(' / ',NULLIF(d.nm_dokter,''),NULLIF(pt.nama,''))", "LEFT JOIN dokter d ON d.kd_dokter=x.kd_dokter LEFT JOIN petugas pt ON pt.nip=x.nip"},
	}
	for _, s := range sumber {
		query := `SELECT x.no_rawat,x.kd_jenis_prw,COALESCE(j.nm_perawatan,''),DATE_FORMAT(x.tgl_perawatan,'%Y-%m-%d'),TIME_FORMAT(x.jam_rawat,'%H:%i:%s'),` + s.pelaksana + `,COALESCE(x.biaya_rawat,0) FROM ` + s.tabel + ` x LEFT JOIN ` + s.master + ` j ON j.kd_jenis_prw=x.kd_jenis_prw ` + s.joinPelaksana + ` WHERE x.no_rawat IN (` + ph + `) ORDER BY x.tgl_perawatan DESC,x.jam_rawat DESC`
		rows, err := r.simrsDB.QueryContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("baca tindakan %s: %w", s.jenis, err)
		}
		for rows.Next() {
			var no string
			item := Tindakan{Jenis: s.jenis}
			if err := rows.Scan(&no, &item.Kode, &item.Nama, &item.Tanggal, &item.Jam, &item.Pelaksana, &item.Biaya); err != nil {
				rows.Close()
				return err
			}
			data.Kunjungan[indeks[no]].Tindakan = append(data.Kunjungan[indeks[no]].Tindakan, item)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
	}
	return nil
}

func (r *Repositori) isiKamar(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	rows, err := r.simrsDB.QueryContext(ctx, `SELECT ki.no_rawat,ki.kd_kamar,COALESCE(b.nm_bangsal,''),DATE_FORMAT(ki.tgl_masuk,'%Y-%m-%d'),TIME_FORMAT(ki.jam_masuk,'%H:%i:%s'),IF(ki.tgl_keluar='0000-00-00','',DATE_FORMAT(ki.tgl_keluar,'%Y-%m-%d')),IF(ki.jam_keluar='00:00:00','',TIME_FORMAT(ki.jam_keluar,'%H:%i:%s')),COALESCE(ki.lama,0),COALESCE(ki.stts_pulang,''),COALESCE(ki.ttl_biaya,0) FROM kamar_inap ki LEFT JOIN kamar k ON k.kd_kamar=ki.kd_kamar LEFT JOIN bangsal b ON b.kd_bangsal=k.kd_bangsal WHERE ki.no_rawat IN (`+ph+`) ORDER BY ki.tgl_masuk DESC,ki.jam_masuk DESC`, args...)
	if err != nil {
		return fmt.Errorf("baca penggunaan kamar: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var no string
		var item PenggunaanKamar
		if err := rows.Scan(&no, &item.KodeKamar, &item.Bangsal, &item.TanggalMasuk, &item.JamMasuk, &item.TanggalKeluar, &item.JamKeluar, &item.Lama, &item.StatusPulang, &item.Total); err != nil {
			return err
		}
		data.Kunjungan[indeks[no]].Kamar = append(data.Kunjungan[indeks[no]].Kamar, item)
	}
	return rows.Err()
}

func (r *Repositori) isiOperasi(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	rows, err := r.simrsDB.QueryContext(ctx, `SELECT o.no_rawat,o.kode_paket,COALESCE(p.nm_perawatan,''),DATE_FORMAT(o.tgl_operasi,'%Y-%m-%d %H:%i:%s'),COALESCE(o.jenis_anasthesi,''),COALESCE(d.nm_dokter,''),COALESCE(o.biayaoperator1,0)+COALESCE(o.biayaoperator2,0)+COALESCE(o.biayaoperator3,0)+COALESCE(o.biayaasisten_operator1,0)+COALESCE(o.biayaasisten_operator2,0)+COALESCE(o.biayaasisten_operator3,0)+COALESCE(o.biayainstrumen,0)+COALESCE(o.biayadokter_anak,0)+COALESCE(o.biayaperawaat_resusitas,0)+COALESCE(o.biayadokter_anestesi,0)+COALESCE(o.biayaasisten_anestesi,0)+COALESCE(o.biayaasisten_anestesi2,0)+COALESCE(o.biayabidan,0)+COALESCE(o.biayabidan2,0)+COALESCE(o.biayabidan3,0)+COALESCE(o.biayaperawat_luar,0)+COALESCE(o.biayaalat,0)+COALESCE(o.biayasewaok,0)+COALESCE(o.akomodasi,0)+COALESCE(o.bagian_rs,0)+COALESCE(o.biaya_omloop,0)+COALESCE(o.biaya_omloop2,0)+COALESCE(o.biaya_omloop3,0)+COALESCE(o.biaya_omloop4,0)+COALESCE(o.biaya_omloop5,0)+COALESCE(o.biayasarpras,0)+COALESCE(o.biaya_dokter_pjanak,0)+COALESCE(o.biaya_dokter_umum,0) FROM operasi o LEFT JOIN paket_operasi p ON p.kode_paket=o.kode_paket LEFT JOIN dokter d ON d.kd_dokter=o.operator1 WHERE o.no_rawat IN (`+ph+`) ORDER BY o.tgl_operasi DESC`, args...)
	if err != nil {
		return fmt.Errorf("baca operasi: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var no string
		var item Operasi
		if err := rows.Scan(&no, &item.Kode, &item.Nama, &item.Tanggal, &item.Anestesi, &item.Operator, &item.Total); err != nil {
			return err
		}
		data.Kunjungan[indeks[no]].Operasi = append(data.Kunjungan[indeks[no]].Operasi, item)
	}
	return rows.Err()
}

func (r *Repositori) isiBiayaLain(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	query := `SELECT no_rawat,'Tambahan',COALESCE(nama_biaya,''),COALESCE(besar_biaya,0) FROM tambahan_biaya WHERE no_rawat IN (` + ph + `) UNION ALL SELECT no_rawat,'Potongan',COALESCE(nama_pengurangan,''),-COALESCE(besar_pengurangan,0) FROM pengurangan_biaya WHERE no_rawat IN (` + ph + `)`
	dua := append(append([]any{}, args...), args...)
	rows, err := r.simrsDB.QueryContext(ctx, query, dua...)
	if err != nil {
		return fmt.Errorf("baca biaya tambahan dan potongan: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var no string
		var item BiayaLain
		if err := rows.Scan(&no, &item.Jenis, &item.Nama, &item.Total); err != nil {
			return err
		}
		data.Kunjungan[indeks[no]].BiayaLain = append(data.Kunjungan[indeks[no]].BiayaLain, item)
	}
	return rows.Err()
}

func (r *Repositori) isiResepPulang(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	rows, err := r.simrsDB.QueryContext(ctx, `SELECT r.no_rawat,r.kode_brng,COALESCE(d.nama_brng,''),COALESCE(r.dosis,''),COALESCE(r.jml_barang,0),COALESCE(d.kode_sat,''),COALESCE(r.total,0) FROM resep_pulang r LEFT JOIN databarang d ON d.kode_brng=r.kode_brng WHERE r.no_rawat IN (`+ph+`) ORDER BY d.nama_brng`, args...)
	if err != nil {
		return fmt.Errorf("baca resep pulang: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var no string
		var item ResepPulang
		if err := rows.Scan(&no, &item.Kode, &item.Nama, &item.Dosis, &item.Jumlah, &item.Satuan, &item.Total); err != nil {
			return err
		}
		data.Kunjungan[indeks[no]].ResepPulang = append(data.Kunjungan[indeks[no]].ResepPulang, item)
	}
	return rows.Err()
}

func (r *Repositori) isiResume(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	rows, err := r.simrsDB.QueryContext(ctx, `SELECT r.no_rawat,COALESCE(d.nm_dokter,''),COALESCE(r.kondisi_pulang,''),COALESCE(r.keluhan_utama,''),COALESCE(r.jalannya_penyakit,''),COALESCE(r.pemeriksaan_penunjang,''),COALESCE(r.hasil_laborat,''),CONCAT_WS(' - ',NULLIF(r.kd_diagnosa_utama,''),NULLIF(r.diagnosa_utama,'')),CONCAT_WS('; ',NULLIF(CONCAT_WS(' - ',NULLIF(r.kd_diagnosa_sekunder,''),NULLIF(r.diagnosa_sekunder,'')),''),NULLIF(CONCAT_WS(' - ',NULLIF(r.kd_diagnosa_sekunder2,''),NULLIF(r.diagnosa_sekunder2,'')),''),NULLIF(CONCAT_WS(' - ',NULLIF(r.kd_diagnosa_sekunder3,''),NULLIF(r.diagnosa_sekunder3,'')),''),NULLIF(CONCAT_WS(' - ',NULLIF(r.kd_diagnosa_sekunder4,''),NULLIF(r.diagnosa_sekunder4,'')),'')),CONCAT_WS(' - ',NULLIF(r.kd_prosedur_utama,''),NULLIF(r.prosedur_utama,'')),CONCAT_WS('; ',NULLIF(CONCAT_WS(' - ',NULLIF(r.kd_prosedur_sekunder,''),NULLIF(r.prosedur_sekunder,'')),''),NULLIF(CONCAT_WS(' - ',NULLIF(r.kd_prosedur_sekunder2,''),NULLIF(r.prosedur_sekunder2,'')),''),NULLIF(CONCAT_WS(' - ',NULLIF(r.kd_prosedur_sekunder3,''),NULLIF(r.prosedur_sekunder3,'')),'')),COALESCE(r.obat_pulang,'') FROM resume_pasien r LEFT JOIN dokter d ON d.kd_dokter=r.kd_dokter WHERE r.no_rawat IN (`+ph+`)`, args...)
	if err != nil {
		return fmt.Errorf("baca resume pasien: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var no string
		var item Resume
		if err := rows.Scan(&no, &item.Dokter, &item.KondisiPulang, &item.KeluhanUtama, &item.JalannyaPenyakit, &item.PemeriksaanPenunjang, &item.HasilLaboratorium, &item.DiagnosisUtama, &item.DiagnosisSekunder, &item.ProsedurUtama, &item.ProsedurSekunder, &item.ObatPulang); err != nil {
			return err
		}
		data.Kunjungan[indeks[no]].Resume = &item
	}
	return rows.Err()
}
