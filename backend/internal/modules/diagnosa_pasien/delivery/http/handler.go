package diagnosapasienhttp

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	modul "simrs-backend/internal/modules/diagnosa_pasien"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct{ layanan *modul.Layanan }

func NewHandler(layanan *modul.Layanan) *Handler { return &Handler{layanan: layanan} }

func (h *Handler) Register(group *gin.RouterGroup) {
	group.GET("", h.Data)
	group.GET("/cari-coding", h.CariCoding)
	group.PUT("", h.Simpan)
	group.DELETE("/:jenis/:kode", h.Hapus)
}

func (h *Handler) Hapus(c *gin.Context) {
	input := modul.HapusInput{
		NoRawat: c.Query("no_rawat"),
		Status:  c.Query("status"),
		Jenis:   c.Param("jenis"),
		Kode:    c.Param("kode"),
	}
	if err := h.layanan.Hapus(c.Request.Context(), input); err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Coding pasien berhasil dihapus"})
}

func (h *Handler) Data(c *gin.Context) {
	data, err := h.layanan.Data(c.Request.Context(), c.Query("no_rawat"), c.Query("status"))
	if err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) CariCoding(c *gin.Context) {
	data, err := h.layanan.CariCoding(
		c.Request.Context(), c.Query("jenis"), c.Query("q"),
		c.Query("utama") == "1" || c.Query("utama") == "true",
	)
	if err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) Simpan(c *gin.Context) {
	var input modul.Input
	if err := c.ShouldBindJSON(&input); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "DIAGNOSA_PASIEN_VALIDATION_ERROR", "Data diagnosa pasien tidak lengkap")
		return
	}
	if err := h.layanan.Simpan(c.Request.Context(), input); err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Diagnosa dan prosedur pasien berhasil disimpan"})
}

func (h *Handler) error(c *gin.Context, err error) {
	log.Printf("diagnosa pasien gagal diproses: %v", err)
	switch {
	case errors.Is(err, modul.ErrInputTidakValid):
		httpresponse.Error(c, http.StatusUnprocessableEntity, "DIAGNOSA_PASIEN_VALIDATION_ERROR", err.Error())
	case errors.Is(err, modul.ErrBillingTerkunci):
		httpresponse.Error(c, http.StatusConflict, "DIAGNOSA_PASIEN_BILLING_LOCKED", "Billing sudah terverifikasi atau kunjungan dibatalkan. Diagnosa dan prosedur hanya dapat dilihat.")
	case errors.Is(err, modul.ErrCodingDuplikat):
		httpresponse.Error(c, http.StatusConflict, "DIAGNOSA_PASIEN_DUPLICATE", err.Error())
	case errors.Is(err, modul.ErrCodingTidakAda):
		httpresponse.Error(c, http.StatusNotFound, "DIAGNOSA_PASIEN_CODING_NOT_FOUND", err.Error())
	case errors.Is(err, modul.ErrKunjunganInvalid):
		httpresponse.Error(c, http.StatusNotFound, "DIAGNOSA_PASIEN_NOT_FOUND", err.Error())
	default:
		httpresponse.Error(c, http.StatusServiceUnavailable, "DIAGNOSA_PASIEN_SIMRS_UNAVAILABLE", "Diagnosa pasien tidak dapat diproses pada SIMRS Khanza")
	}
}
