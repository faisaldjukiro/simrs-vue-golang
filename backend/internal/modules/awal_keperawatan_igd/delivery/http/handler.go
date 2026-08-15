package awalkeperawatanigdhttp

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/autentikasi"
	modul "simrs-backend/internal/modules/awal_keperawatan_igd"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct{ layanan *modul.Layanan }

func NewHandler(layanan *modul.Layanan) *Handler { return &Handler{layanan: layanan} }
func (h *Handler) Register(group *gin.RouterGroup) {
	group.GET("", h.Data)
	group.POST("", h.Simpan)
	group.PUT("", h.Ubah)
	group.DELETE("", h.Hapus)
}

func (h *Handler) Data(c *gin.Context) {
	user, ok := penggunaLogin(c)
	if !ok {
		tidakLogin(c)
		return
	}
	data, err := h.layanan.Data(c.Request.Context(), c.Query("no_rawat"), user.Username)
	if err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}
func (h *Handler) Simpan(c *gin.Context) { h.simpan(c, false) }
func (h *Handler) Ubah(c *gin.Context)   { h.simpan(c, true) }
func (h *Handler) simpan(c *gin.Context, ubah bool) {
	if _, ok := penggunaLogin(c); !ok {
		tidakLogin(c)
		return
	}
	var input modul.Input
	if c.ShouldBindJSON(&input) != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "AWAL_KEPERAWATAN_IGD_VALIDATION_ERROR", "Data penilaian awal keperawatan IGD tidak lengkap")
		return
	}
	if err := h.layanan.Simpan(c.Request.Context(), input, ubah); err != nil {
		h.error(c, err)
		return
	}
	status, pesan := http.StatusCreated, "Penilaian awal keperawatan IGD berhasil disimpan"
	if ubah {
		status, pesan = http.StatusOK, "Penilaian awal keperawatan IGD berhasil diperbarui"
	}
	httpresponse.Success(c, status, gin.H{"pesan": pesan})
}
func (h *Handler) Hapus(c *gin.Context) {
	if _, ok := penggunaLogin(c); !ok {
		tidakLogin(c)
		return
	}
	if err := h.layanan.Hapus(c.Request.Context(), c.Query("no_rawat")); err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Penilaian awal keperawatan IGD berhasil dihapus"})
}
func (h *Handler) error(c *gin.Context, err error) {
	log.Printf("penilaian awal keperawatan IGD gagal diproses: %v", err)
	switch {
	case errors.Is(err, modul.ErrInputTidakValid):
		httpresponse.Error(c, http.StatusUnprocessableEntity, "AWAL_KEPERAWATAN_IGD_VALIDATION_ERROR", err.Error())
	case errors.Is(err, modul.ErrSudahAda):
		httpresponse.Error(c, http.StatusConflict, "AWAL_KEPERAWATAN_IGD_DUPLICATE", err.Error())
	case errors.Is(err, modul.ErrTidakDitemukan):
		httpresponse.Error(c, http.StatusNotFound, "AWAL_KEPERAWATAN_IGD_NOT_FOUND", err.Error())
	case errors.Is(err, modul.ErrBillingTerkunci):
		httpresponse.Error(c, http.StatusConflict, "AWAL_KEPERAWATAN_IGD_BILLING_LOCKED", err.Error())
	default:
		httpresponse.Error(c, http.StatusServiceUnavailable, "AWAL_KEPERAWATAN_IGD_SIMRS_UNAVAILABLE", "Data penilaian awal keperawatan IGD tidak dapat diproses pada SIMRS Khanza")
	}
}
func penggunaLogin(c *gin.Context) (autentikasi.Pengguna, bool) {
	value, ok := c.Get("authenticated_user")
	if !ok {
		return autentikasi.Pengguna{}, false
	}
	user, ok := value.(autentikasi.Pengguna)
	return user, ok && strings.TrimSpace(user.Username) != ""
}
func tidakLogin(c *gin.Context) {
	httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
}
