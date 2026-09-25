package news_anak

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Bidang struct {
	Key     string   `json:"key"`
	Skor    string   `json:"skor"`
	Label   string   `json:"label"`
	Pilihan []string `json:"pilihan"`
	Nilai   []int    `json:"nilai"`
}

// Pilihan dan nilai mengikuti isCombo1..5 pada RMPemantauanNEWS.java.
func BidangPenilaian() []Bidang {
	return []Bidang{
		{"temperatur_axila", "skor_temperatur_axila", "Temperatur Axila", []string{"> 38", "37.5 - 38", "36.5 - 37.4", "35.5 - 36.4", "< 35.5", " "}, []int{2, 1, 0, 1, 2, 0}},
		{"respirasi_rate", "skor_respirasi_rate", "Respirasi Rate", []string{"> 80", "60 - 79", "30 - 59", "20 - 29", "< 20", "Grunting", "Reguler"}, []int{2, 1, 0, 1, 3, 1, 0}},
		{"denyut_nadi", "skor_denyut_nadi", "Denyut Nadi", []string{"> 190", "150 - 189", "91 - 149", "71 - 90", "<= 70"}, []int{2, 1, 0, 1, 3}},
		{"spo2", "skor_spo2", "SpO2", []string{"> 94 %", "90 - 94 %", "< 90 %"}, []int{0, 1, 3}},
		{"satus_neurology", "skor_satus_neurology", "Status Neurologi", []string{"Active / Wakes to feed", "Jittery / Iritable / Lethargy", "Floppy  / Dificult to arouse", "Seizures"}, []int{0, 1, 3, 3}},
	}
}

type Panduan struct {
	Kode     string `json:"kode"`
	Syarat   string `json:"syarat"`
	Respon   string `json:"respon"`
	Eskalasi string `json:"eskalasi"`
}

func PanduanPenilaian() []Panduan {
	return []Panduan{
		{"merah", "Ada skor 3 atau total > 5", "Segera hubungi tim respon cepat (Tim Code Blue/MET < 5 menit)", "Perawatan primer /PJ shif mengaktifkan tim medis respon cepat/tim code blue < 5 menit"},
		{"oranye", "Ada skor 2 atau 3–4 parameter berskor 1, tanpa kriteria merah", "-Pemantauan segera dan informasikan ke dr spesialis dan NICU\n-Ulangi NEWS 15 menit dengan rencana perawatan yang teridentifikasi sampai stabil\n(PPDS/DPJP/Dokter jaga konsulen)", "-perawat penanggung jawab/PPDS menginformasikan ke dr. spesialis(DPJP)\n-Jika dalam 5 menit tidak ada perubahan,hubungi NICU untuk observasi lebih lanjut"},
		{"kuning", "1–2 parameter berskor 1, tanpa kriteria di atas", "-Observasi ulang dalam 30-60 menit\n-Kaji kembali terapi pengobatan untuk perawatan selanjutnya\n-Jika Kondisi tidak stabil dalam 60 menit transfer ke ruang intensif untuk observasi (perawat/bidan primer/ketua tim/Pj shif)", "-Perawat primer lapor perawat penanggung jawab.\n-Perawat penganggung jawab menginformasikan ke dokter anak jika sudah dalam 3x5 menit kondisi pasien masih dikuning\n-Jika tidak ada perubahan dalam 5 menit menginformasikan PPDS untuk menghubungi konsultan NICU"},
		{"hijau", "Semua skor 0; kode lama juga menggunakan respon ini untuk lima skor 1 (perlu verifikasi)", "Lanjutkan pengamatan / 3 jam (perawat/bidan pelaksana)", "Berdasarkan penilaian klinis -> informasi ke perawat PJ dan dr. Spesialis anak"},
	}
}

// Urutan kondisi sengaja mengikuti isHitung(); lima skor 1 masuk fallback legacy.
func Kategori(nilai []int) string {
	total, kuning, oranye, merah := 0, 0, 0, 0
	for _, n := range nilai {
		total += n
		switch n {
		case 1:
			kuning++
		case 2:
			oranye++
		case 3:
			merah++
		}
	}
	if merah >= 1 || total > 5 {
		return "merah"
	}
	if (kuning >= 3 && kuning <= 4) || oranye >= 1 {
		return "oranye"
	}
	if kuning >= 1 && kuning <= 2 {
		return "kuning"
	}
	return "hijau"
}
func Validasi(in *Input) error {
	if strings.TrimSpace(in.NoRawat) == "" || len(in.NoRawat) > 17 || in.Data == nil {
		return ErrValidasi
	}
	if _, err := time.Parse("2006-01-02 15:04:05", in.Data["tanggal"]); err != nil {
		return fmt.Errorf("%w: tanggal dan jam tidak valid", ErrValidasi)
	}
	in.Data["nip"] = strings.TrimSpace(in.Data["nip"])
	if in.Data["nip"] == "" || utf8.RuneCountInString(in.Data["nip"]) > 20 {
		return fmt.Errorf("%w: petugas wajib diisi, maksimal 20 karakter", ErrValidasi)
	}
	total := 0
	nilaiSkor := []int{}
	for _, b := range BidangPenilaian() {
		skor := -1
		for i, pilihan := range b.Pilihan {
			if in.Data[b.Key] == pilihan {
				skor = b.Nilai[i]
				break
			}
		}
		if skor < 0 {
			return fmt.Errorf("%w: pilihan %s tidak valid", ErrValidasi, b.Label)
		}
		in.Data[b.Skor] = strconv.Itoa(skor)
		total += skor
		nilaiSkor = append(nilaiSkor, skor)
	}
	in.Data["skor_total"] = strconv.Itoa(total)
	for _, p := range PanduanPenilaian() {
		if p.Kode == Kategori(nilaiSkor) {
			in.Data["respon"] = p.Respon
			in.Data["proses_eskalasi"] = p.Eskalasi
			break
		}
	}
	return nil
}
