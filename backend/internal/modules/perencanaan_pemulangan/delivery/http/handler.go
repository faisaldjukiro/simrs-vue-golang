package pemulanganhttp

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"net/http"
	"simrs-backend/internal/modules/autentikasi"
	modul "simrs-backend/internal/modules/perencanaan_pemulangan"
	"simrs-backend/internal/shared/httpresponse"
	"simrs-backend/internal/shared/khanzamutasi"
	"time"
)

type Handler struct{ repo *modul.Repositori }

func NewHandler(r *modul.Repositori) *Handler { return &Handler{repo: r} }
func (h *Handler) Register(g *gin.RouterGroup) {
	g.GET("", h.proses)
	g.GET("/referensi", h.proses)
	g.GET("/kop", h.proses)
	g.POST("", h.proses)
	g.PUT("", h.proses)
	g.DELETE("", h.proses)
}
func (h *Handler) proses(c *gin.Context) {
	v, _ := c.Get("authenticated_user")
	u, ok := v.(autentikasi.Pengguna)
	if !ok || u.ID == 0 {
		httpresponse.Error(c, 401, "UNAUTHENTICATED", "Sesi tidak valid")
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
	if c.Request.Method == "GET" {
		var data any
		var err error
		if c.FullPath() == "/api/perencanaan-pemulangan/kop" {
			data, err = h.repo.Kop(ctx)
		} else if c.FullPath() == "/api/perencanaan-pemulangan/referensi" {
			data, err = h.repo.Referensi(ctx, c.Query("q"))
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
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32768)
	var in modul.Input
	if c.ShouldBindJSON(&in) != nil {
		httpresponse.Error(c, 422, "VALIDATION_ERROR", "Format form tidak sesuai; muat ulang halaman")
		return
	}
	if err := h.repo.Mutasi(ctx, in, c.Request.Method, u.Username, admin); err != nil {
		gagal(c, err)
		return
	}
	pesan := "Perencanaan pemulangan berhasil disimpan di SIMRS"
	if c.Request.Method == "DELETE" {
		pesan = "Perencanaan pemulangan berhasil dihapus"
	}
	httpresponse.Success(c, 200, gin.H{"pesan": pesan})
}
func gagal(c *gin.Context, err error) {
	var dbErr *mysql.MySQLError
	status, kode, pesan := 500, "PEMULANGAN_ERROR", "Perencanaan pemulangan belum dapat diproses. Periksa koneksi, izin CRUD, dan struktur tabel SIMRS."
	switch {
	case errors.Is(err, modul.ErrValidasi):
		status, kode, pesan = 422, "VALIDATION_ERROR", err.Error()
	case errors.Is(err, modul.ErrAkses):
		status, kode, pesan = 403, "FORBIDDEN", err.Error()
	case errors.Is(err, khanzamutasi.ErrKonflik):
		status, kode, pesan = 409, "CONFLICT", "Catatan berubah atau hilang. Muat ulang sebelum mengedit."
	case errors.As(err, &dbErr) && dbErr.Number == 1062:
		status, kode, pesan = 409, "CONFLICT", "Perencanaan kunjungan ini sudah ada. Gunakan Edit pada riwayat."
	}
	httpresponse.Error(c, status, kode, pesan)
}
