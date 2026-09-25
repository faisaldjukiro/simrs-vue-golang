package risikojatuhanakhttp

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/modules/risiko_jatuh_anak"
	"simrs-backend/internal/shared/httpresponse"
	"time"
)

type Handler struct {
	repo *risiko_jatuh_anak.Repositori
}

func NewHandler(repo *risiko_jatuh_anak.Repositori) *Handler { return &Handler{repo: repo} }
func (h *Handler) Register(g *gin.RouterGroup) {
	g.GET("", h.proses)
	g.POST("", h.proses)
	g.PUT("", h.proses)
	g.DELETE("", h.proses)
}
func (h *Handler) proses(c *gin.Context) {
	v, ada := c.Get("authenticated_user")
	u, ok := v.(autentikasi.Pengguna)
	if !ada || !ok || u.ID == 0 {
		httpresponse.Error(c, 401, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	admin := false
	for _, p := range u.Permissions {
		if p == "*" {
			admin = true
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()
	if c.Request.Method == http.MethodGet {
		hasil, err := h.repo.Daftar(ctx, c.Query("no_rawat"), u.Username, admin)
		if err != nil {
			tulisError(c, err)
			return
		}
		httpresponse.Success(c, 200, hasil)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32*1024)
	var in risiko_jatuh_anak.Input
	if c.ShouldBindJSON(&in) != nil {
		tulisError(c, risiko_jatuh_anak.ErrValidasi)
		return
	}
	if err := h.repo.Mutasi(ctx, in, c.Request.Method, u.Username, admin); err != nil {
		tulisError(c, err)
		return
	}
	pesan := "Penilaian risiko jatuh berhasil disimpan ke Khanza"
	if c.Request.Method == http.MethodDelete {
		pesan = "Penilaian risiko jatuh berhasil dihapus dari Khanza"
	}
	httpresponse.Success(c, 200, gin.H{"pesan": pesan})
}
func tulisError(c *gin.Context, err error) {
	status, kode, pesan := 500, "RISIKO_JATUH_ERROR", "Penilaian belum dapat diproses. Periksa koneksi dan tabel Khanza, lalu muat ulang riwayat."
	switch {
	case errors.Is(err, risiko_jatuh_anak.ErrValidasi):
		status, kode, pesan = 422, "VALIDATION_ERROR", err.Error()
	case errors.Is(err, risiko_jatuh_anak.ErrAkses):
		status, kode, pesan = 403, "FORBIDDEN", err.Error()
	case errors.Is(err, risiko_jatuh_anak.ErrKonflik):
		status, kode, pesan = 409, "CONFLICT", err.Error()
	}
	httpresponse.Error(c, status, kode, pesan)
}
