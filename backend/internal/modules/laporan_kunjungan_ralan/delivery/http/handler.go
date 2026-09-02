package laporankunjunganralanhttp

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"simrs-backend/internal/modules/laporan_kunjungan_ralan"
	"simrs-backend/internal/shared/httpresponse"
	"strings"
	"time"
)

type Handler struct {
	repositori *laporan_kunjungan_ralan.Repositori
}

func NewHandler(r *laporan_kunjungan_ralan.Repositori) *Handler { return &Handler{repositori: r} }
func (h *Handler) Register(g *gin.RouterGroup)                  { g.GET("", h.daftar); g.GET("/referensi", h.referensi) }
func (h *Handler) referensi(c *gin.Context) {
	hasil, err := h.repositori.CariReferensi(c.Request.Context(), strings.TrimSpace(c.Query("jenis")), strings.TrimSpace(c.Query("q")))
	if err != nil {
		httpresponse.Error(c, http.StatusServiceUnavailable, "REFERENSI_SIMRS_UNAVAILABLE", "Referensi filter tidak dapat dibaca")
		return
	}
	httpresponse.Success(c, http.StatusOK, hasil)
}
func (h *Handler) daftar(c *gin.Context) {
	mulai, selesai := strings.TrimSpace(c.Query("tanggal_mulai")), strings.TrimSpace(c.Query("tanggal_selesai"))
	if !tanggalValid(mulai) || !tanggalValid(selesai) || mulai > selesai {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "LAPORAN_KUNJUNGAN_VALIDATION_ERROR", "Periode laporan tidak valid")
		return
	}
	jenis := strings.TrimSpace(c.DefaultQuery("jenis", "detail"))
	if jenis != "detail" && jenis != "rekap" && jenis != "berulang" {
		jenis = "detail"
	}
	hasil, err := h.repositori.Daftar(c.Request.Context(), laporan_kunjungan_ralan.Filter{TanggalMulai: mulai, TanggalSelesai: selesai, Jenis: jenis, Status: c.Query("status"), Poli: c.Query("poli"), Dokter: c.Query("dokter"), Penjamin: c.Query("penjamin"), Kabupaten: c.Query("kabupaten"), Kecamatan: c.Query("kecamatan"), Kelurahan: c.Query("kelurahan"), KataKunci: c.Query("q")})
	if err != nil {
		httpresponse.Error(c, http.StatusServiceUnavailable, "LAPORAN_KUNJUNGAN_SIMRS_UNAVAILABLE", "Laporan kunjungan rawat jalan tidak dapat dibaca")
		return
	}
	httpresponse.Success(c, http.StatusOK, hasil)
}
func tanggalValid(v string) bool { _, err := time.Parse("2006-01-02", v); return err == nil }
