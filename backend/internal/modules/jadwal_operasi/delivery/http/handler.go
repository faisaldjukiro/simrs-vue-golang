package jadwaloperasihttp

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/modules/jadwal_operasi"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct {
	repo *jadwal_operasi.Repositori
}

func NewHandler(repo *jadwal_operasi.Repositori) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Register(group *gin.RouterGroup) {
	group.GET("", h.daftar)
	group.GET("/referensi", h.referensi)
	group.GET("/pendukung", h.pendukung)
	group.POST("/pendukung", h.simpanPendukung)
	group.PUT("/pendukung", h.simpanPendukung)
	group.DELETE("/pendukung", h.simpanPendukung)
	group.POST("", h.simpan)
	group.PUT("/:id", h.simpan)
	group.DELETE("/:id", h.hapus)
}

func (h *Handler) simpanPendukung(c *gin.Context) {
	user, admin, ok := pengguna(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256*1024)
	var input jadwal_operasi.InputPendukung
	if err := c.ShouldBindJSON(&input); err != nil {
		tulisError(c, jadwal_operasi.ErrValidasi)
		return
	}
	if err := h.repo.SimpanPendukung(c.Request.Context(), input, c.Request.Method, user.Username, admin); err != nil {
		tulisError(c, err)
		return
	}
	pesan := "Catatan berhasil disimpan langsung ke Khanza"
	if c.Request.Method == http.MethodDelete {
		pesan = "Catatan berhasil dihapus dari Khanza"
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": pesan})
}

func (h *Handler) pendukung(c *gin.Context) {
	user, admin, ok := pengguna(c)
	if !ok {
		return
	}
	data, err := h.repo.Pendukung(c.Request.Context(), c.Query("jenis"), c.Query("no_rawat"), c.Query("tanggal"))
	if err != nil {
		if errors.Is(err, jadwal_operasi.ErrValidasi) {
			tulisError(c, err)
		} else {
			httpresponse.Error(c, http.StatusInternalServerError, "RIWAYAT_OPERASI_ERROR", "Riwayat belum dapat dibaca. Periksa koneksi dan ketersediaan tabel modul ini di Khanza.")
		}
		return
	}
	for _, row := range data.Catatan {
		row["_bisa_ubah"] = strconv.FormatBool(jadwal_operasi.BolehUbahPendukung(c.Query("jenis"), user.Username, admin, row))
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func pengguna(c *gin.Context) (autentikasi.Pengguna, bool, bool) {
	value, exists := c.Get("authenticated_user")
	user, ok := value.(autentikasi.Pengguna)
	if !exists || !ok || user.ID == 0 {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Sesi login tidak valid")
		return user, false, false
	}
	admin := false
	for _, permission := range user.Permissions {
		if permission == "*" {
			admin = true
		}
	}
	return user, admin, true
}

func tulisError(c *gin.Context, err error) {
	if errors.Is(err, jadwal_operasi.ErrAksesPendukung) {
		httpresponse.Error(c, http.StatusForbidden, "PENDUKUNG_OPERASI_FORBIDDEN", err.Error())
	} else if errors.Is(err, jadwal_operasi.ErrValidasi) {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "JADWAL_OPERASI_VALIDATION_ERROR", err.Error())
	} else if errors.Is(err, jadwal_operasi.ErrKonflik) || errors.Is(err, jadwal_operasi.ErrBentrok) || errors.Is(err, jadwal_operasi.ErrTerkunci) {
		httpresponse.Error(c, http.StatusConflict, "JADWAL_OPERASI_CONFLICT", err.Error())
	} else {
		httpresponse.Error(c, http.StatusInternalServerError, "JADWAL_OPERASI_ERROR", "Jadwal belum dapat diproses. Periksa koneksi/izin database Khanza dan migration SIRAPI. Muat ulang riwayat sebelum mencoba simpan kembali.")
	}
}

func (h *Handler) daftar(c *gin.Context) {
	user, admin, ok := pengguna(c)
	if !ok {
		return
	}
	data, err := h.repo.Daftar(c.Request.Context(), c.Query("no_rawat"), user.ID, admin)
	if err != nil {
		tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) referensi(c *gin.Context) {
	if _, _, ok := pengguna(c); !ok {
		return
	}
	data, err := h.repo.Referensi(c.Request.Context(), c.Query("jenis"), c.Query("q"), c.Query("no_rawat"))
	if err != nil {
		tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func bacaInput(c *gin.Context) (jadwal_operasi.Input, bool) {
	var input jadwal_operasi.Input
	if err := c.ShouldBindJSON(&input); err != nil {
		tulisError(c, jadwal_operasi.ErrValidasi)
		return input, false
	}
	// ID hanya berasal dari URL, tidak boleh disisipkan ke endpoint pembuatan.
	input.ID = 0
	input.Sumber = "SIRAPI"
	if c.Request.Method != http.MethodPost && c.Param("id") == "khanza" {
		input.Sumber = "Khanza"
		if input.Asli == nil {
			tulisError(c, jadwal_operasi.ErrValidasi)
			return input, false
		}
		return input, true
	}
	if c.Request.Method != http.MethodPost {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			tulisError(c, jadwal_operasi.ErrValidasi)
			return input, false
		}
		input.ID = id
	}
	return input, true
}

func (h *Handler) simpan(c *gin.Context) {
	user, admin, ok := pengguna(c)
	if !ok {
		return
	}
	input, ok := bacaInput(c)
	if !ok {
		return
	}
	if err := h.repo.Simpan(c.Request.Context(), input, user.ID, admin); err != nil {
		tulisError(c, err)
		return
	}
	status := http.StatusOK
	if c.Request.Method == http.MethodPost {
		status = http.StatusCreated
	}
	pesan := "Perubahan jadwal lokal disimpan di SIRAPI"
	if input.ID == 0 {
		pesan = "Jadwal operasi berhasil disimpan langsung di Khanza"
	}
	httpresponse.Success(c, status, gin.H{"pesan": pesan, "waktu": time.Now().UTC()})
}

func (h *Handler) hapus(c *gin.Context) {
	user, admin, ok := pengguna(c)
	if !ok {
		return
	}
	input, ok := bacaInput(c)
	if !ok {
		return
	}
	if err := h.repo.Hapus(c.Request.Context(), input, user.ID, admin); err != nil {
		tulisError(c, err)
		return
	}
	pesan := "Jadwal dihapus dari daftar aktif SIRAPI"
	if input.Sumber == "Khanza" {
		pesan = "Jadwal berhasil dihapus dari Khanza"
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": pesan})
}
