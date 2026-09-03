package laporan10penyakithttp

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/laporan_10_penyakit"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct {
	repositori *laporan_10_penyakit.Repositori
}

func NewHandler(repositori *laporan_10_penyakit.Repositori) *Handler {
	return &Handler{repositori: repositori}
}

func (h *Handler) Register(group *gin.RouterGroup) {
	group.GET("", h.daftar)
}

func (h *Handler) daftar(c *gin.Context) {
	mulai := strings.TrimSpace(c.Query("tanggal_mulai"))
	selesai := strings.TrimSpace(c.Query("tanggal_selesai"))
	if !tanggalValid(mulai) || !tanggalValid(selesai) || mulai > selesai {
		httpresponse.Error(c, http.StatusUnprocessableEntity,
			"LAPORAN_10_PENYAKIT_VALIDATION_ERROR", "Periode laporan tidak valid")
		return
	}

	status := strings.TrimSpace(c.DefaultQuery("status", "Ralan"))
	if status != "Semua" && status != "Ralan" && status != "Ranap" {
		status = "Ralan"
	}

	hasil, err := h.repositori.Daftar(c.Request.Context(), laporan_10_penyakit.Filter{
		TanggalMulai:   mulai,
		TanggalSelesai: selesai,
		Status:         status,
		KataKunci:      c.Query("q"),
	})
	if err != nil {
		httpresponse.Error(c, http.StatusServiceUnavailable,
			"LAPORAN_10_PENYAKIT_SIMRS_UNAVAILABLE",
			"Laporan 10 penyakit tidak dapat dibaca")
		return
	}

	httpresponse.Success(c, http.StatusOK, hasil)
}

func tanggalValid(value string) bool {
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}
