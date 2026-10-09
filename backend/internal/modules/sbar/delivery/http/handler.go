package sbarhttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/modules/sbar"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct{ repo *sbar.Repositori }

func NewHandler(repo *sbar.Repositori) *Handler { return &Handler{repo: repo} }
func (h *Handler) Register(g *gin.RouterGroup) {
	g.GET("", h.proses)
	g.GET("/referensi", h.proses)
	g.POST("", h.proses)
	g.PUT("", h.proses)
	g.DELETE("", h.proses)
	g.POST("/verifikasi", h.proses)
}

func (h *Handler) proses(c *gin.Context) {
	v, _ := c.Get("authenticated_user")
	u, ok := v.(autentikasi.Pengguna)
	if !ok || u.ID == 0 {
		httpresponse.Error(c, 401, "UNAUTHENTICATED", "Sesi tidak valid")
		return
	}
	admin, akses := false, false
	for _, p := range u.Permissions {
		if p == "*" {
			admin = true
			akses = true
		}
		if p == "kamar_inap" || p == "daftar_pasien_ranap" {
			akses = true
		}
	}
	if !akses {
		httpresponse.Error(c, 403, "FORBIDDEN", "Akses rawat inap diperlukan")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()
	if c.Request.Method == http.MethodGet {
		var data any
		var err error
		if c.FullPath() == "/api/sbar/referensi" {
			data, err = h.repo.Referensi(ctx, c.Query("jenis"), c.Query("q"))
		} else {
			data, err = h.repo.Daftar(ctx, c.Query("no_rawat"), u.Username, admin)
		}
		if err != nil {
			gagal(c, err)
			return
		}
		httpresponse.Success(c, 200, data)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 65536)
	var in sbar.Input
	if c.ShouldBindJSON(&in) != nil {
		httpresponse.Error(c, 422, "VALIDATION_ERROR", "Format SBAR tidak sesuai; muat ulang halaman")
		return
	}
	aksi := c.Request.Method
	if c.FullPath() == "/api/sbar/verifikasi" {
		aksi = "verifikasi"
	}
	if err := h.repo.Mutasi(ctx, in, aksi, u.Username, admin); err != nil {
		gagal(c, err)
		return
	}
	pesan := "SBAR berhasil disimpan di SIMRS"
	if aksi == "DELETE" {
		pesan = "SBAR berhasil dihapus"
	}
	if aksi == "verifikasi" {
		pesan = "SBAR berhasil diverifikasi"
	}
	httpresponse.Success(c, 200, gin.H{"pesan": pesan})
}

func gagal(c *gin.Context, err error) {
	var dbErr *mysql.MySQLError
	status, kode, pesan := 503, "SBAR_ERROR", "SBAR belum dapat diproses. Periksa koneksi, izin CRUD, dan tabel SBAR SIMRS."
	switch {
	case errors.Is(err, sbar.ErrValidasi):
		status, kode, pesan = 422, "VALIDATION_ERROR", err.Error()
	case errors.Is(err, sbar.ErrAkses):
		status, kode, pesan = 403, "FORBIDDEN", err.Error()
	case errors.Is(err, sbar.ErrKonflik):
		status, kode, pesan = 409, "CONFLICT", err.Error()
	case errors.As(err, &dbErr) && dbErr.Number == 1062:
		status, kode, pesan = 409, "CONFLICT", "SBAR atau verifikasi pada waktu ini sudah ada. Muat ulang riwayat."
	case errors.As(err, &dbErr) && (dbErr.Number == 1406 || dbErr.Number == 1366):
		status, kode, pesan = 422, "VALIDATION_ERROR", "Isi SBAR melampaui panjang atau karakter yang didukung kolom SIMRS."
	}
	httpresponse.Error(c, status, kode, pesan)
}
