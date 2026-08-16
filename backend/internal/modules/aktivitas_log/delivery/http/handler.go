package aktivitasloghttp

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/aktivitas_log"
	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct {
	repositori *aktivitas_log.Repositori
}

func NewHandler(repositori *aktivitas_log.Repositori) *Handler {
	return &Handler{repositori: repositori}
}

func (h *Handler) Register(router *gin.RouterGroup) {
	router.GET("", h.daftar)
}

func (h *Handler) daftar(c *gin.Context) {
	pengguna, exists := c.Get(autentikasi.ContextKeyPengguna)
	user, valid := pengguna.(autentikasi.Pengguna)
	if !exists || !valid {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Silakan login terlebih dahulu")
		return
	}
	boleh, err := h.repositori.BolehMelihat(c.Request.Context(), user.ID)
	if err != nil {
		httpresponse.Error(c, http.StatusInternalServerError, "AKTIVITAS_LOG_UNAVAILABLE", "Hak akses log aktivitas tidak dapat diperiksa")
		return
	}
	if !boleh {
		httpresponse.Error(c, http.StatusForbidden, "FORBIDDEN", "Anda tidak memiliki akses melihat log aktivitas")
		return
	}

	hasil, err := h.repositori.Daftar(c.Request.Context(), aktivitas_log.Filter{
		Halaman:        angka(c.DefaultQuery("halaman", "1"), 1),
		Batas:          angka(c.DefaultQuery("batas", "25"), 25),
		KataKunci:      strings.TrimSpace(c.Query("kata_kunci")),
		Username:       strings.TrimSpace(c.Query("username")),
		Aksi:           strings.ToUpper(strings.TrimSpace(c.Query("aksi"))),
		Modul:          strings.TrimSpace(c.Query("modul")),
		Berhasil:       strings.TrimSpace(c.Query("berhasil")),
		TanggalMulai:   strings.TrimSpace(c.Query("tanggal_mulai")),
		TanggalSelesai: strings.TrimSpace(c.Query("tanggal_selesai")),
	})
	if err != nil {
		httpresponse.Error(c, http.StatusInternalServerError, "AKTIVITAS_LOG_UNAVAILABLE", "Log aktivitas tidak dapat dibaca")
		return
	}
	httpresponse.Success(c, http.StatusOK, hasil)
}

func angka(value string, fallback int) int {
	result, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return result
}
