package cppthttp

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/modules/cppt"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct {
	layanan *cppt.Layanan
}

func NewHandler(layanan *cppt.Layanan) *Handler {
	return &Handler{layanan: layanan}
}

func (h *Handler) Register(group *gin.RouterGroup) {
	group.GET("", h.Daftar)
	group.GET("/petugas", h.CariPetugas)
	group.POST("", h.Simpan)
	group.PUT("", h.Ubah)
	group.DELETE("", h.Hapus)
}

func (h *Handler) CariPetugas(c *gin.Context) {
	user, ok := penggunaLogin(c)
	if !ok {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	petugas, err := h.layanan.CariPetugas(c.Request.Context(), user.ID, user.Username, c.Query("q"))
	if err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, petugas)
}

func (h *Handler) Daftar(c *gin.Context) {
	user, ok := penggunaLogin(c)
	if !ok {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	data, err := h.layanan.Daftar(c.Request.Context(), user.ID, user.Username, c.Query("no_rawat"), c.Query("jenis_rawat"))
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
	var input cppt.Input
	if err := c.ShouldBindJSON(&input); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "CPPT_VALIDATION_ERROR", "Data CPPT/SOAP tidak lengkap")
		return
	}
	if err := h.layanan.Simpan(c.Request.Context(), user.ID, user.Username, input); err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusCreated, gin.H{"pesan": "Catatan CPPT/SOAP berhasil disimpan"})
}

func (h *Handler) Ubah(c *gin.Context) {
	user, ok := penggunaLogin(c)
	if !ok {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	var input cppt.Input
	if err := c.ShouldBindJSON(&input); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "CPPT_VALIDATION_ERROR", "Data CPPT/SOAP tidak lengkap")
		return
	}
	if err := h.layanan.Ubah(c.Request.Context(), user.ID, user.Username, input); err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Catatan CPPT/SOAP berhasil diperbarui"})
}

func (h *Handler) Hapus(c *gin.Context) {
	user, ok := penggunaLogin(c)
	if !ok {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	var kunci cppt.Kunci
	if err := c.ShouldBindJSON(&kunci); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "CPPT_VALIDATION_ERROR", "Kunci catatan CPPT/SOAP tidak lengkap")
		return
	}
	if err := h.layanan.Hapus(c.Request.Context(), user.ID, user.Username, kunci); err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Catatan CPPT/SOAP berhasil dihapus"})
}

func (h *Handler) tulisError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, cppt.ErrInputTidakValid):
		httpresponse.Error(c, http.StatusUnprocessableEntity, "CPPT_VALIDATION_ERROR", err.Error())
	case errors.Is(err, cppt.ErrPetugasTidakDitemukan):
		httpresponse.Error(c, http.StatusUnprocessableEntity, "CPPT_PETUGAS_NOT_LINKED", err.Error())
	case errors.Is(err, cppt.ErrCatatanSudahAda):
		httpresponse.Error(c, http.StatusConflict, "CPPT_DUPLICATE", err.Error())
	case errors.Is(err, cppt.ErrCatatanTidakDitemukan):
		httpresponse.Error(c, http.StatusNotFound, "CPPT_NOT_FOUND", err.Error())
	case errors.Is(err, cppt.ErrTidakBerhak):
		httpresponse.Error(c, http.StatusForbidden, "CPPT_FORBIDDEN", err.Error())
	case errors.Is(err, cppt.ErrBillingTerkunci):
		httpresponse.Error(c, http.StatusConflict, "CPPT_BILLING_LOCKED", err.Error())
	default:
		httpresponse.Error(c, http.StatusServiceUnavailable, "CPPT_SIMRS_UNAVAILABLE", "Data CPPT/SOAP tidak dapat diproses pada SIMRS Khanza")
	}
}

func penggunaLogin(c *gin.Context) (autentikasi.Pengguna, bool) {
	value, exists := c.Get("authenticated_user")
	if !exists {
		return autentikasi.Pengguna{}, false
	}
	user, ok := value.(autentikasi.Pengguna)
	if !ok || strings.TrimSpace(user.Username) == "" {
		return autentikasi.Pengguna{}, false
	}
	return user, true
}
