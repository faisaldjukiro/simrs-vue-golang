package triase_igd

import (
	"errors"
	"testing"
)

func inputValid() Input {
	return Input{NoRawat: "2026/08/14/000001", Jenis: "primer", TanggalKunjungan: "2026-08-14 10:00:00", CaraMasuk: "Jalan", AlatTransportasi: "Sendiri", AlasanKedatangan: "Datang Sendiri", KeteranganKedatangan: "Datang dengan keluarga", KodeKasus: "1", TekananDarah: "120/80", Nadi: "80", Pernapasan: "20", Suhu: "36.5", SaturasiO2: "99", Nyeri: "2", HandOver: "Penyakit Dalam", IsiUtama: "Sesak", KebutuhanKhusus: "-", Catatan: "Observasi", Plan: "Ruang Resusitasi", TanggalTriase: "2026-08-14 10:01:00", NIP: "PEG001", Skala: 1, KodeKriteria: []string{"S1"}}
}

func TestValidasiPrimer(t *testing.T) {
	input := inputValid()
	if err := validasi(input); err != nil {
		t.Fatalf("input valid ditolak: %v", err)
	}
	input.Skala = 3
	if err := validasi(input); !errors.Is(err, ErrInputTidakValid) {
		t.Fatalf("skala primer tidak valid seharusnya ditolak: %v", err)
	}
}

func TestValidasiSekunder(t *testing.T) {
	input := inputValid()
	input.Jenis, input.Skala, input.Plan = "sekunder", 4, "Zona Kuning"
	if err := validasi(input); err != nil {
		t.Fatalf("input sekunder valid ditolak: %v", err)
	}
	input.KodeKriteria = nil
	if err := validasi(input); !errors.Is(err, ErrInputTidakValid) {
		t.Fatalf("kriteria kosong seharusnya ditolak: %v", err)
	}
}
