package penanganandokterpetugashttp

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"simrs-backend/internal/modules/autentikasi"
	modul "simrs-backend/internal/modules/penanganan_dokter_petugas"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct{ layanan *modul.Layanan }

func NewHandler(l *modul.Layanan) *Handler { return &Handler{layanan: l} }
func (h *Handler) Register(g *gin.RouterGroup) {
	g.GET("", h.Daftar)
	g.GET("/dokter", h.CariDokter)
	g.GET("/petugas", h.CariPetugas)
	g.GET("/tindakan", h.CariTindakan)
	g.POST("", h.Simpan)
	g.POST("/banyak", h.SimpanBanyak)
	g.PUT("", h.Ubah)
	g.DELETE("", h.Hapus)
}

func pengguna(c *gin.Context) (autentikasi.Pengguna, bool) {
	v, ok := c.Get("authenticated_user")
	if !ok {
		return autentikasi.Pengguna{}, false
	}
	u, ok := v.(autentikasi.Pengguna)
	return u, ok && strings.TrimSpace(u.Username) != ""
}
func (h *Handler) Daftar(c *gin.Context) {
	u, ok := pengguna(c)
	if !ok {
		h.unauth(c)
		return
	}
	data, err := h.layanan.Daftar(c.Request.Context(), u.Username, c.Query("no_rawat"), c.Query("jenis_rawat"))
	if err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}
func (h *Handler) CariDokter(c *gin.Context) {
	data, err := h.layanan.CariDokter(c.Request.Context(), c.Query("q"))
	if err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}
func (h *Handler) CariPetugas(c *gin.Context) {
	data, err := h.layanan.CariPetugas(c.Request.Context(), c.Query("q"))
	if err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}
func (h *Handler) CariTindakan(c *gin.Context) {
	data, err := h.layanan.CariTindakan(c.Request.Context(), c.Query("no_rawat"), c.Query("q"), c.Query("jenis_rawat"))
	if err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}
func (h *Handler) Simpan(c *gin.Context) {
	var input modul.Input
	if c.ShouldBindJSON(&input) != nil {
		h.validasi(c, "Data penanganan tidak lengkap")
		return
	}
	if err := h.layanan.Simpan(c.Request.Context(), input); err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusCreated, gin.H{"pesan": "Penanganan dokter dan petugas berhasil disimpan"})
}
func (h *Handler) SimpanBanyak(c *gin.Context) {
	var input modul.InputBanyak
	if c.ShouldBindJSON(&input) != nil {
		h.validasi(c, "Daftar tindakan tidak lengkap")
		return
	}
	if err := h.layanan.SimpanBanyak(c.Request.Context(), input); err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusCreated, gin.H{
		"pesan":           "Penanganan dokter dan petugas berhasil disimpan",
		"jumlah_disimpan": len(input.Catatan),
	})
}
func (h *Handler) Ubah(c *gin.Context) {
	var input modul.Input
	if c.ShouldBindJSON(&input) != nil {
		h.validasi(c, "Data penanganan tidak lengkap")
		return
	}
	if err := h.layanan.Ubah(c.Request.Context(), input); err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Penanganan dokter dan petugas berhasil diperbarui"})
}
func (h *Handler) Hapus(c *gin.Context) {
	var k modul.Kunci
	if c.ShouldBindJSON(&k) != nil {
		h.validasi(c, "Kunci penanganan tidak lengkap")
		return
	}
	if err := h.layanan.Hapus(c.Request.Context(), k); err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Penanganan dokter dan petugas berhasil dihapus"})
}
func (h *Handler) unauth(c *gin.Context) {
	httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
}
func (h *Handler) validasi(c *gin.Context, p string) {
	httpresponse.Error(c, http.StatusUnprocessableEntity, "PENANGANAN_VALIDATION_ERROR", p)
}
func (h *Handler) error(c *gin.Context, err error) {
	switch {
	case errors.Is(err, modul.ErrInputTidakValid):
		httpresponse.Error(c, http.StatusUnprocessableEntity, "PENANGANAN_VALIDATION_ERROR", err.Error())
	case errors.Is(err, modul.ErrBillingTerkunci):
		httpresponse.Error(c, http.StatusConflict, "PENANGANAN_BILLING_LOCKED", "Billing sudah terverifikasi atau kunjungan dibatalkan. Data tidak dapat diubah.")
	case errors.Is(err, modul.ErrTidakDitemukan):
		httpresponse.Error(c, http.StatusNotFound, "PENANGANAN_NOT_FOUND", err.Error())
	default:
		httpresponse.Error(c, http.StatusServiceUnavailable, "PENANGANAN_SIMRS_UNAVAILABLE", "Penanganan dokter dan petugas tidak dapat diproses pada SIMRS Khanza")
	}
}
