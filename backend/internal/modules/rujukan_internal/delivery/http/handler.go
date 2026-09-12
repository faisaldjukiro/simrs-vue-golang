package rujukaninternalhttp

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/modules/rujukan_internal"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct {
	repositori *rujukan_internal.Repositori
	jenis      string
}

func NewHandler(repositori *rujukan_internal.Repositori, jenis string) *Handler {
	return &Handler{repositori: repositori, jenis: jenis}
}

func (h *Handler) Register(group *gin.RouterGroup) {
	group.GET("", h.Daftar)
	group.GET("/referensi/:jenis", h.CariReferensi)
	group.POST("", h.Simpan)
	group.PUT("", func(c *gin.Context) { h.mutasi(c, "ubah") })
	group.DELETE("", func(c *gin.Context) { h.mutasi(c, "hapus") })
	group.POST("/kirim", func(c *gin.Context) { h.mutasi(c, "kirim") })
}

func (h *Handler) Daftar(c *gin.Context) {
	data, err := h.repositori.Daftar(c.Request.Context(), h.jenis, c.Query("no_rawat"))
	if err != nil {
		tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) CariReferensi(c *gin.Context) {
	data, err := h.repositori.CariReferensi(c.Request.Context(), c.Param("jenis"), c.Query("q"))
	if err != nil {
		tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) Simpan(c *gin.Context) {
	value, exists := c.Get("authenticated_user")
	user, ok := value.(autentikasi.Pengguna)
	if !exists || !ok || strings.TrimSpace(user.Username) == "" {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	var input rujukan_internal.Input
	if err := c.ShouldBindJSON(&input); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "RUJUKAN_INPUT_INVALID", "Data rujukan tidak lengkap")
		return
	}
	if err := h.repositori.Simpan(c.Request.Context(), h.jenis, input); err != nil {
		tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusCreated, gin.H{"pesan": "Rujukan berhasil disimpan langsung di Khanza."})
}

func (h *Handler) mutasi(c *gin.Context, aksi string) {
	value, exists := c.Get("authenticated_user")
	user, ok := value.(autentikasi.Pengguna)
	if !exists || !ok || strings.TrimSpace(user.Username) == "" {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	var input rujukan_internal.Mutasi
	if err := c.ShouldBindJSON(&input); err != nil {
		tulisError(c, rujukan_internal.ErrInput)
		return
	}
	var err error
	pesan := "Rujukan berhasil diperbarui di " + input.Asal.Sumber + "."
	switch aksi {
	case "kirim":
		pesan, err = h.repositori.Kirim(c.Request.Context(), h.jenis, input.Asal)
	case "hapus":
		err = h.repositori.Mutasi(c.Request.Context(), h.jenis, input.Asal, nil)
		pesan = "Rujukan berhasil dihapus dari " + input.Asal.Sumber + "."
	case "ubah":
		err = h.repositori.Mutasi(c.Request.Context(), h.jenis, input.Asal, &input.Baru)
	}
	if err != nil {
		tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{
		"pesan":      pesan,
		"peringatan": pesan == rujukan_internal.PesanArsipGagal,
	})
}

func tulisError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, rujukan_internal.ErrInput):
		httpresponse.Error(c, http.StatusUnprocessableEntity, "RUJUKAN_INPUT_INVALID", err.Error())
	case errors.Is(err, rujukan_internal.ErrTidakAda):
		httpresponse.Error(c, http.StatusNotFound, "RUJUKAN_KUNJUNGAN_NOT_FOUND", err.Error())
	case errors.Is(err, rujukan_internal.ErrDuplikat), errors.Is(err, rujukan_internal.ErrKunjungan), errors.Is(err, rujukan_internal.ErrBerubah), errors.Is(err, rujukan_internal.ErrBilling):
		httpresponse.Error(c, http.StatusConflict, "RUJUKAN_CONFLICT", err.Error())
	default:
		httpresponse.Error(c, http.StatusServiceUnavailable, "RUJUKAN_UNAVAILABLE", "Rujukan tidak dapat diproses. Periksa koneksi database dan migration lokal.")
	}
}
