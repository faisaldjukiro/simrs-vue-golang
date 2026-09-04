package laporanpenggunaanbedhttp

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"simrs-backend/internal/modules/laporan_penggunaan_bed"
	"simrs-backend/internal/shared/httpresponse"
	"strings"
	"time"
)

type Handler struct {
	repositori *laporan_penggunaan_bed.Repositori
}

func NewHandler(r *laporan_penggunaan_bed.Repositori) *Handler { return &Handler{repositori: r} }
func (h *Handler) Register(g *gin.RouterGroup)                 { g.GET("", h.daftar); g.GET("/referensi", h.referensi) }
func (h *Handler) referensi(c *gin.Context) {
	x, e := h.repositori.CariBangsal(c.Request.Context(), c.Query("q"))
	if e != nil {
		httpresponse.Error(c, 503, "REFERENSI_BANGSAL_UNAVAILABLE", "Referensi bangsal tidak dapat dibaca")
		return
	}
	httpresponse.Success(c, 200, x)
}
func (h *Handler) daftar(c *gin.Context) {
	a, b := strings.TrimSpace(c.Query("tanggal_mulai")), strings.TrimSpace(c.Query("tanggal_selesai"))
	if !valid(a) || !valid(b) || a > b {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "LAPORAN_BED_VALIDATION_ERROR", "Periode laporan tidak valid")
		return
	}
	x, e := h.repositori.Daftar(c.Request.Context(), laporan_penggunaan_bed.Filter{
		TanggalMulai:   a,
		TanggalSelesai: b,
		Bangsal:        c.Query("bangsal"),
	})
	if e != nil {
		httpresponse.Error(c, 503, "LAPORAN_BED_UNAVAILABLE", "Laporan penggunaan bed tidak dapat dibaca")
		return
	}
	httpresponse.Success(c, 200, x)
}
func valid(v string) bool { _, e := time.Parse("2006-01-02", v); return e == nil }
