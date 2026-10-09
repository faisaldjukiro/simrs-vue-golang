package ringkasanpasienhttp

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/modules/ringkasan_pasien"
	"simrs-backend/internal/shared/httpresponse"
	"time"
)

type Handler struct{ repo *ringkasan_pasien.Repositori }

func NewHandler(repo *ringkasan_pasien.Repositori) *Handler { return &Handler{repo: repo} }
func (h *Handler) Register(g *gin.RouterGroup)              { g.GET("", h.daftar) }
func (h *Handler) daftar(c *gin.Context) {
	v, _ := c.Get("authenticated_user")
	u, ok := v.(autentikasi.Pengguna)
	if !ok || u.ID == 0 {
		httpresponse.Error(c, 401, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	modul := c.Query("modul")
	_, _, izin := ringkasan_pasien.Konteks(modul)
	if izin == nil {
		httpresponse.Error(c, 422, "VALIDATION_ERROR", "Modul tidak valid")
		return
	}
	boleh := false
	for _, p := range u.Permissions {
		if p == "*" {
			boleh = true
		}
		for _, i := range izin {
			if p == i {
				boleh = true
			}
		}
	}
	if !boleh {
		httpresponse.Error(c, 403, "FORBIDDEN", "Tidak memiliki akses ke layanan ini")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()
	hasil, err := h.repo.Daftar(ctx, c.Query("no_rawat"), modul)
	if err != nil {
		switch {
		case errors.Is(err, ringkasan_pasien.ErrValidasi):
			httpresponse.Error(c, 422, "VALIDATION_ERROR", err.Error())
		case errors.Is(err, ringkasan_pasien.ErrTidakAda):
			httpresponse.Error(c, 404, "NOT_FOUND", err.Error())
		default:
			httpresponse.Error(c, 503, "RINGKASAN_ERROR", "Ringkasan belum dapat dibaca dari SIMRS. Silakan muat ulang.")
		}
		return
	}
	httpresponse.Success(c, 200, hasil)
}
