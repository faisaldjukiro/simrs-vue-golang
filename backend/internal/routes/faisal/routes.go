// Package faisal registers routes maintained by Faisal.
package faisal

import (
	"net/http"

	"github.com/gin-gonic/gin"

	aktivitasloghttp "simrs-backend/internal/modules/aktivitas_log/delivery/http"
	autentikasihttp "simrs-backend/internal/modules/autentikasi/delivery/http"
	awalkeperawatanigdhttp "simrs-backend/internal/modules/awal_keperawatan_igd/delivery/http"
	awalMedisIgdHttp "simrs-backend/internal/modules/awal_medis_igd/delivery/http"
	awalmedisranaphttp "simrs-backend/internal/modules/awal_medis_ranap/delivery/http"
	awalmedisumumhttp "simrs-backend/internal/modules/awal_medis_umum/delivery/http"
	berandahttp "simrs-backend/internal/modules/beranda/delivery/http"
	berkasdigitalhttp "simrs-backend/internal/modules/berkas_digital/delivery/http"
	bpjshttp "simrs-backend/internal/modules/bpjs/delivery/http"
	dataklaimhttp "simrs-backend/internal/modules/bpjs/vclaim/monitoring/data_klaim/delivery/http"
	cppthttp "simrs-backend/internal/modules/cppt/delivery/http"
	diagnosapasienhttp "simrs-backend/internal/modules/diagnosa_pasien/delivery/http"
	ewsranaphttp "simrs-backend/internal/modules/ews_ranap/delivery/http"
	idrghttp "simrs-backend/internal/modules/idrg/delivery/http"
	implementasikeperawatanhttp "simrs-backend/internal/modules/implementasi_keperawatan/delivery/http"
	kelolamenuhttp "simrs-backend/internal/modules/kelola_menu/delivery/http"
	laporan10penyakithttp "simrs-backend/internal/modules/laporan_10_penyakit/delivery/http"
	laporanborlostoihttp "simrs-backend/internal/modules/laporan_bor_los_toi/delivery/http"
	laporankunjunganralanhttp "simrs-backend/internal/modules/laporan_kunjungan_ralan/delivery/http"
	laporankunjunganranaphttp "simrs-backend/internal/modules/laporan_kunjungan_ranap/delivery/http"
	laporanpenggunaanbedhttp "simrs-backend/internal/modules/laporan_penggunaan_bed/delivery/http"
	manajemenpenggunahttp "simrs-backend/internal/modules/manajemen_pengguna/delivery/http"
	"simrs-backend/internal/modules/monitoring_bed"
	penanganandokterpetugashttp "simrs-backend/internal/modules/penanganan_dokter_petugas/delivery/http"
	permintaanlaboratoriumhttp "simrs-backend/internal/modules/permintaan_laboratorium/delivery/http"
	permintaanradiologihttp "simrs-backend/internal/modules/permintaan_radiologi/delivery/http"
	resephttp "simrs-backend/internal/modules/resep/delivery/http"
	resumepasienhttp "simrs-backend/internal/modules/resume_pasien/delivery/http"
	resumepasienranaphttp "simrs-backend/internal/modules/resume_pasien_ranap/delivery/http"
	riwayatperawatanhttp "simrs-backend/internal/modules/riwayat_perawatan/delivery/http"
	triaseigdhttp "simrs-backend/internal/modules/triase_igd/delivery/http"
	ventilatorhttp "simrs-backend/internal/modules/ventilator/delivery/http"
	whatsappgatewayhttp "simrs-backend/internal/modules/whatsapp_gateway/delivery/http"
	"simrs-backend/internal/shared/httpresponse"
)

type Dependencies struct {
	AktivitasLog            *aktivitasloghttp.Handler
	Autentikasi             *autentikasihttp.Handler
	AwalKeperawatanIGD      *awalkeperawatanigdhttp.Handler
	AwalMedisUmum           *awalmedisumumhttp.Handler
	AwalMedisRanap          *awalmedisranaphttp.Handler
	AwalMedisIgd            *awalMedisIgdHttp.Handler
	Beranda                 *berandahttp.Handler
	BPJS                    *bpjshttp.Handler
	BPJSDataKlaim           *dataklaimhttp.Handler
	BerkasDigital           *berkasdigitalhttp.Handler
	IDRG                    *idrghttp.Handler
	KelolaMenu              *kelolamenuhttp.Handler
	Laporan10Penyakit       *laporan10penyakithttp.Handler
	LaporanBORLOSTOI        *laporanborlostoihttp.Handler
	LaporanPenggunaanBed    *laporanpenggunaanbedhttp.Handler
	MonitoringBed           *monitoring_bed.Handler
	LaporanKunjunganRalan   *laporankunjunganralanhttp.Handler
	LaporanKunjunganRanap   *laporankunjunganranaphttp.Handler
	ManajemenPengguna       *manajemenpenggunahttp.Handler
	CPPT                    *cppthttp.Handler
	DiagnosaPasien          *diagnosapasienhttp.Handler
	EWSRanap                *ewsranaphttp.Handler
	ImplementasiKeperawatan *implementasikeperawatanhttp.Handler
	PenangananDokterPetugas *penanganandokterpetugashttp.Handler
	PermintaanRadiologi     *permintaanradiologihttp.Handler
	PermintaanLaboratorium  *permintaanlaboratoriumhttp.Handler
	ResumePasien            *resumepasienhttp.Handler
	ResumePasienRanap       *resumepasienranaphttp.Handler
	Resep                   *resephttp.Handler
	RiwayatPerawatan        *riwayatperawatanhttp.Handler
	TriaseIGD               *triaseigdhttp.Handler
	WhatsAppGateway         *whatsappgatewayhttp.Handler
	Ventilator              *ventilatorhttp.Handler
}

// Register attaches Faisal's routes to the shared API router. Developer names
// are intentionally not included in public URLs.
func Register(router *gin.Engine, dependencies Dependencies) {
	router.GET("/health", func(c *gin.Context) {
		httpresponse.Success(c, http.StatusOK, gin.H{"status": "ok"})
	})

	dependencies.Autentikasi.Register(router.Group("/api/auth"))

	protectedAPI := router.Group("/api")
	protectedAPI.Use(dependencies.Autentikasi.Middleware())
	dependencies.AktivitasLog.Register(protectedAPI.Group("/aktivitas-log"))
	protectedAPI.GET("/beranda", dependencies.Beranda.Beranda)
	protectedAPI.GET("/beranda/koneksi-eklaim", dependencies.Beranda.KoneksiEKlaim)
	protectedAPI.GET("/beranda/koneksi-bpjs", dependencies.Beranda.KoneksiBPJS)
	dependencies.BPJS.Register(protectedAPI.Group("/bpjs"))
	dependencies.BPJSDataKlaim.Register(protectedAPI.Group("/bpjs/monitoring/klaim"))
	dependencies.IDRG.Register(protectedAPI.Group("/idrg"))
	laporan10PenyakitGroup := protectedAPI.Group("/laporan-10-penyakit")
	laporan10PenyakitGroup.Use(dependencies.Autentikasi.WajibPermission("laporan_10_penyakit"))
	dependencies.Laporan10Penyakit.Register(laporan10PenyakitGroup)
	laporanBORLOSTOIGroup := protectedAPI.Group("/laporan-bor-los-toi")
	laporanBORLOSTOIGroup.Use(dependencies.Autentikasi.WajibPermission("laporan_bor_los_toi"))
	dependencies.LaporanBORLOSTOI.Register(laporanBORLOSTOIGroup)
	laporanPenggunaanBedGroup := protectedAPI.Group("/laporan-penggunaan-bed")
	laporanPenggunaanBedGroup.Use(dependencies.Autentikasi.WajibPermission("laporan_penggunaan_bed"))
	dependencies.LaporanPenggunaanBed.Register(laporanPenggunaanBedGroup)
	monitoringBedGroup := protectedAPI.Group("/monitoring-bed")
	monitoringBedGroup.Use(dependencies.Autentikasi.WajibPermission("monitoring_bed"))
	dependencies.MonitoringBed.Route(monitoringBedGroup)
	laporanKunjunganGroup := protectedAPI.Group("/laporan-kunjungan-ralan")
	laporanKunjunganGroup.Use(dependencies.Autentikasi.WajibPermission("laporan_kunjungan_ralan"))
	dependencies.LaporanKunjunganRalan.Register(laporanKunjunganGroup)
	laporanRanapGroup := protectedAPI.Group("/laporan-kunjungan-ranap")
	laporanRanapGroup.Use(dependencies.Autentikasi.WajibPermission("laporan_kunjungan_ranap"))
	dependencies.LaporanKunjunganRanap.Register(laporanRanapGroup)
	whatsAppGatewayGroup := protectedAPI.Group("/whatsapp-gateway")
	whatsAppGatewayGroup.Use(dependencies.Autentikasi.WajibPermission("whatsapp_gateway"))
	dependencies.WhatsAppGateway.Register(whatsAppGatewayGroup)
	kelolaMenuGroup := protectedAPI.Group("/kelola-menu")
	kelolaMenuGroup.Use(dependencies.Autentikasi.WajibPermission("kelola_menu"))
	dependencies.KelolaMenu.Register(kelolaMenuGroup)
	masterVentilatorGroup := protectedAPI.Group("/master-ventilator")
	masterVentilatorGroup.Use(dependencies.Autentikasi.WajibPermission("master_ventilator"))
	dependencies.Ventilator.RegisterMaster(masterVentilatorGroup)
	daftarRouteSidebar := []struct {
		path     string
		register func(*gin.RouterGroup)
	}{
		{"/cppt", dependencies.CPPT.Register},
		{"/diagnosa-pasien", dependencies.DiagnosaPasien.Register},
		{"/ews-ranap", dependencies.EWSRanap.Register},
		{"/implementasi-keperawatan", dependencies.ImplementasiKeperawatan.Register},
		{"/penanganan-dokter-petugas", dependencies.PenangananDokterPetugas.Register},
		{"/permintaan-radiologi", dependencies.PermintaanRadiologi.Register},
		{"/permintaan-laboratorium", dependencies.PermintaanLaboratorium.Register},
		{"/berkas-digital", dependencies.BerkasDigital.Register},
		{"/resume-pasien", dependencies.ResumePasien.Register},
		{"/resume-pasien-ranap", dependencies.ResumePasienRanap.Register},
		{"/riwayat-perawatan", dependencies.RiwayatPerawatan.Register},
		{"/triase-igd", dependencies.TriaseIGD.Register},
		{"/awal-keperawatan-igd", dependencies.AwalKeperawatanIGD.Register},
		{"/awal-medis-umum", dependencies.AwalMedisUmum.Register},
		{"/awal-medis-ranap", dependencies.AwalMedisRanap.Register},
		{"/awal-medis-igd", dependencies.AwalMedisIgd.Register},
		{"/resep", dependencies.Resep.Register},
		{"/ventilator", dependencies.Ventilator.RegisterPasien},
	}
	aksesPelayananPasien := []string{"igd", "registrasi", "tindakan_ralan", "billing_ralan", "kamar_inap", "daftar_pasien_ranap"}
	for _, routeSidebar := range daftarRouteSidebar {
		group := protectedAPI.Group(routeSidebar.path)
		group.Use(dependencies.Autentikasi.WajibPermission(aksesPelayananPasien...))
		routeSidebar.register(group)
	}
	protectedAPI.GET("/user-management", dependencies.ManajemenPengguna.Daftar)
	protectedAPI.GET("/user-management/pegawai", dependencies.ManajemenPengguna.CariPegawai)
	protectedAPI.POST("/user-management", dependencies.ManajemenPengguna.Tambah)
	protectedAPI.PUT("/user-management/:id/akses", dependencies.ManajemenPengguna.UbahAkses)
}
