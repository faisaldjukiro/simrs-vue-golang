package resumepasienhttp

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	modul "simrs-backend/internal/modules/resume_pasien"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct{ layanan *modul.Layanan }

func NewHandler(layanan *modul.Layanan) *Handler { return &Handler{layanan: layanan} }

func (h *Handler) Register(group *gin.RouterGroup) {
	group.GET("", h.Data)
	group.GET("/validasi-coding", h.ValidasiCoding)
	group.GET("/cari-coding", h.CariCoding)
	group.GET("/referensi", h.Referensi)
	group.POST("", h.Simpan)
	group.PUT("", h.Ubah)
	group.DELETE("", h.Hapus)
}

func (h *Handler) Referensi(c *gin.Context) {
	data, err := h.layanan.Referensi(
		c.Request.Context(),
		c.Query("no_rawat"),
		c.Query("jenis"),
		c.Query("q"),
	)
	if err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) CariCoding(c *gin.Context) {
	data, err := h.layanan.CariCoding(
		c.Request.Context(),
		c.Query("jenis"),
		c.Query("q"),
		c.Query("utama") == "true" || c.Query("utama") == "1",
	)
	if err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) ValidasiCoding(c *gin.Context) {
	data, err := h.layanan.ValidasiCoding(
		c.Request.Context(),
		c.Query("jenis"),
		c.Query("kode"),
		c.Query("utama") == "true" || c.Query("utama") == "1",
	)
	if err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) Data(c *gin.Context) {
	data, err := h.layanan.Data(c.Request.Context(), c.Query("no_rawat"))
	if err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) Simpan(c *gin.Context) { h.simpan(c, false) }
func (h *Handler) Ubah(c *gin.Context)   { h.simpan(c, true) }

func (h *Handler) simpan(c *gin.Context, ubah bool) {
	var input modul.Input
	if c.ShouldBindJSON(&input) != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "RESUME_PASIEN_VALIDATION_ERROR", "Data resume pasien rawat jalan tidak lengkap")
		return
	}
	var err error
	if ubah {
		err = h.layanan.Ubah(c.Request.Context(), input)
	} else {
		err = h.layanan.Simpan(c.Request.Context(), input)
	}
	if err != nil {
		h.error(c, err)
		return
	}
	status, pesan := http.StatusCreated, "Resume pasien rawat jalan berhasil disimpan"
	if ubah {
		status, pesan = http.StatusOK, "Resume pasien rawat jalan berhasil diperbarui"
	}
	httpresponse.Success(c, status, gin.H{"pesan": pesan})
}

func (h *Handler) Hapus(c *gin.Context) {
	if err := h.layanan.Hapus(c.Request.Context(), c.Query("no_rawat")); err != nil {
		h.error(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Resume pasien rawat jalan berhasil dihapus"})
}

func (h *Handler) error(c *gin.Context, err error) {
	log.Printf("resume pasien rawat jalan gagal diproses: %v", err)
	switch {
	case errors.Is(err, modul.ErrInputTidakValid):
		httpresponse.Error(c, http.StatusUnprocessableEntity, "RESUME_PASIEN_VALIDATION_ERROR", err.Error())
	case errors.Is(err, modul.ErrSudahAda):
		httpresponse.Error(c, http.StatusConflict, "RESUME_PASIEN_DUPLICATE", err.Error())
	case errors.Is(err, modul.ErrBillingTerkunci):
		httpresponse.Error(c, http.StatusConflict, "RESUME_PASIEN_BILLING_LOCKED", "Billing sudah terverifikasi atau kunjungan dibatalkan. Resume tidak dapat ditambah, diubah, atau dihapus.")
	case errors.Is(err, modul.ErrTidakDitemukan):
		httpresponse.Error(c, http.StatusNotFound, "RESUME_PASIEN_NOT_FOUND", err.Error())
	default:
		httpresponse.Error(c, http.StatusServiceUnavailable, "RESUME_PASIEN_SIMRS_UNAVAILABLE", "Resume pasien rawat jalan tidak dapat diproses pada SIMRS Khanza")
	}
}
