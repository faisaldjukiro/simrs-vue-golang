package permintaan_radiologi

import "testing"

func TestKodeCaraBayarTarifRadiologi(t *testing.T) {
	tests := []struct {
		kode  string
		ingin string
	}{
		{kode: "BPJ", ingin: "BPJ"},
		{kode: "bpj", ingin: "BPJ"},
		{kode: "36", ingin: "BPJ"},
		{kode: "A09", ingin: "A09"},
		{kode: "UMU", ingin: "A09"},
		{kode: "", ingin: "A09"},
	}

	for _, tt := range tests {
		if hasil := kodeCaraBayarTarifRadiologi(tt.kode); hasil != tt.ingin {
			t.Fatalf("kode %q menghasilkan %q, ingin %q", tt.kode, hasil, tt.ingin)
		}
	}
}

func TestPerbaruiStatusPermintaan(t *testing.T) {
	tests := []struct {
		nama              string
		permintaan        Permintaan
		statusPemeriksaan string
		statusBayar       string
		dapatDiubah       bool
	}{
		{
			nama: "menunggu dan belum bayar masih dapat diubah",
			permintaan: Permintaan{Pemeriksaan: []Pemeriksaan{
				{StatusBayar: "Belum"},
			}},
			statusPemeriksaan: "Menunggu Radiologi",
			statusBayar:       "Belum Bayar",
			dapatDiubah:       true,
		},
		{
			nama: "sudah diterima menjadi sedang dikerjakan dan terkunci",
			permintaan: Permintaan{TanggalDiterima: "2026-08-13", Pemeriksaan: []Pemeriksaan{
				{StatusBayar: "Belum"},
			}},
			statusPemeriksaan: "Sedang Dikerjakan",
			statusBayar:       "Belum Bayar",
			dapatDiubah:       false,
		},
		{
			nama: "hasil selesai dan pembayaran sebagian",
			permintaan: Permintaan{TanggalHasil: "2026-08-13", Pemeriksaan: []Pemeriksaan{
				{StatusBayar: "Sudah"}, {StatusBayar: "Belum"},
			}},
			statusPemeriksaan: "Selesai",
			statusBayar:       "Sebagian Dibayar",
			dapatDiubah:       false,
		},
		{
			nama: "semua tindakan sudah dibayar",
			permintaan: Permintaan{Pemeriksaan: []Pemeriksaan{
				{StatusBayar: "Sudah"}, {StatusBayar: "Sudah"},
			}},
			statusPemeriksaan: "Menunggu Radiologi",
			statusBayar:       "Sudah Bayar",
			dapatDiubah:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			perbaruiStatusPermintaan(&tt.permintaan)
			if tt.permintaan.StatusPemeriksaan != tt.statusPemeriksaan {
				t.Fatalf("status pemeriksaan = %q, ingin %q", tt.permintaan.StatusPemeriksaan, tt.statusPemeriksaan)
			}
			if tt.permintaan.StatusBayar != tt.statusBayar {
				t.Fatalf("status bayar = %q, ingin %q", tt.permintaan.StatusBayar, tt.statusBayar)
			}
			if tt.permintaan.DapatDiubah != tt.dapatDiubah || tt.permintaan.DapatDihapus != tt.dapatDiubah {
				t.Fatalf("izin ubah/hapus = %v/%v, ingin %v", tt.permintaan.DapatDiubah, tt.permintaan.DapatDihapus, tt.dapatDiubah)
			}
		})
	}
}
