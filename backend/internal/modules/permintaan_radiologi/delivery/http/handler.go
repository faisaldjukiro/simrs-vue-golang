package permintaanradiologihttp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	modul "simrs-backend/internal/modules/permintaan_radiologi"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct{ layanan *modul.Layanan }

func NewHandler(layanan *modul.Layanan) *Handler { return &Handler{layanan: layanan} }

func (h *Handler) Register(g *gin.RouterGroup) {
	g.GET("", h.Data)
	g.GET("/dokter", h.CariDokter)
	g.GET("/tindakan", h.CariTindakan)
	g.POST("", h.Simpan)
	g.PUT("/:nomor", h.Ubah)
	g.DELETE("/:nomor", h.Hapus)
}

func (h *Handler) Data(c *gin.Context) {
	data, err := h.layanan.Data(c.Request.Context(), c.Query("no_rawat"))
	if err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) CariDokter(c *gin.Context) {
	data, err := h.layanan.CariDokter(c.Request.Context(), c.Query("q"))
	if err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) CariTindakan(c *gin.Context) {
	data, err := h.layanan.CariTindakan(c.Request.Context(), c.Query("no_rawat"), c.Query("q"))
	if err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) Simpan(c *gin.Context) {
	var input modul.Input
	if c.ShouldBindJSON(&input) != nil {
		h.validasi(c, "Data permintaan radiologi tidak lengkap")
		return
	}
	nomor, err := h.layanan.Simpan(c.Request.Context(), input)
	if err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusCreated, gin.H{
		"pesan":           "Permintaan radiologi berhasil disimpan",
		"nomor":           nomor,
		"jumlah_tindakan": len(input.KodeTindakan),
	})
}

func (h *Handler) Ubah(c *gin.Context) {
	var input modul.Input
	if c.ShouldBindJSON(&input) != nil {
		h.validasi(c, "Data perubahan permintaan radiologi tidak lengkap")
		return
	}
	if err := h.layanan.Ubah(c.Request.Context(), c.Param("nomor"), input); err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{
		"pesan":           "Permintaan radiologi berhasil diperbarui",
		"nomor":           c.Param("nomor"),
		"jumlah_tindakan": len(input.KodeTindakan),
	})
}

func (h *Handler) Hapus(c *gin.Context) {
	if err := h.layanan.Hapus(c.Request.Context(), c.Query("no_rawat"), c.Param("nomor")); err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Permintaan radiologi berhasil dihapus"})
}

func (h *Handler) validasi(c *gin.Context, pesan string) {
	httpresponse.Error(c, http.StatusUnprocessableEntity, "PERMINTAAN_RADIOLOGI_VALIDATION_ERROR", pesan)
}

func (h *Handler) error(c *gin.Context, err error) {
	switch {
	case errors.Is(err, modul.ErrInputTidakValid):
		httpresponse.Error(c, http.StatusUnprocessableEntity, "PERMINTAAN_RADIOLOGI_VALIDATION_ERROR", err.Error())
	case errors.Is(err, modul.ErrBillingTerkunci):
		httpresponse.Error(c, http.StatusConflict, "PERMINTAAN_RADIOLOGI_BILLING_LOCKED", "Billing sudah terverifikasi atau kunjungan dibatalkan. Data tidak dapat diubah.")
	case errors.Is(err, modul.ErrSudahDiproses):
		httpresponse.Error(c, http.StatusConflict, "PERMINTAAN_RADIOLOGI_PAID", "Permintaan sudah dibayar di kasir dan tidak dapat diubah atau dihapus.")
	case errors.Is(err, modul.ErrSudahDiterima):
		httpresponse.Error(c, http.StatusConflict, "PERMINTAAN_RADIOLOGI_ACCEPTED", "Permintaan sudah diterima petugas radiologi dan tidak dapat diubah atau dihapus.")
	case errors.Is(err, modul.ErrTidakDitemukan):
		httpresponse.Error(c, http.StatusNotFound, "PERMINTAAN_RADIOLOGI_NOT_FOUND", err.Error())
	default:
		httpresponse.Error(c, http.StatusServiceUnavailable, "PERMINTAAN_RADIOLOGI_SIMRS_UNAVAILABLE", "Permintaan radiologi tidak dapat diproses pada SIMRS Khanza")
	}
}
