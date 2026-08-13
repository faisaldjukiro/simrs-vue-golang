package riwayat_perawatan

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
)

type sumberDokumen struct {
	Tabel string
	Jenis string
}

// sumberDokumenKhanza mengikuti bagian yang dipanggil oleh tampilPerawatan()
// pada DlgRawatInap.java. Daftar ini sengaja berupa whitelist: nama tabel tidak
// pernah berasal dari request pengguna dan seluruh query hanya SELECT.
var sumberDokumenKhanza = []sumberDokumen{
	{"rujukan_internal_poli", "Rujukan Internal Poli"},
	{"data_triase_igd", "Triase IGD"},
	{"data_triase_igdprimer", "Triase IGD Primer"},
	{"data_triase_igdsekunder", "Triase IGD Sekunder"},
	{"data_triase_igddetail_skala1", "Detail Triase Skala 1"},
	{"data_triase_igddetail_skala2", "Detail Triase Skala 2"},
	{"data_triase_igddetail_skala3", "Detail Triase Skala 3"},
	{"data_triase_igddetail_skala4", "Detail Triase Skala 4"},
	{"data_triase_igddetail_skala5", "Detail Triase Skala 5"},
	{"penilaian_awal_keperawatan_igd", "Asesmen Awal Keperawatan IGD"},
	{"penilaian_awal_keperawatan_igd_masalah", "Masalah Keperawatan IGD"},
	{"penilaian_awal_keperawatan_ralan_rencana_igd", "Rencana Keperawatan IGD"},
	{"penilaian_awal_keperawatan_ralan", "Asesmen Awal Keperawatan Rawat Jalan"},
	{"penilaian_awal_keperawatan_ralan_masalah", "Masalah Keperawatan Rawat Jalan"},
	{"penilaian_awal_keperawatan_ralan_rencana", "Rencana Keperawatan Rawat Jalan"},
	{"penilaian_awal_keperawatan_gigi", "Asesmen Awal Keperawatan Gigi"},
	{"penilaian_awal_keperawatan_gigi_masalah", "Masalah Keperawatan Gigi"},
	{"penilaian_awal_keperawatan_ralan_rencana_gigi", "Rencana Keperawatan Gigi"},
	{"penilaian_awal_keperawatan_ralan_bayi", "Asesmen Awal Keperawatan Anak/Bayi"},
	{"penilaian_awal_keperawatan_ralan_bayi_masalah", "Masalah Keperawatan Anak/Bayi"},
	{"penilaian_awal_keperawatan_ralan_rencana_anak", "Rencana Keperawatan Anak/Bayi"},
	{"penilaian_awal_keperawatan_kebidanan", "Asesmen Awal Keperawatan Kebidanan"},
	{"penilaian_awal_keperawatan_ralan_psikiatri", "Asesmen Awal Keperawatan Psikiatri"},
	{"penilaian_awal_keperawatan_ralan_masalah_psikiatri", "Masalah Keperawatan Psikiatri"},
	{"penilaian_awal_keperawatan_ralan_rencana_psikiatri", "Rencana Keperawatan Psikiatri"},
	{"penilaian_fisioterapi", "Asesmen Fisioterapi"},
	{"penilaian_psikologi", "Asesmen Psikologi"},
	{"penilaian_medis_igd", "Asesmen Awal Medis IGD"},
	{"penilaian_medis_ralan", "Asesmen Awal Medis Rawat Jalan"},
	{"penilaian_medis_ralan_kandungan", "Asesmen Medis Kandungan"},
	{"penilaian_medis_ralan_anak", "Asesmen Medis Anak/Bayi"},
	{"penilaian_medis_ralan_tht", "Asesmen Medis THT"},
	{"penilaian_medis_ralan_psikiatrik", "Asesmen Medis Psikiatrik"},
	{"penilaian_medis_ralan_penyakit_dalam", "Asesmen Medis Penyakit Dalam"},
	{"penilaian_medis_ralan_mata", "Asesmen Medis Mata"},
	{"penilaian_medis_ralan_neurologi", "Asesmen Medis Neurologi"},
	{"penilaian_medis_ralan_orthopedi", "Asesmen Medis Orthopedi"},
	{"penilaian_medis_ralan_bedah", "Asesmen Medis Bedah"},
	{"penilaian_medis_ralan_geriatri", "Asesmen Medis Geriatri"},
	{"uji_fungsi_kfr", "Uji Fungsi KFR"},
	{"hemodialisa", "Hemodialisa"},
	{"penilaian_awal_keperawatan_ranap", "Asesmen Awal Keperawatan Rawat Inap"},
	{"penilaian_awal_keperawatan_ranap_masalah", "Masalah Keperawatan Rawat Inap"},
	{"penilaian_awal_keperawatan_ranap_rencana", "Rencana Keperawatan Rawat Inap"},
	{"penilaian_awal_keperawatan_kebidanan_ranap", "Asesmen Kebidanan Rawat Inap"},
	{"penilaian_medis_ranap", "Asesmen Awal Medis Rawat Inap"},
	{"penilaian_medis_ranap_kandungan", "Asesmen Medis Rawat Inap Kebidanan"},
	{"perencanaan_pemulangan", "Perencanaan Pemulangan"},
	{"catatan_observasi_igd", "Catatan Observasi IGD"},
	{"catatan_cek_gds", "Catatan Pemeriksaan GDS"},
	{"catatan_observasi_ranap", "Catatan Observasi Rawat Inap"},
	{"catatan_observasi_ranap_kebidanan", "Catatan Observasi Kebidanan"},
	{"catatan_observasi_ranap_postpartum", "Catatan Observasi Postpartum"},
	{"catatan_keperawatan_ranap", "Catatan Keperawatan Rawat Inap"},
	{"pemeriksaan_obstetri_ralan", "Pemeriksaan Obstetri Rawat Jalan"},
	{"pemeriksaan_ginekologi_ralan", "Pemeriksaan Ginekologi Rawat Jalan"},
	{"pemeriksaan_obstetri_ranap", "Pemeriksaan Obstetri Rawat Inap"},
	{"pemeriksaan_ginekologi_ranap", "Pemeriksaan Ginekologi Rawat Inap"},
	{"penilaian_lanjutan_resiko_jatuh_dewasa", "Penilaian Risiko Jatuh Dewasa"},
	{"penilaian_lanjutan_resiko_jatuh_anak", "Penilaian Risiko Jatuh Anak"},
	{"penilaian_tambahan_geriatri", "Penilaian Tambahan Geriatri"},
	{"pemantauan_pews_anak", "Pemantauan EWS/PEWS"},
	{"checklist_pre_operasi", "Checklist Pre Operasi"},
	{"signin_sebelum_anestesi", "Sign In Sebelum Anestesi"},
	{"timeout_sebelum_insisi", "Time Out Sebelum Insisi"},
	{"signout_sebelum_menutup_luka", "Sign Out Sebelum Menutup Luka"},
	{"checklist_post_operasi", "Checklist Post Operasi"},
	{"penilaian_pre_operasi", "Asesmen Pre Operasi"},
	{"penilaian_pre_anestesi", "Asesmen Pre Anestesi"},
	{"hasil_pemeriksaan_usg", "Hasil Pemeriksaan USG"},
	{"skrining_nutrisi_dewasa", "Skrining Nutrisi Dewasa"},
	{"skrining_nutrisi_lansia", "Skrining Nutrisi Lansia"},
	{"skrining_nutrisi_anak", "Skrining Nutrisi Anak"},
	{"skrining_gizi", "Skrining Gizi"},
	{"asuhan_gizi", "Asuhan Gizi"},
	{"monitoring_asuhan_gizi", "Monitoring Asuhan Gizi"},
	{"konseling_farmasi", "Konseling Farmasi"},
	{"pelayanan_informasi_obat", "Pelayanan Informasi Obat"},
	{"transfer_pasien_antar_ruang", "Transfer Pasien Antar Ruang"},
	{"aturan_pakai", "Aturan Pakai Obat"},
	{"laporan_operasi", "Laporan Operasi"},
	{"permintaan_labpa", "Permintaan Patologi Anatomi"},
	{"saran_kesan_lab", "Saran dan Kesan Laboratorium"},
	{"resume_pasien_ranap", "Resume Pasien Rawat Inap"},
}

func (r *Repositori) isiDPJP(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	rows, err := r.simrsDB.QueryContext(ctx, `SELECT x.no_rawat,COALESCE(x.kd_dokter,''),COALESCE(d.nm_dokter,''),COALESCE(x.urutan,0) FROM dpjp_ranap x LEFT JOIN dokter d ON d.kd_dokter=x.kd_dokter WHERE x.no_rawat IN (`+ph+`) ORDER BY x.no_rawat,x.urutan`, args...)
	if err != nil {
		return fmt.Errorf("baca DPJP riwayat: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var noRawat, kodeDokter, dokter string
		var urutan int
		if err := rows.Scan(&noRawat, &kodeDokter, &dokter, &urutan); err != nil {
			return err
		}
		kunjungan := &data.Kunjungan[indeks[noRawat]]
		kunjungan.DPJP = append(kunjungan.DPJP, dokter)
		kunjungan.TandaTangan = append(kunjungan.TandaTangan, r.tandaTangan(kodeDokter, dokter, "Dokter DPJP", urutan))
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i := range data.Kunjungan {
		kunjungan := &data.Kunjungan[i]
		if len(kunjungan.TandaTangan) > 0 || strings.TrimSpace(kunjungan.KodeDokter) == "" {
			continue
		}
		peran := "Dokter Poli"
		if kunjungan.StatusRawat == "Ranap" {
			peran = "Dokter DPJP"
		}
		kunjungan.TandaTangan = append(kunjungan.TandaTangan, r.tandaTangan(kunjungan.KodeDokter, kunjungan.Dokter, peran, 1))
	}
	return nil
}

func (r *Repositori) tandaTangan(kodeDokter, dokter, peran string, urutan int) TandaTangan {
	kodeDokter = strings.TrimSpace(kodeDokter)
	item := TandaTangan{KodeDokter: kodeDokter, Dokter: dokter, Peran: peran, Urutan: urutan}
	if kodeDokter == "" || r.webBaseURL == "" {
		return item
	}
	item.URLQRCode = r.urlBerkas("penggajian/temp", kodeDokter+".png")
	item.URLGenerator = r.urlBerkas("penggajian", "generateqrcode.php") + "?kodedokter=" + url.QueryEscape(strings.ReplaceAll(kodeDokter, " ", "_"))
	return item
}

func (r *Repositori) isiCatatanDokter(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	rows, err := r.simrsDB.QueryContext(ctx, `SELECT x.no_rawat,DATE_FORMAT(x.tanggal,'%Y-%m-%d'),TIME_FORMAT(x.jam,'%H:%i:%s'),COALESCE(d.nm_dokter,''),COALESCE(x.catatan,'') FROM catatan_perawatan x LEFT JOIN dokter d ON d.kd_dokter=x.kd_dokter WHERE x.no_rawat IN (`+ph+`) ORDER BY x.tanggal DESC,x.jam DESC`, args...)
	if err != nil {
		return fmt.Errorf("baca catatan dokter: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var noRawat string
		var item CatatanDokter
		if err := rows.Scan(&noRawat, &item.Tanggal, &item.Jam, &item.Dokter, &item.Catatan); err != nil {
			return err
		}
		data.Kunjungan[indeks[noRawat]].CatatanDokter = append(data.Kunjungan[indeks[noRawat]].CatatanDokter, item)
	}
	return rows.Err()
}

func (r *Repositori) isiDetailPenunjang(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	if err := r.isiHasilRadiologi(ctx, data, indeks, ph, args); err != nil {
		return err
	}
	if err := r.isiDetailLaboratorium(ctx, data, indeks, ph, args); err != nil {
		return err
	}
	return r.isiObatOperasi(ctx, data, indeks, ph, args)
}

func (r *Repositori) isiHasilRadiologi(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	rows, err := r.simrsDB.QueryContext(ctx, `SELECT no_rawat,DATE_FORMAT(tgl_periksa,'%Y-%m-%d'),TIME_FORMAT(jam,'%H:%i:%s'),COALESCE(hasil,'') FROM hasil_radiologi WHERE no_rawat IN (`+ph+`)`, args...)
	if err != nil {
		return fmt.Errorf("baca hasil radiologi: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var noRawat, tanggal, jam, hasil string
		if err := rows.Scan(&noRawat, &tanggal, &jam, &hasil); err != nil {
			return err
		}
		kunjungan := &data.Kunjungan[indeks[noRawat]]
		for i := range kunjungan.Radiologi {
			if kunjungan.Radiologi[i].Tanggal == tanggal && kunjungan.Radiologi[i].Jam == jam {
				kunjungan.Radiologi[i].Hasil = hasil
			}
		}
	}
	return rows.Err()
}

func (r *Repositori) isiDetailLaboratorium(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	query := `SELECT x.no_rawat,x.kd_jenis_prw,DATE_FORMAT(x.tgl_periksa,'%Y-%m-%d'),TIME_FORMAT(x.jam,'%H:%i:%s'),COALESCE(t.Pemeriksaan,''),COALESCE(x.nilai,''),COALESCE(t.satuan,''),COALESCE(x.nilai_rujukan,''),COALESCE(x.keterangan,''),COALESCE(x.biaya_item,0) FROM detail_periksa_lab x LEFT JOIN template_laboratorium t ON t.id_template=x.id_template WHERE x.no_rawat IN (` + ph + `) ORDER BY x.no_rawat,x.tgl_periksa,x.jam,t.urut`
	rows, err := r.simrsDB.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("baca detail hasil laboratorium: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var noRawat, kode, tanggal, jam string
		var item DetailLaboratorium
		if err := rows.Scan(&noRawat, &kode, &tanggal, &jam, &item.Pemeriksaan, &item.Nilai, &item.Satuan, &item.NilaiRujukan, &item.Keterangan, &item.Biaya); err != nil {
			return err
		}
		kunjungan := &data.Kunjungan[indeks[noRawat]]
		for i := range kunjungan.Laboratorium {
			if kunjungan.Laboratorium[i].Kode == kode && kunjungan.Laboratorium[i].Tanggal == tanggal && kunjungan.Laboratorium[i].Jam == jam {
				kunjungan.Laboratorium[i].Detail = append(kunjungan.Laboratorium[i].Detail, item)
				break
			}
		}
	}
	return rows.Err()
}

func (r *Repositori) isiObatOperasi(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	query := `SELECT x.no_rawat,x.kd_obat,COALESCE(o.nm_obat,''),DATE_FORMAT(x.tanggal,'%Y-%m-%d'),TIME_FORMAT(x.tanggal,'%H:%i:%s'),COALESCE(x.jumlah,0),COALESCE(o.kode_sat,''),COALESCE(x.hargasatuan,0)*COALESCE(x.jumlah,0) FROM beri_obat_operasi x LEFT JOIN obatbhp_ok o ON o.kd_obat=x.kd_obat WHERE x.no_rawat IN (` + ph + `) ORDER BY x.tanggal DESC`
	rows, err := r.simrsDB.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("baca obat dan BHP operasi: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var noRawat string
		item := Obat{Jenis: "Obat Operasi"}
		if err := rows.Scan(&noRawat, &item.Kode, &item.Nama, &item.Tanggal, &item.Jam, &item.Jumlah, &item.Satuan, &item.Total); err != nil {
			return err
		}
		data.Kunjungan[indeks[noRawat]].Obat = append(data.Kunjungan[indeks[noRawat]].Obat, item)
	}
	return rows.Err()
}

func (r *Repositori) isiDokumenKlinis(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	tersedia, err := r.tabelDokumenTersedia(ctx)
	if err != nil {
		return err
	}
	for _, sumber := range sumberDokumenKhanza {
		if !tersedia[sumber.Tabel] {
			continue
		}
		query := "SELECT * FROM `" + sumber.Tabel + "` WHERE no_rawat IN (" + ph + ")"
		rows, err := r.simrsDB.QueryContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("baca %s: %w", sumber.Jenis, err)
		}
		if err := tambahkanDokumen(rows, sumber, data, indeks); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
	}
	return nil
}

func (r *Repositori) tabelDokumenTersedia(ctx context.Context) (map[string]bool, error) {
	nama := make([]any, 0, len(sumberDokumenKhanza))
	for _, sumber := range sumberDokumenKhanza {
		nama = append(nama, sumber.Tabel)
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(nama)), ",")
	rows, err := r.simrsDB.QueryContext(ctx, `SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name IN (`+ph+`)`, nama...)
	if err != nil {
		return nil, fmt.Errorf("baca daftar tabel dokumen klinis: %w", err)
	}
	defer rows.Close()
	tersedia := make(map[string]bool, len(nama))
	for rows.Next() {
		var tabel string
		if err := rows.Scan(&tabel); err != nil {
			return nil, err
		}
		tersedia[tabel] = true
	}
	return tersedia, rows.Err()
}

func tambahkanDokumen(rows *sql.Rows, sumber sumberDokumen, data *Data, indeks map[string]int) error {
	kolom, err := rows.Columns()
	if err != nil {
		return err
	}
	for rows.Next() {
		nilai := make([]sql.RawBytes, len(kolom))
		tujuan := make([]any, len(kolom))
		for i := range nilai {
			tujuan[i] = &nilai[i]
		}
		if err := rows.Scan(tujuan...); err != nil {
			return err
		}
		noRawat := ""
		baris := make(map[string]string, len(kolom)-1)
		for i, namaKolom := range kolom {
			isi := string(nilai[i])
			if strings.EqualFold(namaKolom, "no_rawat") {
				noRawat = isi
				continue
			}
			baris[namaKolom] = isi
		}
		posisi, ada := indeks[noRawat]
		if !ada {
			continue
		}
		kunjungan := &data.Kunjungan[posisi]
		posisiDokumen := -1
		for i := range kunjungan.DokumenKlinis {
			if kunjungan.DokumenKlinis[i].Tabel == sumber.Tabel {
				posisiDokumen = i
				break
			}
		}
		if posisiDokumen == -1 {
			kunjungan.DokumenKlinis = append(kunjungan.DokumenKlinis, DokumenKlinis{Jenis: sumber.Jenis, Tabel: sumber.Tabel, Data: make([]map[string]string, 0, 1)})
			posisiDokumen = len(kunjungan.DokumenKlinis) - 1
		}
		kunjungan.DokumenKlinis[posisiDokumen].Data = append(kunjungan.DokumenKlinis[posisiDokumen].Data, baris)
	}
	return rows.Err()
}

func (r *Repositori) isiBerkasDigital(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any) error {
	query := `SELECT b.no_rawat,COALESCE(m.nama,''),COALESCE(b.lokasi_file,'')
		FROM berkas_digital_perawatan b
		LEFT JOIN master_berkas_digital m ON m.kode=b.kode
		WHERE b.no_rawat IN (` + ph + `) ORDER BY b.no_rawat,m.nama`
	rows, err := r.simrsDB.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("baca berkas digital perawatan: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var noRawat string
		var item BerkasDigital
		if err := rows.Scan(&noRawat, &item.Nama, &item.LokasiFile); err != nil {
			return err
		}
		item.Jenis = "Berkas Digital Perawatan"
		item.URL = r.urlBerkas("berkasrawat", item.LokasiFile)
		data.Kunjungan[indeks[noRawat]].BerkasDigital = append(data.Kunjungan[indeks[noRawat]].BerkasDigital, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if err := r.isiGambarSederhana(ctx, data, indeks, ph, args,
		`SELECT no_rawat,COALESCE(photo,'') FROM hasil_pemeriksaan_usg_gambar WHERE no_rawat IN (`+ph+`)`,
		"Gambar USG", "hasilpemeriksaanusg"); err != nil {
		return err
	}
	return r.isiGambarSederhana(ctx, data, indeks, ph, args,
		`SELECT no_rawat,COALESCE(photo,'') FROM detail_periksa_labpa_gambar WHERE no_rawat IN (`+ph+`)`,
		"Gambar Patologi Anatomi", "labpa")
}

func (r *Repositori) isiGambarSederhana(ctx context.Context, data *Data, indeks map[string]int, ph string, args []any, query, jenis, folder string) error {
	rows, err := r.simrsDB.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("baca %s: %w", strings.ToLower(jenis), err)
	}
	defer rows.Close()
	for rows.Next() {
		var noRawat, lokasi string
		if err := rows.Scan(&noRawat, &lokasi); err != nil {
			return err
		}
		if strings.TrimSpace(lokasi) == "" || strings.TrimSpace(lokasi) == "-" {
			continue
		}
		item := BerkasDigital{Jenis: jenis, Nama: jenis, LokasiFile: lokasi, URL: r.urlBerkas(folder, lokasi)}
		data.Kunjungan[indeks[noRawat]].BerkasDigital = append(data.Kunjungan[indeks[noRawat]].BerkasDigital, item)
	}
	return rows.Err()
}
