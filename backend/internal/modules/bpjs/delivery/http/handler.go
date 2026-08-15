package bpjshttp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/bpjs"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct {
	layanan *bpjs.Layanan
}

func NewHandler(layanan *bpjs.Layanan) *Handler {
	return &Handler{layanan: layanan}
}

func (h *Handler) Register(g *gin.RouterGroup) {
	g.GET("/signature", h.Signature)
}

func (h *Handler) Signature(c *gin.Context) {
	hasil, err := h.layanan.BuatSignature()
	if err != nil {
		if errors.Is(err, bpjs.ErrKredensialTidakLengkap) {
			httpresponse.Error(c, http.StatusServiceUnavailable, "BPJS_CREDENTIALS_INCOMPLETE", err.Error())
			return
		}
		httpresponse.Error(c, http.StatusInternalServerError, "BPJS_SIGNATURE_ERROR", "Signature BPJS tidak dapat dibuat")
		return
	}

	httpresponse.Success(c, http.StatusOK, gin.H{
		"pesan": "Signature BPJS berhasil dibuat tanpa memanggil servis BPJS",
		"header": gin.H{
			"X-timestamp": hasil.Timestamp,
			"X-signature": hasil.Signature,
		},
		"algoritma":   hasil.Algoritma,
		"format_data": hasil.FormatData,
	})
}
