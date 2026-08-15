package dataklaimhttp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	dataklaim "simrs-backend/internal/modules/bpjs/vclaim/monitoring/data_klaim"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct {
	layanan *dataklaim.Layanan
}

func NewHandler(layanan *dataklaim.Layanan) *Handler {
	return &Handler{layanan: layanan}
}

func (h *Handler) Register(g *gin.RouterGroup) {
	g.GET("", h.Data)
}

func (h *Handler) Data(c *gin.Context) {
	hasil, err := h.layanan.Data(c.Request.Context(), dataklaim.Input{
		TanggalPulang:  c.Query("tanggal_pulang"),
		TanggalMulai:   c.Query("tanggal_mulai"),
		TanggalSelesai: c.Query("tanggal_selesai"),
		JenisPelayanan: c.Query("jenis_pelayanan"),
		StatusKlaim:    c.Query("status_klaim"),
	})
	if err != nil {
		var kesalahanBPJS *dataklaim.KesalahanBPJS
		switch {
		case errors.Is(err, dataklaim.ErrInputTidakValid):
			httpresponse.Error(c, http.StatusUnprocessableEntity, "BPJS_MONITORING_VALIDATION_ERROR", err.Error())
		case errors.As(err, &kesalahanBPJS):
			httpresponse.Error(c, http.StatusBadGateway, "BPJS_VCLAIM_REJECTED", err.Error())
		case errors.Is(err, dataklaim.ErrVClaimTidakTersedia):
			httpresponse.Error(c, http.StatusServiceUnavailable, "BPJS_VCLAIM_UNAVAILABLE", err.Error())
		default:
			httpresponse.Error(c, http.StatusBadGateway, "BPJS_VCLAIM_RESPONSE_ERROR", err.Error())
		}
		return
	}

	httpresponse.Success(c, http.StatusOK, hasil)
}
