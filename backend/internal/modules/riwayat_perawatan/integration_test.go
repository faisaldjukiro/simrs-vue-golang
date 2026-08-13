package riwayat_perawatan

import (
	"context"
	"os"
	"testing"
	"time"

	"simrs-backend/internal/config"
	"simrs-backend/internal/platform/database"
)

func TestIntegrasiBacaRiwayatPerawatanSIMRS(t *testing.T) {
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

	var noRawat string
	noRawat = os.Getenv("SIRAVA_SIMRS_TEST_NO_RAWAT")
	if noRawat == "" {
		if err := db.QueryRowContext(ctx, `
		SELECT no_rawat FROM reg_periksa
		WHERE stts<>'Batal' AND no_rkm_medis<>''
		ORDER BY tgl_registrasi DESC,jam_reg DESC LIMIT 1
		`).Scan(&noRawat); err != nil {
			t.Fatal(err)
		}
	}
	hasil, err := NewRepositori(db, config.SIMRSWebBaseURL()).Data(ctx, Filter{NoRawat: noRawat, Mode: "nomor", NomorRawat: noRawat})
	if err != nil {
		t.Fatal(err)
	}
	if hasil.NoRekamMedis == "" || len(hasil.Kunjungan) == 0 {
		t.Fatal("riwayat pasien dari SIMRS kosong")
	}
	if len(hasil.Kunjungan[0].TandaTangan) == 0 {
		t.Fatalf("tanda tangan/verifikasi untuk nomor rawat %s belum terbaca", noRawat)
	}
	if hasil.Kunjungan[0].TandaTangan[0].URLQRCode == "" {
		t.Fatal("alamat QR tanda tangan/verifikasi belum terbentuk")
	}
	hasilTerakhir, err := NewRepositori(db, config.SIMRSWebBaseURL()).Data(ctx, Filter{NoRawat: noRawat, Mode: "terakhir"})
	if err != nil {
		t.Fatal(err)
	}
	if len(hasilTerakhir.Kunjungan) == 0 || len(hasilTerakhir.Kunjungan) > 5 {
		t.Fatalf("5 riwayat terakhir tidak valid: ditemukan %d kunjungan", len(hasilTerakhir.Kunjungan))
	}

	var noRawatTriase string
	if err := db.QueryRowContext(ctx, `SELECT no_rawat FROM data_triase_igd ORDER BY no_rawat DESC LIMIT 1`).Scan(&noRawatTriase); err != nil {
		t.Fatalf("contoh data triase tidak dapat dibaca: %v", err)
	}
	hasilTriase, err := NewRepositori(db, config.SIMRSWebBaseURL()).Data(ctx, Filter{NoRawat: noRawatTriase, Mode: "nomor", NomorRawat: noRawatTriase})
	if err != nil {
		t.Fatal(err)
	}
	triaseDitemukan := false
	for _, dokumen := range hasilTriase.Kunjungan[0].DokumenKlinis {
		if dokumen.Tabel == "data_triase_igd" && len(dokumen.Data) > 0 {
			triaseDitemukan = true
			break
		}
	}
	if !triaseDitemukan {
		t.Fatalf("triase untuk nomor rawat %s belum ikut dalam riwayat perawatan", noRawatTriase)
	}
	t.Logf("no_rawat=%s kunjungan=%d berkas_digital=%d tanda_tangan=%d", noRawat, len(hasil.Kunjungan), len(hasil.Kunjungan[0].BerkasDigital), len(hasil.Kunjungan[0].TandaTangan))
	t.Logf("no_rawat_triase=%s triase_terbaca=%t", noRawatTriase, triaseDitemukan)
}
