package whatsappgatewayhttp

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/whatsapp_gateway"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct{ layanan *whatsapp_gateway.Layanan }

func NewHandler(layanan *whatsapp_gateway.Layanan) *Handler { return &Handler{layanan: layanan} }

func (h *Handler) Register(g *gin.RouterGroup) {
	g.GET("/status", h.Status)
	g.GET("/perangkat", h.DaftarPerangkat)
	g.POST("/perangkat", h.BuatPerangkat)
	g.GET("/perangkat/:id", h.DetailPerangkat)
	g.POST("/perangkat/:id/hubungkan", h.HubungkanPerangkat)
	g.GET("/perangkat/:id/qr", h.QR)
	g.POST("/perangkat/:id/kode-pasangan", h.KodePasangan)
	g.GET("/pesan", h.DaftarPesan)
	g.POST("/pesan", h.KirimPesan)
}

func (h *Handler) Status(c *gin.Context) {
	httpresponse.Success(c, http.StatusOK, h.layanan.StatusKonfigurasi())
}

func (h *Handler) DaftarPerangkat(c *gin.Context) {
	hasil, err := h.layanan.DaftarPerangkat(c.Request.Context())
	h.jawab(c, hasil, err, http.StatusOK)
}

func (h *Handler) BuatPerangkat(c *gin.Context) {
	var input struct {
		Nama string `json:"nama"`
	}
	if c.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.Nama) == "" || len(strings.TrimSpace(input.Nama)) > 100 {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "WHATSAPP_VALIDATION_ERROR", "Nama perangkat wajib diisi maksimal 100 karakter")
		return
	}
	hasil, err := h.layanan.BuatPerangkat(c.Request.Context(), strings.TrimSpace(input.Nama))
	h.jawab(c, hasil, err, http.StatusCreated)
}

func (h *Handler) DetailPerangkat(c *gin.Context) {
	hasil, err := h.layanan.DetailPerangkat(c.Request.Context(), c.Param("id"))
	h.jawab(c, hasil, err, http.StatusOK)
}

func (h *Handler) HubungkanPerangkat(c *gin.Context) {
	hasil, err := h.layanan.HubungkanPerangkat(c.Request.Context(), c.Param("id"))
	h.jawab(c, hasil, err, http.StatusAccepted)
}

func (h *Handler) QR(c *gin.Context) {
	gambar, kedaluwarsa, err := h.layanan.QR(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.jawab(c, nil, err, http.StatusOK)
		return
	}
	if kedaluwarsa != "" {
		c.Header("X-QR-Expires-At", kedaluwarsa)
	}
	c.Data(http.StatusOK, "image/png", gambar)
}

func (h *Handler) KodePasangan(c *gin.Context) {
	var input struct {
		Nomor string `json:"nomor"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "WHATSAPP_VALIDATION_ERROR", "Nomor WhatsApp pengirim wajib diisi")
		return
	}
	nomor := whatsapp_gateway.NormalisasiNomor(input.Nomor)
	if !strings.HasPrefix(nomor, "62") || len(nomor) < 10 {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "WHATSAPP_PHONE_INVALID", "Nomor WhatsApp harus menggunakan format 62 tanpa tanda +")
		return
	}
	hasil, err := h.layanan.KodePasangan(c.Request.Context(), c.Param("id"), nomor)
	h.jawab(c, hasil, err, http.StatusOK)
}

func (h *Handler) DaftarPesan(c *gin.Context) {
	hasil, err := h.layanan.DaftarPesan(c.Request.Context(), c.Query("device_id"), c.Query("status"), 25)
	h.jawab(c, hasil, err, http.StatusOK)
}

func (h *Handler) KirimPesan(c *gin.Context) {
	var input struct {
		DeviceID     string `json:"device_id"`
		Tujuan       string `json:"tujuan"`
		NamaPenerima string `json:"nama_penerima"`
		ExternalID   string `json:"external_id"`
		Idempotency  string `json:"idempotency_key"`
		Pesan        string `json:"pesan"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "WHATSAPP_VALIDATION_ERROR", "Data pesan WhatsApp tidak valid")
		return
	}
	input.DeviceID = strings.TrimSpace(input.DeviceID)
	input.Tujuan = whatsapp_gateway.NormalisasiNomor(input.Tujuan)
	input.Pesan = strings.TrimSpace(input.Pesan)
	if input.DeviceID == "" || !strings.HasPrefix(input.Tujuan, "62") || len(input.Tujuan) < 10 || input.Pesan == "" {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "WHATSAPP_VALIDATION_ERROR", "Perangkat, nomor tujuan format 62, dan isi pesan wajib diisi")
		return
	}
	if strings.TrimSpace(input.Idempotency) == "" {
		input.Idempotency = fmt.Sprintf("sirapi-%d", time.Now().UnixNano())
	}
	hasil, err := h.layanan.KirimPesanTeks(c.Request.Context(), whatsapp_gateway.PesanTeks{
		DeviceID: input.DeviceID, Tujuan: input.Tujuan, NamaPenerima: strings.TrimSpace(input.NamaPenerima),
		ExternalID: strings.TrimSpace(input.ExternalID), Isi: input.Pesan, Idempotency: strings.TrimSpace(input.Idempotency),
	})
	h.jawab(c, hasil, err, http.StatusAccepted)
}

func (h *Handler) jawab(c *gin.Context, hasil any, err error, status int) {
	if err == nil {
		httpresponse.Success(c, status, hasil)
		return
	}
	if errors.Is(err, whatsapp_gateway.ErrBelumDikonfigurasi) {
		httpresponse.Error(c, http.StatusServiceUnavailable, "WHATSAPP_NOT_CONFIGURED", err.Error())
		return
	}
	var gatewayError *whatsapp_gateway.ErrorGateway
	if errors.As(err, &gatewayError) {
		kode := gatewayError.Status
		if kode < 400 || kode > 599 {
			kode = http.StatusBadGateway
		}
		httpresponse.Error(c, kode, "WHATSAPP_GATEWAY_REJECTED", gatewayError.Pesan)
		return
	}
	httpresponse.Error(c, http.StatusBadGateway, "WHATSAPP_GATEWAY_UNAVAILABLE", err.Error())
}
