package checklistpreoperasihttp

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/modules/checklist_pre_operasi"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct {
	repo *checklist_pre_operasi.Repositori
}

func NewHandler(repo *checklist_pre_operasi.Repositori) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Register(group *gin.RouterGroup) {
	group.GET("", h.daftar)
	group.GET("/referensi", h.referensi)
	group.POST("", h.simpan)
	group.PUT("/:id", h.simpan)
	group.DELETE("/:id", h.hapus)
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
	if errors.Is(err, checklist_pre_operasi.ErrValidasi) {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "CHECKLIST_VALIDATION_ERROR", err.Error())
	} else if errors.Is(err, checklist_pre_operasi.ErrKonflik) {
		httpresponse.Error(c, http.StatusConflict, "CHECKLIST_CONFLICT", err.Error())
	} else {
		httpresponse.Error(c, http.StatusInternalServerError, "CHECKLIST_ERROR", "Checklist pre operasi belum dapat diproses. Periksa koneksi dan migration database SIRAPI.")
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
	data, err := h.repo.Referensi(c.Request.Context(), c.Query("jenis"), c.Query("q"))
	if err != nil {
		tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func bacaInput(c *gin.Context) (checklist_pre_operasi.Input, bool) {
	var input checklist_pre_operasi.Input
	if err := c.ShouldBindJSON(&input); err != nil {
		tulisError(c, checklist_pre_operasi.ErrValidasi)
		return input, false
	}
	// ID hanya berasal dari URL, tidak boleh disisipkan ke endpoint pembuatan.
	input.ID = 0
	if c.Request.Method != http.MethodPost {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			tulisError(c, checklist_pre_operasi.ErrValidasi)
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
	httpresponse.Success(c, status, gin.H{"pesan": "Checklist disimpan di SIRAPI", "waktu": time.Now().UTC()})
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
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Checklist dihapus dari daftar aktif SIRAPI"})
}
