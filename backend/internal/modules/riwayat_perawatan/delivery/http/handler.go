package riwayatperawatanhttp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	modul "simrs-backend/internal/modules/riwayat_perawatan"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct{ layanan *modul.Layanan }

func NewHandler(layanan *modul.Layanan) *Handler { return &Handler{layanan: layanan} }

func (h *Handler) Register(g *gin.RouterGroup) { g.GET("", h.Data) }

func (h *Handler) Data(c *gin.Context) {
	data, err := h.layanan.Data(c.Request.Context(), modul.Filter{
		NoRawat: c.Query("no_rawat"), Mode: c.Query("mode"),
		TanggalMulai: c.Query("tanggal_mulai"), TanggalSelesai: c.Query("tanggal_selesai"),
		NomorRawat: c.Query("nomor_rawat"),
	})
	if err != nil {
		if errors.Is(err, modul.ErrInputTidakValid) {
			httpresponse.Error(c, http.StatusUnprocessableEntity, "RIWAYAT_PERAWATAN_VALIDATION_ERROR", err.Error())
			return
		}
		httpresponse.Error(c, http.StatusServiceUnavailable, "RIWAYAT_PERAWATAN_SIMRS_UNAVAILABLE", "Riwayat perawatan tidak dapat dibaca dari SIMRS Khanza")
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}
