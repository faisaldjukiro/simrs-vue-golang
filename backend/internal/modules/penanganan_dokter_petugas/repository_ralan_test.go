package penanganan_dokter_petugas

import (
	"strings"
	"testing"
)

func TestKondisiTarifRalan(t *testing.T) {
	tests := []struct {
		nama        string
		aturan      aturanTarifRalan
		adaPoli     bool
		adaPenjamin bool
	}{
		{nama: "keduanya aktif", aturan: aturanTarifRalan{Poliklinik: true, CaraBayar: true}, adaPoli: true, adaPenjamin: true},
		{nama: "hanya poliklinik", aturan: aturanTarifRalan{Poliklinik: true}, adaPoli: true},
		{nama: "hanya cara bayar", aturan: aturanTarifRalan{CaraBayar: true}, adaPenjamin: true},
		{nama: "tanpa filter", aturan: aturanTarifRalan{}},
	}

	for _, test := range tests {
		t.Run(test.nama, func(t *testing.T) {
			kondisi := kondisiTarifRalan(test.aturan)
			if strings.Contains(kondisi, "kd_poli") != test.adaPoli {
				t.Fatalf("filter poliklinik tidak sesuai: %q", kondisi)
			}
			if strings.Contains(kondisi, "kd_pj") != test.adaPenjamin {
				t.Fatalf("filter cara bayar tidak sesuai: %q", kondisi)
			}
		})
	}
}

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
