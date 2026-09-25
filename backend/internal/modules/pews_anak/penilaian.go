package pews_anak

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
}

// Urutan pilihan adalah skor 0, 1, 2, 3 dari RMPemantauanPEWS.java.
func BidangPenilaian() []Bidang {
	return []Bidang{
		{"parameter_perilaku", "skor_perilaku", "Perilaku", []string{
			"Sadar / Bermain", "Tidur / Perubahan Perilaku", "Gelisah",
			"Tidak Merespon Terhadap Nyeri Penurunan Kesadaran",
		}},
		{"parameter_crt_atau_warna_kulit", "skor_crt_atau_warna_kulit", "CRT / Warna Kulit", []string{
			"1 - 2 dtk / Pink", "3 dtk / Pucat", "4 dtk / Sianosis", ">=5 dtk / Mottle",
		}},
		{"parameter_perespirasi", "skor_perespirasi", "Pernapasan", []string{
			"Tidak Ada Retraksi", "Cuping Hidung / O2 1-3 Lpm", "Retraksi Dada / O2 4-6 Lpm", "Stridor / O2 7-8 Lpm",
		}},
	}
}

type Panduan struct {
	Minimal   int    `json:"minimal"`
	Maksimal  int    `json:"maksimal"`
	Parameter string `json:"parameter"`
}

// Ikuti isHitung() yang dijalankan sebelum simpan, bukan teks reset legacy.
func PanduanPenilaian() []Panduan {
	return []Panduan{
		{0, 2, "Monitoring setiap 4 jam oleh perawat pelaksana dan di lanjutkan observasi atau monitoring secara rutin"},
		{3, 3, "Monitoring 1 sampai 2 jam. Pengkajian ulang dilakukan oleh PJ sift dan laporkan ke dokter jaga"},
		{4, 4, "Monitor per 1 jam. Laporkan ke dokter jaga dan kemudian tindak lanjut lapor ke DPJP untuk advis selanjutnya. Kolaborasi langkah selanjutnya dengan seluruh tim perawatan. Jika masih di perlukan lapor ulang keperawat ketua tim dan DPJP"},
		{5, 9, "Laporkan perubahan klinis ke perawat ketua tim, dokter jaga dan DPJP. Kolaborasikan langkah selanjutnya dengan seluruh tim perawatan"},
	}
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
	for _, b := range BidangPenilaian() {
		skor := -1
		for i, pilihan := range b.Pilihan {
			if in.Data[b.Key] == pilihan {
				skor = i
				break
			}
		}
		if skor < 0 {
			return fmt.Errorf("%w: pilihan %s tidak valid", ErrValidasi, b.Label)
		}
		in.Data[b.Skor] = strconv.Itoa(skor)
		total += skor
	}
	in.Data["skor_total"] = strconv.Itoa(total)
	for _, p := range PanduanPenilaian() {
		if total >= p.Minimal && total <= p.Maksimal {
			in.Data["parameter_total"] = p.Parameter
			break
		}
	}
	return nil
}
