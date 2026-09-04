package laporanborlostoihttp

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/laporan_bor_los_toi"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct {
	repositori *laporan_bor_los_toi.Repositori
}

func NewHandler(repositori *laporan_bor_los_toi.Repositori) *Handler {
	return &Handler{repositori: repositori}
}

func (h *Handler) Register(group *gin.RouterGroup) {
	group.GET("", h.daftar)
}

func (h *Handler) daftar(c *gin.Context) {
	mulai, errMulai := parseBulan(c.Query("bulan_mulai"))
	selesai, errSelesai := parseBulan(c.Query("bulan_selesai"))
	if errMulai != nil || errSelesai != nil || mulai.After(selesai) {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "LAPORAN_BOR_LOS_TOI_VALIDATION_ERROR", "Periode bulan tidak valid")
		return
	}
	if selesai.Year()*12+int(selesai.Month())-(mulai.Year()*12+int(mulai.Month())) > 59 {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "LAPORAN_BOR_LOS_TOI_PERIOD_TOO_LONG", "Periode laporan maksimal 60 bulan")
		return
	}

	hasil, err := h.repositori.Daftar(c.Request.Context(), mulai, selesai)
	if err != nil {
		httpresponse.Error(c, http.StatusServiceUnavailable, "LAPORAN_BOR_LOS_TOI_UNAVAILABLE", "Laporan BOR, LOS, dan TOI tidak dapat dibaca")
		return
	}
	httpresponse.Success(c, http.StatusOK, hasil)
}

func parseBulan(nilai string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(nilai)+"-01")
}
