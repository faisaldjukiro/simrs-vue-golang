package triase_igd

import (
	"context"
	"os"
	"testing"
	"time"

	"simrs-backend/internal/config"
	"simrs-backend/internal/platform/database"
)

func TestIntegrasiBacaReferensiTriaseIGD(t *testing.T) {
	if os.Getenv("SIRAVA_SIMRS_INTEGRATION") != "1" {
		t.Skip("aktifkan SIRAVA_SIMRS_INTEGRATION=1 untuk menguji koneksi SIMRS")
	}
	if err := config.LoadEnvFile("../../../.env"); err != nil {
		t.Fatal(err)
	}
	konfigurasi, err := config.SIMRSDatabase()
	if err != nil {
		t.Fatal(err)
	}
	ctx, batal := context.WithTimeout(context.Background(), 30*time.Second)
	defer batal()
	db, err := database.OpenMySQL(ctx, konfigurasi)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRepositori(db)
	pemeriksaan, err := repo.pemeriksaan(ctx)
	if err != nil {
		t.Fatal(err)
	}
	kriteria, err := repo.kriteria(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(pemeriksaan) == 0 {
		t.Fatal("master_triase_pemeriksaan kosong")
	}
	if len(kriteria) == 0 {
		t.Fatal("master kriteria skala triase kosong")
	}
	var noRawat string
	if err := db.QueryRowContext(ctx, `SELECT no_rawat FROM data_triase_igd ORDER BY tgl_kunjungan DESC LIMIT 1`).Scan(&noRawat); err != nil {
		t.Fatal(err)
	}
	hasil, err := repo.Data(ctx, noRawat, "")
	if err != nil {
		t.Fatalf("data triase yang sudah ada gagal dibaca: %v", err)
	}
	if hasil.Triase == nil {
		t.Fatalf("data triase %s tidak ditemukan", noRawat)
	}
	t.Logf("pemeriksaan=%d kriteria=%d no_rawat=%s primer=%t sekunder=%t", len(pemeriksaan), len(kriteria), noRawat, hasil.Triase.Primer != nil, hasil.Triase.Sekunder != nil)
}
