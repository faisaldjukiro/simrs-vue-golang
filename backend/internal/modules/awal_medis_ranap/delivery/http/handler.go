package awalmedisranaphttp

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	modul "simrs-backend/internal/modules/awal_medis_ranap"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct{ layanan *modul.Layanan }

func NewHandler(layanan *modul.Layanan) *Handler { return &Handler{layanan: layanan} }

func (h *Handler) Register(group *gin.RouterGroup) {
	group.GET("", h.Data)
	group.GET("/dokter", h.CariDokter)
	group.POST("", h.Simpan)
	group.PUT("", h.Ubah)
	group.DELETE("", h.Hapus)
}

func (h *Handler) Data(c *gin.Context) {
	data, err := h.layanan.Data(c.Request.Context(), c.Query("no_rawat"))
	if err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) CariDokter(c *gin.Context) {
	data, err := h.layanan.CariDokter(c.Request.Context(), c.Query("q"))
	if err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) Simpan(c *gin.Context) { h.simpan(c, false) }
func (h *Handler) Ubah(c *gin.Context)   { h.simpan(c, true) }

func (h *Handler) simpan(c *gin.Context, ubah bool) {
	var input modul.Input
	if c.ShouldBindJSON(&input) != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "awal_medis_ranap_VALIDATION_ERROR", "Data penilaian awal medis ranap tidak lengkap")
		return
	}
	if err := h.layanan.Simpan(c.Request.Context(), input, ubah); err != nil {
		h.tulisError(c, err)
		return
	}
	status, pesan := http.StatusCreated, "Penilaian awal medis ranap berhasil disimpan"
	if ubah {
		status, pesan = http.StatusOK, "Penilaian awal medis ranap berhasil diperbarui"
	}
	httpresponse.Success(c, status, gin.H{"pesan": pesan})
}

func (h *Handler) Hapus(c *gin.Context) {
	if err := h.layanan.Hapus(c.Request.Context(), c.Query("no_rawat")); err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Penilaian awal medis ranap berhasil dihapus"})
}

func (h *Handler) tulisError(c *gin.Context, err error) {
	log.Printf("penilaian awal medis ranap gagal diproses: %v", err)
	switch {
	case errors.Is(err, modul.ErrInputTidakValid):
		httpresponse.Error(c, http.StatusUnprocessableEntity, "awal_medis_ranap_VALIDATION_ERROR", err.Error())
	case errors.Is(err, modul.ErrSudahAda):
		httpresponse.Error(c, http.StatusConflict, "awal_medis_ranap_DUPLICATE", err.Error())
	case errors.Is(err, modul.ErrTidakDitemukan):
		httpresponse.Error(c, http.StatusNotFound, "awal_medis_ranap_NOT_FOUND", err.Error())
	case errors.Is(err, modul.ErrBillingTerkunci):
		httpresponse.Error(c, http.StatusConflict, "awal_medis_ranap_BILLING_LOCKED", err.Error())
	default:
		httpresponse.Error(c, http.StatusServiceUnavailable, "awal_medis_ranap_SIMRS_UNAVAILABLE", "Penilaian awal medis ranap tidak dapat diproses pada SIMRS Khanza")
	}
}
