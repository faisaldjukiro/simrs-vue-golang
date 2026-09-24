package jadwal_operasi

import "strconv"

// Kontrak statis disalin dari field simpan/edit dan pilihan dialog Java.
// Jangan membangkitkan form tulis dari skema database yang tidak dibatasi.
type BidangOperasi struct {
	Kode    string   `json:"kode"`
	Label   string   `json:"label"`
	Jenis   string   `json:"jenis"`
	Batas   int      `json:"batas"`
	Wajib   bool     `json:"wajib"`
	Pilihan []string `json:"pilihan,omitempty"`
	NilaiKe string   `json:"nilai_ke,omitempty"`
}

func teks(k, label string, batas int, wajib bool) BidangOperasi {
	return BidangOperasi{Kode: k, Label: label, Jenis: "text", Batas: batas, Wajib: wajib}
}
func pilih(k, label string, opsi ...string) BidangOperasi {
	b := teks(k, label, 200, true)
	b.Jenis = "select"
	b.Pilihan = opsi
	return b
}
func tanggal(k, label, jenis string) BidangOperasi {
	b := teks(k, label, 19, true)
	b.Jenis = jenis
	return b
}
func referensi(k, label, jenis string) BidangOperasi {
	b := teks(k, label, 20, true)
	b.Jenis = jenis
	return b
}
func dasarChecklist() []BidangOperasi {
	return []BidangOperasi{
		tanggal("tanggal", "Tanggal / Jam Pencatatan (WITA)", "datetime-local"),
		teks("sncn", "SN/CN", 25, true), teks("tindakan", "Tindakan", 50, true),
		referensi("kd_dokter_bedah", "Dokter Bedah", "dokter"),
		referensi("kd_dokter_anestesi", "Dokter Anestesi", "dokter"),
	}
}

func FormPendukung(jenis string) []BidangOperasi {
	b := dasarChecklist()
	ya := []string{"Ya", "Tidak"}
	ada := []string{"Ada", "Tidak Ada"}
	area := []string{"Ada", "Tidak Ada", "Tidak Diperlukan"}
	switch jenis {
	case "penilaian_pre_operasi", "penilaian_pre_anestesi", "penilaian_pre_induksi":
		return formPenilaian(jenis)
	case "transfer_pasien_antar_ruang":
		return formTransfer()
	case "signin_sebelum_anestesi":
		b = append(b,
			pilih("identitas", "Konfirmasi Identitas", ya...),
			pilih("penandaan_area_operasi", "Penandaan Area Operasi", area...),
			teks("alergi", "Alergi", 30, false),
			pilih("resiko_aspirasi", "Risiko Aspirasi", ada...),
			teks("resiko_aspirasi_rencana_antisipasi", "Rencana Antisipasi Aspirasi", 50, false),
			pilih("resiko_kehilangan_darah", "Risiko Kehilangan Darah", "Tidak Ada", "Ada"),
			teks("resiko_kehilangan_darah_line", "Jalur IV Line", 30, false),
			teks("resiko_kehilangan_darah_rencana_antisipasi", "Antisipasi Kehilangan Darah", 50, false),
			pilih("kesiapan_alat_obat_anestesi", "Kesiapan Alat / Obat Anestesi", "Lengkap", "Pulsa Oximetri", "Tidak Lengkap"),
			teks("kesiapan_alat_obat_anestesi_rencana_antisipasi", "Antisipasi Kesiapan Alat", 50, false),
		)
	case "timeout_sebelum_insisi":
		b = append(b,
			pilih("verbal_identitas", "Konfirmasi Verbal Identitas", ya...),
			pilih("verbal_tindakan", "Konfirmasi Verbal Tindakan", ya...),
			pilih("verbal_area_insisi", "Konfirmasi Verbal Area Insisi", ya...),
			pilih("penandaan_area_operasi", "Penandaan Area Operasi", area...),
			teks("lama_operasi", "Perkiraan Lama Operasi", 10, false),
		)
		for _, k := range []string{"radiologi", "ctscan", "mri"} {
			b = append(b, pilih("penayangan_"+k, "Penayangan "+k, "Ditayangkan", "Benar", "Tidak Diperlukan"))
		}
		b = append(b,
			pilih("antibiotik_profilaks", "Pemberian Antibiotik Profilaksis", ya...),
			teks("nama_antibiotik", "Nama Antibiotik", 50, false),
			teks("jam_pemberian", "Jam Pemberian Antibiotik", 10, false),
			teks("antisipasi_kehilangan_darah", "Antisipasi Kehilangan Darah", 50, false),
			pilih("hal_khusus", "Ada Hal Khusus", ada...),
			teks("hal_khusus_diperhatikan", "Hal Khusus yang Diperhatikan", 100, false),
			tanggal("tanggal_steril", "Tanggal Steril", "date"),
			pilih("petujuk_sterilisasi", "Petunjuk Sterilisasi", ya...),
			pilih("verifikasi_preoperatif", "Verifikasi Preoperatif", ya...),
		)
	case "signout_sebelum_menutup_luka":
		for _, f := range [][2]string{{"verbal_tindakan", "Konfirmasi Verbal Tindakan"}, {"verbal_kelengkapan_kasa", "Kelengkapan Kasa"}, {"verbal_instrumen", "Kelengkapan Instrumen"}, {"verbal_alat_tajam", "Kelengkapan Alat Tajam"}} {
			b = append(b, pilih(f[0], f[1], ya...))
		}
		b = append(b,
			pilih("kelengkapan_specimen_label", "Label Spesimen", "Lengkap", "Tidak Lengkap", "Tidak Ada Pemeriksaan Spesimen"),
			pilih("kelengkapan_specimen_formulir", "Formulir Spesimen", "Lengkap", "Tidak Lengkap", "Tidak Ada Pemeriksaan Spesimen"),
			pilih("peninjauan_kegiatan_dokter_bedah", "Peninjauan Dokter Bedah", ya...),
			pilih("peninjauan_kegiatan_dokter_anestesi", "Peninjauan Dokter Anestesi", ya...),
			pilih("peninjauan_kegiatan_perawat_kamar_ok", "Peninjauan Perawat OK", ya...),
			teks("perhatian_utama_fase_pemulihan", "Perhatian Utama Fase Pemulihan", 100, false),
		)
	case "checklist_post_operasi":
		b = append(b, pilih("keadaan_umum", "Keadaan Umum", "Sadar", "Tidur", "Terintubasi"))
		for _, k := range []string{"rontgen", "ekg", "usg", "ctscan", "mri"} {
			b = append(b, pilih("pemeriksaan_penunjang_"+k, "Pemeriksaan "+k, ada...), teks("keterangan_pemeriksaan_penunjang_"+k, "Keterangan "+k, 20, false))
		}
		tglKateter := tanggal("tanggal_pemasangan_kateter", "Tanggal Pemasangan Kateter", "datetime-local")
		tglKateter.Wajib = false
		b = append(b,
			teks("jenis_cairan_infus", "Jenis Cairan Infus", 40, false),
			pilih("kateter_urine", "Kateter Urine", ada...), tglKateter,
			pilih("warna_kateter", "Warna Urine", "-", "Jernih", "Keruh"),
			teks("jumlah_kateter", "Jumlah Urine", 4, false),
			teks("area_luka_operasi", "Area Luka Operasi", 120, false),
			pilih("drain", "Drain", ada...), teks("jumlah_drain", "Jumlah Drain", 2, false),
			teks("letak_drain", "Letak Drain", 40, false), teks("warna_drain", "Warna Drain", 30, false),
			pilih("jaringan_pa", "Jaringan PA", ada...),
			referensi("nip_perawat_ok", "Perawat OK", "petugas"),
			referensi("nip_perawat_anestesi", "Perawat Anestesi", "petugas"),
		)
		return b
	case "laporan_operasi":
		b = []BidangOperasi{
			tanggal("tanggal", "Mulai Operasi (WITA)", "datetime-local"),
			teks("diagnosa_preop", "Diagnosis Pre Operasi", 100, false),
			teks("diagnosa_postop", "Diagnosis Post Operasi", 100, false),
			teks("jaringan_dieksekusi", "Jaringan yang Dieksisi", 100, false),
			tanggal("selesaioperasi", "Selesai Operasi (WITA)", "datetime-local"),
			pilih("permintaan_pa", "Dikirim ke PA", ya...),
			teks("laporan_operasi", "Laporan Operasi", 60000, true),
		}
		for i := range b {
			if b[i].Jenis == "text" {
				b[i].Jenis = "textarea"
			}
		}
		return b
	case "skor_bromage_pasca_anestesi", "skor_steward_pasca_anestesi", "skor_aldrette_pasca_anestesi":
		return formSkor(jenis)
	default:
		return nil
	}
	return append(b, referensi("nip_perawat_ok", "Perawat OK", "petugas"))
}

func formSkor(jenis string) []BidangOperasi {
	b := []BidangOperasi{tanggal("tanggal", "Tanggal / Jam Penilaian (WITA)", "datetime-local")}
	var skala [][]string
	switch jenis {
	case "skor_bromage_pasca_anestesi":
		skala = [][]string{{"Gerakan Tungkai", "Gerakan Penuh Dari Tungkai", "Tidak Mampu Extensi Tungkai", "Tidak Mampu Flexi Lutut", "Tidak Mampu Flexi Pergelangan Kaki"}}
	case "skor_steward_pasca_anestesi":
		skala = [][]string{
			{"Kesadaran", "Belum Respon", "Bangun Jika Dipanggil", "Sadar Penuh"},
			{"Respirasi", "Perlu Bantuan Bernafas", "Berusaha Bernafas", "Batuk / Menangis"},
			{"Aktivitas Motorik", "Tidak Bergerak", "Gerakan Tanpa Tujuan", "Gerakan Beraturan"},
		}
	default:
		skala = [][]string{
			{"Aktivitas", "Tidak Sanggup Menggerakan Satupun Anggota Gerak", "Sanggup Gerak 2 Anggota Tubuh", "Sanggup Gerak 4 Anggota Tubuh"},
			{"Respirasi", "Apnea Atau Napas Tidak Adekuat", "Sesak Atau Pernapasan Sedikit Terbatas", "Sanggup Bernafas Dalam Serta Disuruh Batuk"},
			{"Tekanan Darah", "± 50% Tekanan Darah Pra Anestesi", "± 20% - 50% Tekanan Darah Pra Anestesi", "± 20% Tekanan Darah Pra Anestesi"},
			{"Kesadaran", "Tidak Ada Respon", "Respon Terhadap Panggilan", "Sadar Penuh"},
			{"Warna Kulit", "Cianosis", "Pucat", "Kemerahan / Normal"},
		}
	}
	for i, s := range skala {
		n := strconv.Itoa(i + 1)
		f := pilih("penilaian_skala"+n, s[0], s[1:]...)
		f.NilaiKe = "penilaian_nilai" + n
		b = append(b, f, BidangOperasi{Kode: f.NilaiKe, Label: "Nilai " + s[0], Jenis: "computed"})
	}
	if len(skala) > 1 {
		b = append(b, BidangOperasi{Kode: "penilaian_totalnilai", Label: "Total Nilai", Jenis: "computed"})
	}
	return append(b, teks("keluar", "Keluar", 200, true), teks("instruksi", "Instruksi / Tindakan di Ruang Pemulihan (RR)", 200, true), referensi("kd_dokter", "Dokter Anestesi", "dokter"), referensi("nip", "Petugas", "petugas"))
}
