package triaseigdhttp

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/autentikasi"
	modul "simrs-backend/internal/modules/triase_igd"
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
		h.tidakLogin(c)
		return
	}
	data, err := h.layanan.Data(c.Request.Context(), c.Query("no_rawat"), user.Username)
	if err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) Simpan(c *gin.Context) { h.simpan(c, false) }
func (h *Handler) Ubah(c *gin.Context)   { h.simpan(c, true) }

func (h *Handler) simpan(c *gin.Context, ubah bool) {
	if _, ok := penggunaLogin(c); !ok {
		h.tidakLogin(c)
		return
	}
	var input modul.Input
	if c.ShouldBindJSON(&input) != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "TRIASE_IGD_VALIDATION_ERROR", "Data triase IGD tidak lengkap")
		return
	}
	if err := h.layanan.Simpan(c.Request.Context(), input, ubah); err != nil {
		h.tulisError(c, err)
		return
	}
	status, pesan := http.StatusCreated, "Triase IGD berhasil disimpan"
	if ubah {
		status, pesan = http.StatusOK, "Triase IGD berhasil diperbarui"
	}
	httpresponse.Success(c, status, gin.H{"pesan": pesan})
}

func (h *Handler) Hapus(c *gin.Context) {
	if _, ok := penggunaLogin(c); !ok {
		h.tidakLogin(c)
		return
	}
	if err := h.layanan.Hapus(c.Request.Context(), c.Query("no_rawat"), c.Query("jenis")); err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Data triase IGD berhasil dihapus"})
}

func (h *Handler) tulisError(c *gin.Context, err error) {
	log.Printf("triase IGD gagal diproses: %v", err)
	switch {
	case errors.Is(err, modul.ErrInputTidakValid):
		httpresponse.Error(c, http.StatusUnprocessableEntity, "TRIASE_IGD_VALIDATION_ERROR", err.Error())
	case errors.Is(err, modul.ErrSudahAda):
		httpresponse.Error(c, http.StatusConflict, "TRIASE_IGD_DUPLICATE", err.Error())
	case errors.Is(err, modul.ErrTidakDitemukan):
		httpresponse.Error(c, http.StatusNotFound, "TRIASE_IGD_NOT_FOUND", err.Error())
	default:
		httpresponse.Error(c, http.StatusServiceUnavailable, "TRIASE_IGD_SIMRS_UNAVAILABLE", "Data triase IGD tidak dapat diproses pada SIMRS Khanza")
	}
}

func (h *Handler) tidakLogin(c *gin.Context) {
	httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
}

func penggunaLogin(c *gin.Context) (autentikasi.Pengguna, bool) {
	value, exists := c.Get("authenticated_user")
	if !exists {
		return autentikasi.Pengguna{}, false
	}
	user, ok := value.(autentikasi.Pengguna)
	return user, ok && strings.TrimSpace(user.Username) != ""
}
