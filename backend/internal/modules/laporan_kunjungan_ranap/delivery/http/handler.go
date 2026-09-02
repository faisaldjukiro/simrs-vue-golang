package laporankunjunganranaphttp

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"simrs-backend/internal/modules/laporan_kunjungan_ranap"
	"simrs-backend/internal/shared/httpresponse"
	"strings"
	"time"
)

type Handler struct {
	r *laporan_kunjungan_ranap.Repositori
}

func NewHandler(r *laporan_kunjungan_ranap.Repositori) *Handler { return &Handler{r: r} }
func (h *Handler) Register(g *gin.RouterGroup)                  { g.GET("", h.daftar); g.GET("/referensi", h.ref) }
func (h *Handler) ref(c *gin.Context) {
	x, e := h.r.CariReferensi(c.Request.Context(), c.Query("jenis"), c.Query("q"))
	if e != nil {
		httpresponse.Error(c, 503, "REFERENSI_SIMRS_UNAVAILABLE", "Referensi filter tidak dapat dibaca")
		return
	}
	httpresponse.Success(c, 200, x)
}
func (h *Handler) daftar(c *gin.Context) {
	a, b := strings.TrimSpace(c.Query("tanggal_mulai")), strings.TrimSpace(c.Query("tanggal_selesai"))
	if _, e := time.Parse("2006-01-02", a); e != nil {
		httpresponse.Error(c, 422, "LAPORAN_RANAP_VALIDATION_ERROR", "Periode tidak valid")
		return
	}
	if _, e := time.Parse("2006-01-02", b); e != nil || a > b {
		httpresponse.Error(c, 422, "LAPORAN_RANAP_VALIDATION_ERROR", "Periode tidak valid")
		return
	}
	jenis := c.DefaultQuery("jenis", "masuk")
	if jenis != "masuk" && jenis != "pulang" && jenis != "berulang" {
		jenis = "masuk"
	}
	x, e := h.r.Daftar(c.Request.Context(), laporan_kunjungan_ranap.Filter{TanggalMulai: a, TanggalSelesai: b, Jenis: jenis, Status: c.Query("status"), Bangsal: c.Query("bangsal"), Dokter: c.Query("dokter"), Penjamin: c.Query("penjamin"), Kabupaten: c.Query("kabupaten"), Kecamatan: c.Query("kecamatan"), Kelurahan: c.Query("kelurahan"), KataKunci: c.Query("q")})
	if e != nil {
		httpresponse.Error(c, http.StatusServiceUnavailable, "LAPORAN_RANAP_SIMRS_UNAVAILABLE", "Laporan kunjungan rawat inap tidak dapat dibaca")
		return
	}
	httpresponse.Success(c, 200, x)
}
