package jadwal_operasi

func narasi(k, label string, batas int, wajib bool) BidangOperasi {
	b := teks(k, label, batas, wajib)
	b.Jenis = "textarea"
	return b
}

func formPenilaian(jenis string) []BidangOperasi {
	b := []BidangOperasi{tanggal("tanggal", "Tanggal / Jam Penilaian (WITA)", "datetime-local"), referensi("kd_dokter", "Dokter Pemeriksa", "dokter")}
	switch jenis {
	case "penilaian_pre_operasi":
		return append(b,
			narasi("ringkasan_klinik", "Ringkasan Klinik", 500, true),
			narasi("pemeriksaan_fisik", "Pemeriksaan Fisik", 500, false),
			narasi("pemeriksaan_diagnostik", "Pemeriksaan Diagnostik", 500, false),
			narasi("diagnosa_pre_operasi", "Diagnosis Pre Operasi", 500, true),
			narasi("rencana_tindakan_bedah", "Rencana Tindakan Bedah", 500, false),
			narasi("hal_hal_yang_perludi_persiapkan", "Hal yang Perlu Dipersiapkan", 500, false),
			narasi("terapi_pre_operasi", "Terapi Pre Operasi", 500, false),
		)
	case "penilaian_pre_anestesi":
		b = append(b, tanggal("tanggal_operasi", "Tanggal / Jam Operasi", "datetime-local"), teks("diagnosa", "Diagnosis", 100, true), teks("rencana_tindakan", "Rencana Tindakan", 100, true))
		for _, f := range [][2]string{{"tb", "Tinggi Badan"}, {"bb", "Berat Badan"}, {"td", "Tekanan Darah"}, {"io2", "IO2"}, {"nadi", "Nadi"}, {"pernapasan", "Pernapasan"}, {"suhu", "Suhu"}} {
			batas := 5
			if f[0] == "td" {
				batas = 8
			}
			b = append(b, teks(f[0], f[1], batas, false))
		}
		for _, f := range [][2]string{{"cardiovasculer", "Kardiovaskular"}, {"paru", "Paru"}, {"abdomen", "Abdomen"}, {"extrimitas", "Ekstremitas"}, {"endokrin", "Endokrin"}, {"ginjal", "Ginjal"}, {"obatobatan", "Obat-obatan"}, {"laborat", "Laboratorium"}, {"penunjang", "Penunjang"}} {
			b = append(b, teks("fisik_"+f[0], "Pemeriksaan: "+f[1], 100, false))
		}
		b = append(b,
			teks("riwayat_penyakit_alergiobat", "Riwayat Alergi Obat", 50, false),
			teks("riwayat_penyakit_alergilainnya", "Riwayat Alergi Lainnya", 50, false),
			teks("riwayat_penyakit_terapi", "Riwayat Terapi", 100, false),
			pilih("riwayat_kebiasaan_merokok", "Kebiasaan Merokok", "Tidak", "Ya"),
			teks("riwayat_kebiasaan_ket_merokok", "Jumlah Rokok", 5, false),
			pilih("riwayat_kebiasaan_alkohol", "Kebiasaan Alkohol", "Tidak", "Ya"),
			teks("riwayat_kebiasaan_ket_alkohol", "Jumlah Alkohol", 5, false),
			pilih("riwayat_kebiasaan_obat", "Kebiasaan Obat", "-", "Obat Obatan", "Vitamin", "Jamu Jamuan"),
			teks("riwayat_kebiasaan_ket_obat", "Obat yang Diminum", 100, false),
		)
		for _, f := range [][2]string{{"cardiovasculer", "Kardiovaskular"}, {"respiratory", "Respirasi"}, {"endocrine", "Endokrin"}, {"lainnya", "Lainnya"}} {
			b = append(b, teks("riwayat_medis_"+f[0], "Riwayat Medis: "+f[1], 100, false))
		}
		return append(b, pilih("asa", "ASA", "1", "2", "3", "4", "5", "E"), tanggal("puasa", "Mulai Puasa", "datetime-local"), pilih("rencana_anestesi", "Rencana Anestesi", "GA", "RA Spinal", "RA Epidural", "RA Combined", "Blok Syaraf"), teks("rencana_perawatan", "Rencana Perawatan", 40, false), teks("catatan_khusus", "Catatan Khusus", 100, false))
	case "penilaian_pre_induksi":
		b = append(b, teks("tensi", "Tekanan Darah", 8, false), teks("nadi", "Nadi", 5, false), teks("rr", "RR", 5, false), teks("suhu", "Suhu", 5, false), teks("ekg", "EKG", 50, false), teks("lain_lain", "Lain-lain", 50, false),
			pilih("asesmen", "Asesmen", "Sesuai Asesmen Pre Sedasi/Anestesi", "Tidak Sesuai Asesmen Pre Sedasi/Anestesi"),
			narasi("perencanaan", "Perencanaan", 300, true), narasi("infus_perifier", "Infus Perifer, Tempat dan Ukuran", 300, true), teks("cvc", "CVC", 70, false),
			pilih("posisi", "Posisi", "Supine", "Lithotomi", "Lateral", "Prone", "Perlindungan Mata", "Kanan", "Kiri", "Lain-lain"),
			pilih("premedikasi", "Premedikasi", "Oral", "IM", "IV"), teks("premedikasi_keterangan", "Keterangan Premedikasi", 50, false),
			pilih("induksi", "Induksi", "Intravena", "Inhalasi"), teks("induksi_keterangan", "Keterangan Induksi", 70, false),
		)
		for _, f := range [][2]string{{"face_mask_no", "Face Mask No."}, {"nasopharing_no", "Nasofaring No."}, {"ett_no", "ETT No."}, {"ett_jenis", "Jenis ETT"}, {"ett_viksasi", "Fiksasi ETT"}, {"lma_no", "LMA No."}, {"lma_jenis", "Jenis LMA"}} {
			batas := 20
			if f[0] == "ett_viksasi" {
				batas = 25
			}
			b = append(b, teks(f[0], f[1], batas, false))
		}
		b = append(b, teks("tracheostomi", "Trakeostomi", 60, false), teks("bronchoscopi_fiberoptik", "Bronkoskopi Fiberoptik", 60, false), teks("glidescopi", "Glideskopi", 60, false), teks("lain_lain_tatalaksana", "Tatalaksana Lainnya", 100, false),
			pilih("intubasi_sesudah_tidur", "Intubasi Sesudah Tidur", "Tidak", "Ya"), pilih("intubasi_oral", "Intubasi Oral", "Tidak", "Ya"), pilih("intubasi_tracheostomi", "Intubasi Trakeostomi", "Tidak", "Ya"),
			teks("intubasi_keterangan", "Keterangan Intubasi", 200, false), teks("sulit_ventilasi", "Sulit Ventilasi", 100, false), teks("sulit_intubasi", "Sulit Intubasi", 100, false), teks("ventilasi", "Ventilasi", 100, false),
			teks("teknik_regional_jenis", "Teknik Regional: Jenis", 100, false), teks("teknik_regional_lokasi", "Teknik Regional: Lokasi", 40, false), teks("teknik_regional_jenis_jarum", "Teknik Regional: Jenis Jarum", 30, false),
			pilih("teknik_regional_kateter", "Teknik Regional: Kateter", "Tidak", "Ya"), teks("teknik_regional_kateter_viksasi", "Teknik Regional: Fiksasi Kateter", 40, false),
			narasi("teknik_regional_obat_obatan", "Teknik Regional: Obat-obatan", 400, true), narasi("teknik_regional_komplikasi", "Teknik Regional: Komplikasi", 200, false), teks("teknik_regional_hasil", "Teknik Regional: Hasil", 100, true),
		)
		return b
	}
	return nil
}

func formTransfer() []BidangOperasi {
	b := []BidangOperasi{
		tanggal("tanggal_masuk", "Tanggal / Jam Masuk", "datetime-local"), tanggal("tanggal_pindah", "Tanggal / Jam Pindah", "datetime-local"),
		teks("asal_ruang", "Asal Ruang", 30, true), teks("ruang_selanjutnya", "Ruang Selanjutnya", 30, true),
		teks("diagnosa_utama", "Diagnosis Utama", 50, true), teks("diagnosa_sekunder", "Diagnosis Sekunder", 100, false),
		pilih("indikasi_pindah_ruang", "Indikasi Pindah Ruang", "Kondisi Pasien Stabil", "Kondisi Pasien Tidak Ada Perubahan", "Kondisi Pasien Memburuk", "Fasilitas Kurang Memadai", "Fasilitas Butuh Lebih Baik", "Tenaga Membutuhkan Yang Lebih Ahli", "Tenaga Kurang", "Lain-lain"),
		teks("keterangan_indikasi_pindah_ruang", "Keterangan Indikasi Pindah", 50, false),
		narasi("prosedur_yang_sudah_dilakukan", "Prosedur yang Sudah Dilakukan", 800, true), narasi("obat_yang_telah_diberikan", "Obat yang Telah Diberikan", 800, true),
		pilih("metode_pemindahan_pasien", "Metode Pemindahan Pasien", "Kursi Roda", "Tempat Tidur", "Brankar"),
		pilih("peralatan_yang_menyertai", "Peralatan yang Menyertai", "Oksigen Portable", "Infus", "NGT", "Syringe Pump", "Suction", "Kateter Urin"),
		teks("keterangan_peralatan_yang_menyertai", "Keterangan Peralatan", 50, false), narasi("pemeriksaan_penunjang_yang_dilakukan", "Pemeriksaan Penunjang", 500, false),
		pilih("pasien_keluarga_menyetujui", "Pasien / Keluarga Menyetujui", "Ya", "Tidak"), teks("nama_menyetujui", "Nama yang Menyetujui", 50, false),
		pilih("hubungan_menyetujui", "Hubungan dengan Pasien", "Kakak", "Adik", "Saudara", "Keluarga", "Kakek", "Nenek", "Orang Tua", "Suami", "Istri", "Penanggung Jawab", "Menantu", "Ipar", "Mertua", "-"),
	}
	for _, w := range []string{"sebelum", "sesudah"} {
		b = append(b, teks("keluhan_utama_"+w+"_transfer", "Keluhan Utama "+w+" Transfer", 200, true), pilih("keadaan_umum_"+w+"_transfer", "Keadaan Umum "+w+" Transfer", "Compos Mentis", "Gelisah", "Delirium", "Koma"))
		for _, f := range [][2]string{{"td", "TD"}, {"nadi", "Nadi"}, {"rr", "RR"}, {"suhu", "Suhu"}} {
			batas := 5
			if f[0] == "td" {
				batas = 7
			}
			b = append(b, teks(f[0]+"_"+w+"_transfer", f[1]+" "+w+" Transfer", batas, true))
		}
	}
	return append(b, referensi("nip_menyerahkan", "Petugas Menyerahkan", "petugas"), referensi("nip_menerima", "Petugas Menerima", "petugas"))
}
