package kelolamenuhttp

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/kelola_menu"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct{ repositori *kelola_menu.Repositori }

func NewHandler(repositori *kelola_menu.Repositori) *Handler { return &Handler{repositori: repositori} }

func (h *Handler) Register(group *gin.RouterGroup) {
	group.GET("", h.daftar)
	group.POST("/sidebar", h.tambahSidebar)
	group.PUT("/sidebar/:id", h.ubahSidebar)
	group.DELETE("/sidebar/:id", h.hapusSidebar)
}

type sidebarRequest struct {
	Kode        string   `json:"kode" binding:"required,max=80"`
	Nama        string   `json:"nama" binding:"required,max=80"`
	Ikon        string   `json:"ikon" binding:"required,max=40"`
	DaftarModul []string `json:"daftar_modul" binding:"required,min=1"`
	Urutan      uint16   `json:"urutan"`
	Aktif       *bool    `json:"aktif" binding:"required"`
}

func (h *Handler) daftar(c *gin.Context) {
	data, err := h.repositori.Daftar(c.Request.Context())
	if err != nil {
		httpresponse.Error(c, http.StatusServiceUnavailable, "KELOLA_MENU_UNAVAILABLE", "Konfigurasi sidebar tidak dapat dibaca")
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) tambahSidebar(c *gin.Context) {
	input, valid := bacaInput(c)
	if !valid {
		return
	}
	sidebar, err := h.repositori.Tambah(c.Request.Context(), input)
	if tanganiError(c, err) {
		return
	}
	httpresponse.Success(c, http.StatusCreated, gin.H{"message": "Sidebar berhasil ditambahkan", "sidebar": sidebar})
}

func (h *Handler) ubahSidebar(c *gin.Context) {
	id, valid := bacaID(c)
	if !valid {
		return
	}
	input, valid := bacaInput(c)
	if !valid {
		return
	}
	if tanganiError(c, h.repositori.Ubah(c.Request.Context(), id, input)) {
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"message": "Sidebar berhasil diperbarui"})
}

func (h *Handler) hapusSidebar(c *gin.Context) {
	id, valid := bacaID(c)
	if !valid {
		return
	}
	if tanganiError(c, h.repositori.Hapus(c.Request.Context(), id)) {
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"message": "Sidebar berhasil dihapus"})
}

func bacaInput(c *gin.Context) (kelola_menu.InputSidebar, bool) {
	var request sidebarRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Kode, nama, ikon, modul, urutan, dan status sidebar wajib diisi dengan benar")
		return kelola_menu.InputSidebar{}, false
	}
	input := kelola_menu.InputSidebar{
		Kode: strings.TrimSpace(request.Kode), Nama: strings.TrimSpace(request.Nama), Ikon: strings.TrimSpace(request.Ikon),
		DaftarModul: request.DaftarModul, Urutan: request.Urutan, Aktif: *request.Aktif,
	}
	if !kelola_menu.ValidasiInput(input) {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Kode hanya boleh memakai huruf kecil, angka, dan garis bawah; pilih minimal satu modul")
		return kelola_menu.InputSidebar{}, false
	}
	return input, true
}

func bacaID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		httpresponse.Error(c, http.StatusBadRequest, "INVALID_SIDEBAR_ID", "ID sidebar tidak valid")
		return 0, false
	}
	return id, true
}

func tanganiError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, kelola_menu.ErrSidebarTidakDitemukan) {
		httpresponse.Error(c, http.StatusNotFound, "SIDEBAR_NOT_FOUND", "Sidebar tidak ditemukan")
		return true
	}
	if errors.Is(err, kelola_menu.ErrKodeSidebarDigunakan) {
		httpresponse.Error(c, http.StatusConflict, "SIDEBAR_CODE_EXISTS", "Kode sidebar sudah digunakan")
		return true
	}
	httpresponse.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Sidebar tidak dapat diproses")
	return true
}
