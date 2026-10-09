package ringkasan_pasien

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

var ErrValidasi = errors.New("nomor rawat atau modul tidak valid")
var ErrTidakAda = errors.New("kunjungan tidak ditemukan")

type Bagian struct {
	Kode   string              `json:"kode"`
	Sumber string              `json:"sumber"`
	Baris  []map[string]string `json:"baris"`
	Error  string              `json:"error,omitempty"`
}

type Hasil struct {
	NoRawat string   `json:"no_rawat"`
	Modul   string   `json:"modul"`
	Bagian  []Bagian `json:"bagian"`
}

type Repositori struct{ db *sql.DB }

func NewRepositori(db *sql.DB) *Repositori { return &Repositori{db: db} }

func Konteks(modul string) (tabel, status string, permissions []string) {
	switch modul {
	case "Rawat Inap":
		return "pemeriksaan_ranap", "Ranap", []string{"kamar_inap", "daftar_pasien_ranap"}
	case "Rawat Jalan":
		return "pemeriksaan_ralan", "Ralan", []string{"tindakan_ralan", "billing_ralan", "registrasi"}
	case "IGD/UGD":
		return "pemeriksaan_ralan", "Ralan", []string{"igd"}
	case "Registrasi":
		return "pemeriksaan_ralan", "Ralan", []string{"registrasi"}
	}
	return "", "", nil
}

func (r *Repositori) Daftar(ctx context.Context, no, modul string) (Hasil, error) {
	tabel, status, _ := Konteks(modul)
	no = strings.TrimSpace(no)
	h := Hasil{NoRawat: no, Modul: modul, Bagian: []Bagian{}}
	if no == "" || len(no) > 17 || tabel == "" {
		return h, ErrValidasi
	}
	kunjungan, err := r.baca(ctx, `SELECT rp.no_rawat, rp.no_rkm_medis no_rm,
		DATE_FORMAT(rp.tgl_registrasi,'%Y-%m-%d') tanggal_registrasi,
		TIME_FORMAT(rp.jam_reg,'%H:%i:%s') jam_registrasi,
		rp.no_reg no_registrasi, rp.status_lanjut, rp.stts status_kunjungan, rp.status_bayar,
		pj.png_jawab penjamin, p.nm_poli poliklinik, d.nm_dokter dokter_registrasi
		FROM reg_periksa rp LEFT JOIN penjab pj ON pj.kd_pj=rp.kd_pj
		LEFT JOIN poliklinik p ON p.kd_poli=rp.kd_poli
		LEFT JOIN dokter d ON d.kd_dokter=rp.kd_dokter WHERE rp.no_rawat=?`, no)
	if err != nil {
		return h, err
	}
	if len(kunjungan) == 0 {
		return h, ErrTidakAda
	}
	h.Bagian = append(h.Bagian, Bagian{Kode: "kunjungan", Sumber: "Registrasi kunjungan SIMRS", Baris: kunjungan})
	tambah := func(kode, sumber, query string, args ...any) {
		b := Bagian{Kode: kode, Sumber: sumber, Baris: []map[string]string{}}
		rows, err := r.baca(ctx, query, args...)
		if err != nil {
			b.Error = "Sumber belum dapat dibaca. Periksa koneksi, tabel, dan izin SELECT SIMRS."
		} else {
			b.Baris = rows
		}
		h.Bagian = append(h.Bagian, b)
	}
	tambah("sep", "SEP pada kunjungan ini", `SELECT no_sep FROM bridging_sep WHERE no_rawat=? ORDER BY no_sep`, no)
	if status == "Ranap" {
		tambah("dokter", "Seluruh DPJP rawat inap pada kunjungan ini", `SELECT dp.kd_dokter kode, d.nm_dokter nama FROM dpjp_ranap dp LEFT JOIN dokter d ON d.kd_dokter=dp.kd_dokter WHERE dp.no_rawat=? ORDER BY d.nm_dokter,dp.kd_dokter`, no)
		tambah("kamar", "Riwayat kamar paling akhir (bukan selalu kamar aktif)", `SELECT ki.kd_kamar kode, b.nm_bangsal bangsal, DATE_FORMAT(ki.tgl_masuk,'%Y-%m-%d') tanggal_masuk, TIME_FORMAT(ki.jam_masuk,'%H:%i:%s') jam_masuk, ki.stts_pulang status_pulang FROM kamar_inap ki LEFT JOIN kamar k ON k.kd_kamar=ki.kd_kamar LEFT JOIN bangsal b ON b.kd_bangsal=k.kd_bangsal WHERE ki.no_rawat=? ORDER BY ki.tgl_masuk DESC,ki.jam_masuk DESC,ki.kd_kamar LIMIT 1`, no)
	} else {
		tambah("dokter", "Dokter registrasi kunjungan", `SELECT rp.kd_dokter kode, d.nm_dokter nama FROM reg_periksa rp LEFT JOIN dokter d ON d.kd_dokter=rp.kd_dokter WHERE rp.no_rawat=?`, no)
	}
	tambah("diagnosis", "Diagnosis terkode pada layanan yang dibuka", `SELECT dp.kd_penyakit kode, p.nm_penyakit nama, CAST(dp.prioritas AS CHAR) prioritas FROM diagnosa_pasien dp LEFT JOIN penyakit p ON p.kd_penyakit=dp.kd_penyakit WHERE dp.no_rawat=? AND dp.status=? ORDER BY dp.prioritas,dp.kd_penyakit`, no, status)
	// Tabel dipilih dari daftar tetap, bukan dari teks query pengguna.
	waktu := `DATE_FORMAT(TIMESTAMP(p.tgl_perawatan,p.jam_rawat),'%Y-%m-%d %H:%i:%s') waktu, p.nip petugas, pg.nama nama_petugas`
	tambah("vital", "CPPT / "+tabel+" — catatan paling akhir", `SELECT `+waktu+`, p.tensi, p.nadi, p.respirasi, p.suhu_tubuh suhu, p.spo2, p.gcs, p.kesadaran FROM `+tabel+` p LEFT JOIN pegawai pg ON pg.nik=p.nip WHERE p.no_rawat=? AND TIMESTAMP(p.tgl_perawatan,p.jam_rawat)=(SELECT MAX(TIMESTAMP(tgl_perawatan,jam_rawat)) FROM `+tabel+` WHERE no_rawat=?) ORDER BY p.nip`, no, no)
	tambah("alergi", "Catatan alergi CPPT pada kunjungan dan layanan ini", `SELECT `+waktu+`, p.alergi FROM `+tabel+` p LEFT JOIN pegawai pg ON pg.nik=p.nip WHERE p.no_rawat=? AND TRIM(COALESCE(p.alergi,'')) NOT IN ('','-') ORDER BY p.tgl_perawatan DESC,p.jam_rawat DESC,p.nip`, no)
	tambah("cppt", "Jumlah catatan CPPT pada layanan ini", `SELECT CAST(COUNT(*) AS CHAR) jumlah FROM `+tabel+` WHERE no_rawat=?`, no)
	resume := "resume_pasien"
	if status == "Ranap" {
		resume = "resume_pasien_ranap"
	}
	tambah("resume", "Keberadaan resume, belum memeriksa isian atau validasi", `SELECT CAST(COUNT(*) AS CHAR) jumlah FROM `+resume+` WHERE no_rawat=?`, no)
	tambah("berkas", "Tautan berkas digital tercatat, bukan verifikasi isi atau akses file", `SELECT CAST(COUNT(*) AS CHAR) jumlah FROM berkas_digital_perawatan WHERE no_rawat=? AND TRIM(COALESCE(lokasi_file,'')) NOT IN ('','-')`, no)
	if status == "Ranap" {
		tambah("edukasi", "Catatan edukasi pada kunjungan ini; bukan penilaian kelengkapan materi", `SELECT CAST(COUNT(*) AS CHAR) jumlah FROM catatan_edukasi WHERE no_rawat=?`, no)
		tambah("pemulangan", "Rencana pemulangan tercatat; bukan konfirmasi pasien boleh pulang", `SELECT CAST(COUNT(*) AS CHAR) jumlah FROM perencanaan_pemulangan WHERE no_rawat=?`, no)
	}
	for _, sumber := range sumberObat {
		tambah(sumber.kode, sumber.nama, sumber.query, no)
	}
	// Satu baris per nomor permintaan, tanpa join detail yang menggandakan jumlah.
	// Nama tabel berasal dari daftar tetap. Hasil kunjungan tidak dipasangkan
	// otomatis ke order karena tabel hasil tidak menyimpan nomor order.
	for _, layanan := range []string{"lab", "radiologi"} {
		sumber := "Permintaan radiologi pada nomor rawat ini, lintas layanan"
		if layanan == "lab" {
			sumber = "Permintaan laboratorium PK pada nomor rawat ini, lintas layanan; belum termasuk PA/MB"
		}
		tambah("permintaan_"+layanan, sumber, `SELECT p.noorder,
			DATE_FORMAT(p.tgl_permintaan,'%Y-%m-%d') tanggal_permintaan,
			TIME_FORMAT(p.jam_permintaan,'%H:%i:%s') jam_permintaan,
			DATE_FORMAT(p.tgl_sampel,'%Y-%m-%d') tanggal_proses,
			TIME_FORMAT(p.jam_sampel,'%H:%i:%s') jam_proses,
			DATE_FORMAT(p.tgl_hasil,'%Y-%m-%d') tanggal_hasil,
			TIME_FORMAT(p.jam_hasil,'%H:%i:%s') jam_hasil,
			p.status layanan, COALESCE(d.nm_dokter,p.dokter_perujuk) dokter
			FROM permintaan_`+layanan+` p
			LEFT JOIN dokter d ON d.kd_dokter=p.dokter_perujuk
			WHERE p.no_rawat=?
			ORDER BY p.tgl_permintaan DESC,p.jam_permintaan DESC,p.noorder DESC`, no)
		for _, baris := range h.Bagian[len(h.Bagian)-1].Baris {
			statusPermintaan(baris, layanan)
		}
	}
	tambah("lab", "Hasil laboratorium pada nomor rawat ini, lintas layanan; bukan status penyelesaian seluruh permintaan", `SELECT DATE_FORMAT(TIMESTAMP(d.tgl_periksa,d.jam),'%Y-%m-%d %H:%i:%s') waktu, d.kd_jenis_prw kode_pemeriksaan, CAST(d.id_template AS CHAR) id_template, t.Pemeriksaan pemeriksaan, d.nilai, t.satuan, d.nilai_rujukan, d.keterangan FROM detail_periksa_lab d LEFT JOIN template_laboratorium t ON t.id_template=d.id_template WHERE d.no_rawat=? AND TRIM(COALESCE(d.nilai,'')) NOT IN ('','-') ORDER BY d.tgl_periksa DESC,d.jam DESC,d.kd_jenis_prw,d.id_template`, no)
	tambah("radiologi", "Laporan teks radiologi pada nomor rawat ini, lintas layanan; bukan gambar atau status seluruh permintaan", `SELECT DATE_FORMAT(TIMESTAMP(tgl_periksa,jam),'%Y-%m-%d %H:%i:%s') waktu, hasil FROM hasil_radiologi WHERE no_rawat=? AND TRIM(COALESCE(hasil,'')) NOT IN ('','-') ORDER BY tgl_periksa DESC,jam DESC`, no)
	return h, nil
}

func statusPermintaan(baris map[string]string, layanan string) {
	// Tanggal nol/NULL bukan bukti proses. Jam 00:00:00 sah bila tanggal valid.
	for _, tahap := range []string{"permintaan", "proses", "hasil"} {
		tanggal := baris["tanggal_"+tahap]
		waktu, err := time.Parse("2006-01-02", tanggal)
		if err != nil || waktu.Year() < 1 {
			baris["waktu_"+tahap] = ""
			continue
		}
		baris["waktu_"+tahap] = strings.TrimSpace(tanggal + " " + baris["jam_"+tahap])
	}
	baris["tahap"] = "menunggu"
	baris["status_progres"] = "Menunggu proses"
	if baris["waktu_proses"] != "" {
		baris["tahap"] = "diproses"
		baris["status_progres"] = "Diterima / diproses"
		if layanan == "lab" {
			baris["status_progres"] = "Sampel diterima"
		}
	}
	if baris["waktu_hasil"] != "" {
		baris["tahap"] = "hasil_dicatat"
		baris["status_progres"] = "Tanggal hasil tercatat"
	}
}

func (r *Repositori) baca(ctx context.Context, query string, args ...any) ([]map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	out := []map[string]string{}
	for rows.Next() {
		values := make([]sql.NullString, len(cols))
		dest := make([]any, len(cols))
		for i := range values {
			dest[i] = &values[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		row := map[string]string{}
		for i, k := range cols {
			row[k] = values[i].String
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
