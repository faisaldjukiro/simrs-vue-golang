// Package faisal registers routes maintained by Faisal.
package faisal

import (
	"net/http"

	"github.com/gin-gonic/gin"

	autentikasihttp "simrs-backend/internal/modules/autentikasi/delivery/http"
	berandahttp "simrs-backend/internal/modules/beranda/delivery/http"
	cppthttp "simrs-backend/internal/modules/cppt/delivery/http"
	idrghttp "simrs-backend/internal/modules/idrg/delivery/http"
	manajemenpenggunahttp "simrs-backend/internal/modules/manajemen_pengguna/delivery/http"
	penanganandokterpetugashttp "simrs-backend/internal/modules/penanganan_dokter_petugas/delivery/http"
	permintaanradiologihttp "simrs-backend/internal/modules/permintaan_radiologi/delivery/http"
	riwayatperawatanhttp "simrs-backend/internal/modules/riwayat_perawatan/delivery/http"
	triaseigdhttp "simrs-backend/internal/modules/triase_igd/delivery/http"
	"simrs-backend/internal/shared/httpresponse"
)

type Dependencies struct {
	Autentikasi             *autentikasihttp.Handler
	Beranda                 *berandahttp.Handler
	IDRG                    *idrghttp.Handler
	ManajemenPengguna       *manajemenpenggunahttp.Handler
	CPPT                    *cppthttp.Handler
	PenangananDokterPetugas *penanganandokterpetugashttp.Handler
	PermintaanRadiologi     *permintaanradiologihttp.Handler
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
	protectedAPI.GET("/beranda", dependencies.Beranda.Beranda)
	dependencies.IDRG.Register(protectedAPI.Group("/idrg"))
	dependencies.CPPT.Register(protectedAPI.Group("/cppt"))
	dependencies.PenangananDokterPetugas.Register(protectedAPI.Group("/penanganan-dokter-petugas"))
	dependencies.PermintaanRadiologi.Register(protectedAPI.Group("/permintaan-radiologi"))
	dependencies.RiwayatPerawatan.Register(protectedAPI.Group("/riwayat-perawatan"))
	dependencies.TriaseIGD.Register(protectedAPI.Group("/triase-igd"))
	protectedAPI.GET("/user-management", dependencies.ManajemenPengguna.Daftar)
	protectedAPI.GET("/user-management/pegawai", dependencies.ManajemenPengguna.CariPegawai)
	protectedAPI.POST("/user-management", dependencies.ManajemenPengguna.Tambah)
	protectedAPI.PUT("/user-management/:id/akses", dependencies.ManajemenPengguna.UbahAkses)
}
