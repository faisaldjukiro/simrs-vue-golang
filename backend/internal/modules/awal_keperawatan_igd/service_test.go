package awal_keperawatan_igd

import (
	"errors"
	"testing"
)

func inputValid() Input {
	return Input{
		NoRawat: "2026/08/14/000001", Tanggal: "2026-08-14 10:00:00",
		Informasi: "Autoanamnesis", KeluhanUtama: "Demam", RPD: "Tidak ada",
		RPO: "Tidak ada", NIP: "PEG001",
	}
}

func TestValidasiWajibMengikutiKhanza(t *testing.T) {
	input := inputValid()
	input.RPD = ""
	if err := validasi(input); !errors.Is(err, ErrInputTidakValid) {
		t.Fatalf("diharapkan validasi gagal, mendapat %v", err)
	}
}

func TestUrutanArgumenSesuai69Kolom(t *testing.T) {
	args := inputArgs(inputValid())
	if len(args) != 69 {
		t.Fatalf("jumlah argumen harus 69, mendapat %d", len(args))
	}
	if args[0] != "2026/08/14/000001" || args[68] != "PEG001" {
		t.Fatalf("urutan argumen utama tidak sesuai: pertama=%v terakhir=%v", args[0], args[68])
	}
}

func TestBersihkanKodeRelasiDuplikat(t *testing.T) {
	input := inputValid()
	input.KodeMasalah = []string{"001", " 001 ", "002", ""}
	bersihkan(&input)
	if len(input.KodeMasalah) != 2 || input.KodeMasalah[0] != "001" || input.KodeMasalah[1] != "002" {
		t.Fatalf("kode relasi belum dinormalisasi: %#v", input.KodeMasalah)
	}
}
