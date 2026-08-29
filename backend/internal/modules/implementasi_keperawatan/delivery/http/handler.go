package implementasikeperawatanhttp

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/modules/implementasi_keperawatan"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct {
	layanan *implementasi_keperawatan.Layanan
}

func NewHandler(layanan *implementasi_keperawatan.Layanan) *Handler {
	return &Handler{layanan: layanan}
}

func (h *Handler) Register(group *gin.RouterGroup) {
	group.GET("", h.Daftar)
	group.GET("/petugas", h.CariPetugas)
	group.POST("", h.Simpan)
	group.PUT("", h.Ubah)
	group.DELETE("", h.Hapus)
}

func (h *Handler) Daftar(c *gin.Context) {
	user, ok := penggunaLogin(c)
	if !ok {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	data, err := h.layanan.Daftar(c.Request.Context(), user.ID, user.Username, c.Query("no_rawat"))
	if err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) CariPetugas(c *gin.Context) {
	user, ok := penggunaLogin(c)
	if !ok {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	data, err := h.layanan.CariPetugas(c.Request.Context(), user.ID, user.Username, c.Query("q"))
	if err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) Simpan(c *gin.Context) {
	user, ok := penggunaLogin(c)
	if !ok {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	var input implementasi_keperawatan.Input
	if err := c.ShouldBindJSON(&input); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "IMPLEMENTASI_KEPERAWATAN_VALIDATION_ERROR", "Data implementasi keperawatan tidak lengkap")
		return
	}
	if err := h.layanan.Simpan(c.Request.Context(), user.ID, user.Username, input); err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusCreated, gin.H{"pesan": "Implementasi keperawatan berhasil disimpan"})
}

func (h *Handler) Ubah(c *gin.Context) {
	user, ok := penggunaLogin(c)
	if !ok {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	var input implementasi_keperawatan.Input
	if err := c.ShouldBindJSON(&input); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "IMPLEMENTASI_KEPERAWATAN_VALIDATION_ERROR", "Data implementasi keperawatan tidak lengkap")
		return
	}
	if err := h.layanan.Ubah(c.Request.Context(), user.ID, user.Username, input); err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Implementasi keperawatan berhasil diperbarui"})
}

func (h *Handler) Hapus(c *gin.Context) {
	user, ok := penggunaLogin(c)
	if !ok {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	var kunci implementasi_keperawatan.Kunci
	if err := c.ShouldBindJSON(&kunci); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "IMPLEMENTASI_KEPERAWATAN_VALIDATION_ERROR", "Kunci implementasi keperawatan tidak lengkap")
		return
	}
	if err := h.layanan.Hapus(c.Request.Context(), user.ID, user.Username, kunci); err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Implementasi keperawatan berhasil dihapus"})
}

func (h *Handler) tulisError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, implementasi_keperawatan.ErrInputTidakValid):
		httpresponse.Error(c, http.StatusUnprocessableEntity, "IMPLEMENTASI_KEPERAWATAN_VALIDATION_ERROR", err.Error())
	case errors.Is(err, implementasi_keperawatan.ErrPetugasTidakDitemukan):
		httpresponse.Error(c, http.StatusUnprocessableEntity, "IMPLEMENTASI_KEPERAWATAN_PETUGAS_NOT_LINKED", err.Error())
	case errors.Is(err, implementasi_keperawatan.ErrCatatanSudahAda):
		httpresponse.Error(c, http.StatusConflict, "IMPLEMENTASI_KEPERAWATAN_DUPLICATE", err.Error())
	case errors.Is(err, implementasi_keperawatan.ErrCatatanTidakDitemukan):
		httpresponse.Error(c, http.StatusNotFound, "IMPLEMENTASI_KEPERAWATAN_NOT_FOUND", err.Error())
	case errors.Is(err, implementasi_keperawatan.ErrTidakBerhak):
		httpresponse.Error(c, http.StatusForbidden, "IMPLEMENTASI_KEPERAWATAN_FORBIDDEN", err.Error())
	case errors.Is(err, implementasi_keperawatan.ErrBillingTerkunci):
		httpresponse.Error(c, http.StatusConflict, "IMPLEMENTASI_KEPERAWATAN_BILLING_LOCKED", err.Error())
	default:
		httpresponse.Error(c, http.StatusServiceUnavailable, "IMPLEMENTASI_KEPERAWATAN_SIMRS_UNAVAILABLE", "Implementasi keperawatan tidak dapat diproses pada SIMRS Khanza")
	}
}

func penggunaLogin(c *gin.Context) (autentikasi.Pengguna, bool) {
	value, exists := c.Get("authenticated_user")
	if !exists {
		return autentikasi.Pengguna{}, false
	}
	user, ok := value.(autentikasi.Pengguna)
	return user, ok && strings.TrimSpace(user.Username) != ""
}
