package cppt

import "testing"

func TestNormalisasiJenisRawat(t *testing.T) {
	for input, ingin := range map[string]string{
		"ralan":       "ralan",
		"IGD":         "ralan",
		"Rawat Jalan": "ralan",
		"ranap":       "ranap",
		"":            "ranap",
	} {
		if hasil := normalisasiJenisRawat(input); hasil != ingin {
			t.Fatalf("normalisasi %q menghasilkan %q, ingin %q", input, hasil, ingin)
		}
	}
}

func TestValidasiLingkarPerutRalan(t *testing.T) {
	catatan := Catatan{
		JenisRawat:       "ralan",
		NoRawat:          "2026/08/14/000001",
		TanggalPerawatan: "2026-08-14",
		JamRawat:         "10:00:00",
		Kesadaran:        "Compos Mentis",
		LingkarPerut:     "123456",
	}
	if err := validasiCatatan(&catatan); err == nil {
		t.Fatal("lingkar perut lebih dari 5 karakter seharusnya ditolak")
	}
}
