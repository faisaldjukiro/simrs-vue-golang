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
	bpjshttp "simrs-backend/internal/modules/bpjs/delivery/http"
	dataklaimhttp "simrs-backend/internal/modules/bpjs/vclaim/monitoring/data_klaim/delivery/http"
	cppthttp "simrs-backend/internal/modules/cppt/delivery/http"
	diagnosapasienhttp "simrs-backend/internal/modules/diagnosa_pasien/delivery/http"
	idrghttp "simrs-backend/internal/modules/idrg/delivery/http"
	kelolamenuhttp "simrs-backend/internal/modules/kelola_menu/delivery/http"
	manajemenpenggunahttp "simrs-backend/internal/modules/manajemen_pengguna/delivery/http"
	penanganandokterpetugashttp "simrs-backend/internal/modules/penanganan_dokter_petugas/delivery/http"
	permintaanlaboratoriumhttp "simrs-backend/internal/modules/permintaan_laboratorium/delivery/http"
	permintaanradiologihttp "simrs-backend/internal/modules/permintaan_radiologi/delivery/http"
	resephttp "simrs-backend/internal/modules/resep/delivery/http"
	resumepasienranaphttp "simrs-backend/internal/modules/resume_pasien_ranap/delivery/http"
	riwayatperawatanhttp "simrs-backend/internal/modules/riwayat_perawatan/delivery/http"
	triaseigdhttp "simrs-backend/internal/modules/triase_igd/delivery/http"
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
	IDRG                    *idrghttp.Handler
	KelolaMenu              *kelolamenuhttp.Handler
	ManajemenPengguna       *manajemenpenggunahttp.Handler
	CPPT                    *cppthttp.Handler
	DiagnosaPasien          *diagnosapasienhttp.Handler
	PenangananDokterPetugas *penanganandokterpetugashttp.Handler
	PermintaanRadiologi     *permintaanradiologihttp.Handler
	PermintaanLaboratorium  *permintaanlaboratoriumhttp.Handler
	ResumePasienRanap       *resumepasienranaphttp.Handler
	Resep                   *resephttp.Handler
	RiwayatPerawatan        *riwayatperawatanhttp.Handler
	TriaseIGD               *triaseigdhttp.Handler
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
	dependencies.BPJS.Register(protectedAPI.Group("/bpjs"))
	dependencies.BPJSDataKlaim.Register(protectedAPI.Group("/bpjs/monitoring/klaim"))
	dependencies.IDRG.Register(protectedAPI.Group("/idrg"))
	kelolaMenuGroup := protectedAPI.Group("/kelola-menu")
	kelolaMenuGroup.Use(dependencies.Autentikasi.WajibPermission("kelola_menu"))
	dependencies.KelolaMenu.Register(kelolaMenuGroup)
	daftarRouteSidebar := []struct {
		path     string
		register func(*gin.RouterGroup)
	}{
		{"/cppt", dependencies.CPPT.Register},
		{"/diagnosa-pasien", dependencies.DiagnosaPasien.Register},
		{"/penanganan-dokter-petugas", dependencies.PenangananDokterPetugas.Register},
		{"/permintaan-radiologi", dependencies.PermintaanRadiologi.Register},
		{"/permintaan-laboratorium", dependencies.PermintaanLaboratorium.Register},
		{"/resume-pasien-ranap", dependencies.ResumePasienRanap.Register},
		{"/riwayat-perawatan", dependencies.RiwayatPerawatan.Register},
		{"/triase-igd", dependencies.TriaseIGD.Register},
		{"/awal-keperawatan-igd", dependencies.AwalKeperawatanIGD.Register},
		{"/awal-medis-umum", dependencies.AwalMedisUmum.Register},
		{"/awal-medis-ranap", dependencies.AwalMedisRanap.Register},
		{"/awal-medis-igd", dependencies.AwalMedisIgd.Register},
		{"/resep", dependencies.Resep.Register},
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
