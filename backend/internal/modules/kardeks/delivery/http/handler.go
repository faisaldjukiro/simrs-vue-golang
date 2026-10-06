package kardekshttp

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/modules/kardeks"
	"simrs-backend/internal/shared/httpresponse"
	"time"
)

type Handler struct{ repo *kardeks.Repositori }

func NewHandler(r *kardeks.Repositori) *Handler { return &Handler{repo: r} }
func (h *Handler) Register(g *gin.RouterGroup) {
	g.GET("", h.daftar)
	g.POST("/cairan", h.simpanCairan)
}

func (h *Handler) simpanCairan(c *gin.Context) {
	v, _ := c.Get("authenticated_user")
	u, ok := v.(autentikasi.Pengguna)
	if !ok || u.ID == 0 {
		httpresponse.Error(c, 401, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32*1024)
	var in kardeks.InputCairan
	if c.ShouldBindJSON(&in) != nil {
		httpresponse.Error(c, 422, "VALIDATION_ERROR", "Data cairan tidak valid")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()
	if err := h.repo.SimpanCairan(ctx, in, u.Username); err != nil {
		if errors.Is(err, kardeks.ErrValidasi) {
			httpresponse.Error(c, 422, "VALIDATION_ERROR", err.Error())
		} else {
			httpresponse.Error(c, 500, "CAIRAN_ERROR", "Catatan cairan gagal diproses. Muat ulang riwayat sebelum mencoba lagi; periksa koneksi dan izin SIMRS.")
		}
		return
	}
	httpresponse.Success(c, 201, gin.H{"pesan": "Catatan cairan berhasil disimpan ke SIMRS"})
}
func (h *Handler) daftar(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 25*time.Second)
	defer cancel()
	hsl, err := h.repo.Daftar(ctx, c.Query("no_rawat"), c.Query("tanggal"), c.DefaultQuery("jam", "08:00"))
	if err != nil {
		if errors.Is(err, kardeks.ErrValidasi) {
			httpresponse.Error(c, 422, "VALIDATION_ERROR", err.Error())
		} else {
			httpresponse.Error(c, 500, "KARDEKS_ERROR", "Kardeks belum dapat dimuat")
		}
		return
	}
	httpresponse.Success(c, 200, hsl)
}
