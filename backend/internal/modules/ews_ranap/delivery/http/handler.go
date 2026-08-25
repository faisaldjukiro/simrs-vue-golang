package ewsranaphttp

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/modules/ews_ranap"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct {
	layanan *ews_ranap.Layanan
}

func NewHandler(layanan *ews_ranap.Layanan) *Handler {
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
	petugas, err := h.layanan.CariPetugas(c.Request.Context(), user.ID, user.Username, c.Query("q"))
	if err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, petugas)
}

func (h *Handler) Simpan(c *gin.Context) {
	user, ok := penggunaLogin(c)
	if !ok {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	var input ews_ranap.Input
	if err := c.ShouldBindJSON(&input); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "EWS_RANAP_VALIDATION_ERROR", "Data EWS Ranap tidak lengkap")
		return
	}
	if err := h.layanan.Simpan(c.Request.Context(), user.ID, user.Username, input); err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusCreated, gin.H{"pesan": "EWS Ranap berhasil disimpan"})
}

func (h *Handler) Ubah(c *gin.Context) {
	user, ok := penggunaLogin(c)
	if !ok {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	var input ews_ranap.Input
	if err := c.ShouldBindJSON(&input); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "EWS_RANAP_VALIDATION_ERROR", "Data EWS Ranap tidak lengkap")
		return
	}
	if err := h.layanan.Ubah(c.Request.Context(), user.ID, user.Username, input); err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "EWS Ranap berhasil diperbarui"})
}

func (h *Handler) Hapus(c *gin.Context) {
	user, ok := penggunaLogin(c)
	if !ok {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	var kunci ews_ranap.Kunci
	if err := c.ShouldBindJSON(&kunci); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "EWS_RANAP_VALIDATION_ERROR", "Kunci EWS Ranap tidak lengkap")
		return
	}
	if err := h.layanan.Hapus(c.Request.Context(), user.ID, user.Username, kunci); err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "EWS Ranap berhasil dihapus"})
}

func (h *Handler) tulisError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ews_ranap.ErrInputTidakValid):
		httpresponse.Error(c, http.StatusUnprocessableEntity, "EWS_RANAP_VALIDATION_ERROR", err.Error())
	case errors.Is(err, ews_ranap.ErrPetugasTidakDitemukan):
		httpresponse.Error(c, http.StatusUnprocessableEntity, "EWS_RANAP_PETUGAS_NOT_LINKED", err.Error())
	case errors.Is(err, ews_ranap.ErrCatatanSudahAda):
		httpresponse.Error(c, http.StatusConflict, "EWS_RANAP_DUPLICATE", err.Error())
	case errors.Is(err, ews_ranap.ErrCatatanTidakDitemukan):
		httpresponse.Error(c, http.StatusNotFound, "EWS_RANAP_NOT_FOUND", err.Error())
	case errors.Is(err, ews_ranap.ErrTidakBerhak):
		httpresponse.Error(c, http.StatusForbidden, "EWS_RANAP_FORBIDDEN", err.Error())
	case errors.Is(err, ews_ranap.ErrBillingTerkunci):
		httpresponse.Error(c, http.StatusConflict, "EWS_RANAP_BILLING_LOCKED", err.Error())
	default:
		httpresponse.Error(c, http.StatusServiceUnavailable, "EWS_RANAP_SIMRS_UNAVAILABLE", "EWS Ranap tidak dapat diproses pada SIMRS Khanza")
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
