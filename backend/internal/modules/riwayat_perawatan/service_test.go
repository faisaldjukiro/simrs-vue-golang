package riwayat_perawatan

import (
	"context"
	"errors"
	"testing"
)

func TestValidasiFilterRiwayatPerawatan(t *testing.T) {
	layanan := NewLayanan(nil)
	tests := []Filter{
		{},
		{NoRawat: "2026/08/13/000001", Mode: "asing"},
		{NoRawat: "2026/08/13/000001", Mode: "tanggal", TanggalMulai: "2026-08-14", TanggalSelesai: "2026-08-13"},
		{NoRawat: "2026/08/13/000001", Mode: "nomor"},
	}
	for _, filter := range tests {
		_, err := layanan.Data(context.Background(), filter)
		if !errors.Is(err, ErrInputTidakValid) {
			t.Fatalf("filter %+v seharusnya tidak valid, dapat %v", filter, err)
		}
	}
}
