package resephttp

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	modul "simrs-backend/internal/modules/resep"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct{ layanan *modul.Layanan }

func NewHandler(layanan *modul.Layanan) *Handler { return &Handler{layanan: layanan} }
func (h *Handler) Register(g *gin.RouterGroup) {
	g.GET("", h.DaftarResep)
	g.GET("/copy", h.DaftarResepPasien)
	g.GET("/detail/:no_resep", h.DetailResep)
	g.GET("/cari-obat", h.CariObat)
	g.GET("/cari-dokter", h.CariDokter)
	g.GET("/cari-depo", h.CariDepoGudang)
	g.GET("/metode-racik", h.DaftarMetodeRacik)
	g.GET("/nomor-auto", h.AutoNomorResep)
	g.GET("/info-pasien", h.InfoPasien)
	g.GET("/depo-default", h.DepoDefault)
	g.POST("", h.SimpanResep)
	g.DELETE("/:no_resep", h.HapusResep)
}
func (h *Handler) DaftarResep(c *gin.Context) {
	d, e := h.layanan.DaftarResep(c.Request.Context(), dq(c, "no_rawat"))
	if e != nil {
		h.error(c, e)
		return
	}
	httpresponse.Success(c, http.StatusOK, d)
}
func (h *Handler) DaftarResepPasien(c *gin.Context) {
	d, e := h.layanan.DaftarResepPasien(c.Request.Context(), dq(c, "no_rkm_medis"))
	if e != nil {
		h.error(c, e)
		return
	}
	httpresponse.Success(c, http.StatusOK, d)
}
func (h *Handler) DetailResep(c *gin.Context) {
	d, e := h.layanan.DetailResep(c.Request.Context(), c.Param("no_resep"))
	if e != nil {
		h.error(c, e)
		return
	}
	httpresponse.Success(c, http.StatusOK, d)
}
func (h *Handler) CariObat(c *gin.Context) {
	d, e := h.layanan.CariObat(c.Request.Context(), c.Query("q"), c.Query("kd_bangsal"), c.Query("kelas"))
	if e != nil {
		h.error(c, e)
		return
	}
	httpresponse.Success(c, http.StatusOK, d)
}
func (h *Handler) CariDokter(c *gin.Context) {
	d, e := h.layanan.CariDokter(c.Request.Context(), c.Query("q"))
	if e != nil {
		h.error(c, e)
		return
	}
	httpresponse.Success(c, http.StatusOK, d)
}
func (h *Handler) CariDepoGudang(c *gin.Context) {
	d, e := h.layanan.CariDepoGudang(c.Request.Context(), c.Query("q"))
	if e != nil {
		h.error(c, e)
		return
	}
	httpresponse.Success(c, http.StatusOK, d)
}
func (h *Handler) DaftarMetodeRacik(c *gin.Context) {
	d, e := h.layanan.DaftarMetodeRacik(c.Request.Context())
	if e != nil {
		h.error(c, e)
		return
	}
	httpresponse.Success(c, http.StatusOK, d)
}
func (h *Handler) AutoNomorResep(c *gin.Context) {
	d, e := h.layanan.AutoNomorResep(c.Request.Context(), c.Query("tanggal"))
	if e != nil {
		h.error(c, e)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"no_resep": d})
}
func (h *Handler) InfoPasien(c *gin.Context) {
	d, e := h.layanan.InfoPasien(c.Request.Context(), c.Query("no_rawat"))
	if e != nil {
		h.error(c, e)
		return
	}
	httpresponse.Success(c, http.StatusOK, d)
}
func (h *Handler) DepoDefault(c *gin.Context) {
	kd, nm := h.layanan.DepoDefault(c.Request.Context(), c.Query("no_rawat"), c.Query("status"))
	httpresponse.Success(c, http.StatusOK, gin.H{"kd_bangsal": kd, "nm_bangsal": nm})
}
func (h *Handler) SimpanResep(c *gin.Context) {
	var in modul.InputSimpanResep
	if e := c.ShouldBindJSON(&in); e != nil {
		httpresponse.Error(c, 422, "RESEP_VALIDATION_ERROR", "Data resep tidak lengkap atau format salah")
		return
	}
	if e := h.layanan.SimpanResep(c.Request.Context(), in); e != nil {
		h.error(c, e)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Resep berhasil disimpan"})
}
func (h *Handler) HapusResep(c *gin.Context) {
	if e := h.layanan.HapusResep(c.Request.Context(), c.Param("no_resep")); e != nil {
		h.error(c, e)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Resep berhasil dihapus"})
}
func dq(c *gin.Context, k string) string { return c.Query(k) }
func (h *Handler) error(c *gin.Context, e error) {
	log.Printf("resep gagal diproses: %v", e)
	switch {
	case errors.Is(e, modul.ErrInputTidakValid):
		httpresponse.Error(c, 422, "RESEP_VALIDATION_ERROR", e.Error())
	case errors.Is(e, modul.ErrResepTidakAda):
		httpresponse.Error(c, 404, "RESEP_NOT_FOUND", "Resep atau pasien tidak ditemukan")
	case errors.Is(e, modul.ErrBillingTerkunci):
		httpresponse.Error(c, 409, "BILLING_TERKUNCI", "Resep tidak dapat diubah karena billing sudah diproses")
	default:
		httpresponse.Error(c, 503, "RESEP_UNAVAILABLE", "Resep tidak dapat diproses pada SIMRS Khanza: "+e.Error())
	}
}
