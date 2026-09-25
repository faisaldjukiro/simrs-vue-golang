package ews_ranap

import (
	"fmt"
	"regexp"
	"strconv"
)

var angkaEWS = regexp.MustCompile(`^[0-9]+(\.[0-9])?$`)

// Batas skor mengikuti EWSRanap.java, bukan skor yang dikirim oleh browser.
func hitungSkorCatatan(c *Catatan) error {
	parameter := []struct {
		nama  string
		nilai string
		skor  *string
		batas []float64
		poin  []int
	}{
		{"pernafasan", c.Pernafasan, &c.SkorPernafasan, []float64{8, 11, 20, 24}, []int{3, 1, 0, 2, 3}},
		{"saturasi", c.Saturasi, &c.SkorSaturasi, []float64{91, 93, 95}, []int{3, 2, 1, 0}},
		{"suhu", c.Suhu, &c.SkorSuhu, []float64{35, 36, 38, 39}, []int{3, 1, 0, 1, 2}},
		{"denyut", c.Denyut, &c.SkorDenyut, []float64{40, 50, 90, 110, 130}, []int{3, 1, 0, 1, 2, 3}},
		{"tekanan sistolik", c.Tekanan, &c.SkorTekanan, []float64{90, 100, 110, 219}, []int{3, 2, 1, 0, 3}},
	}
	total := 0
	for _, p := range parameter {
		n, err := angkaValidEWS(p.nama, p.nilai, p.nama == "suhu")
		if err != nil {
			return err
		}
		skor := skorBatas(n, p.batas, p.poin)
		*p.skor = strconv.Itoa(skor)
		total += skor
	}
	if _, err := angkaValidEWS("tekanan diastolik", c.Diastol, false); err != nil {
		return err
	}
	switch c.Alat {
	case "Ya":
		c.SkorAlat = "2"
		total += 2
	case "Tidak":
		c.SkorAlat = "0"
	default:
		return fmt.Errorf("%w: alat bantu O2 harus Ya atau Tidak", ErrInputTidakValid)
	}
	switch c.Kesadaran {
	case "A":
		c.SkorKesadaran = "0"
	case "P-V-U":
		c.SkorKesadaran = "3"
		total += 3
	default:
		return fmt.Errorf("%w: kesadaran harus A atau P-V-U", ErrInputTidakValid)
	}
	nyeri, err := strconv.Atoi(c.SkalaNyeri)
	if err != nil || nyeri < 0 || nyeri > 10 {
		return fmt.Errorf("%w: skala nyeri harus 0 sampai 10", ErrInputTidakValid)
	}
	for nama, nilai := range map[string]string{"BB": c.BB, "TB": c.TB, "lingkar kepala": c.LK, "lingkar perut": c.LP} {
		if nilai == "" {
			continue
		}
		if _, err := angkaValidEWS(nama, nilai, true); err != nil {
			return err
		}
	}
	masuk, keluar := 0, 0
	for i, nilai := range []string{c.Masuk1, c.Masuk2, c.Keluar1, c.Keluar2, c.Keluar3, c.Keluar4, c.Keluar5} {
		n, err := angkaValidEWS("volume cairan", nilai, false)
		if err != nil {
			return err
		}
		if i < 2 {
			masuk += int(n)
		} else {
			keluar += int(n)
		}
	}
	c.JumlahMasuk = strconv.Itoa(masuk)
	c.JumlahKeluar = strconv.Itoa(keluar)
	c.BC = strconv.Itoa(masuk - keluar)
	for _, nilai := range []string{c.JumlahMasuk, c.JumlahKeluar, c.BC} {
		if len(nilai) > 5 {
			return fmt.Errorf("%w: total cairan melebihi kapasitas kolom Khanza (5 karakter)", ErrInputTidakValid)
		}
	}
	c.TotalSkor = strconv.Itoa(total)
	switch {
	case total == 0:
		c.Klasifikasi = "Sangat Rendah"
	case total <= 4:
		c.Klasifikasi = "Rendah"
	case total <= 6:
		c.Klasifikasi = "Sedang"
	default:
		c.Klasifikasi = "Tinggi"
	}
	return nil
}

func angkaValidEWS(nama, nilai string, desimal bool) (float64, error) {
	n, err := strconv.ParseFloat(nilai, 64)
	if err != nil || len(nilai) > 5 || !angkaEWS.MatchString(nilai) || (!desimal && n != float64(int(n))) {
		return 0, fmt.Errorf("%w: %s harus berupa angka nonnegatif, maksimal 5 karakter%s", ErrInputTidakValid, nama, map[bool]string{true: " dan 1 angka desimal", false: " tanpa pecahan"}[desimal])
	}
	return n, nil
}

func skorBatas(nilai float64, batas []float64, poin []int) int {
	for i, maksimum := range batas {
		if nilai <= maksimum {
			return poin[i]
		}
	}
	return poin[len(poin)-1]
}
