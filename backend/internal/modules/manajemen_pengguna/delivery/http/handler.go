package manajemen_penggunahttp

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/manajemen_pengguna"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct {
	repositori *manajemen_pengguna.Repositori
}

func NewHandler(repositori *manajemen_pengguna.Repositori) *Handler {
	return &Handler{repositori: repositori}
}

type tambahPenggunaRequest struct {
	Username       string   `json:"username" binding:"required,max=100"`
	Nama           string   `json:"nama" binding:"required,max=255"`
	Email          string   `json:"email" binding:"omitempty,email,max=255"`
	Aktif          *bool    `json:"aktif" binding:"required"`
	AllAccess      bool     `json:"all_access"`
	KodePermission []string `json:"permission"`
}

type ubahAksesPenggunaRequest struct {
	Aktif          *bool    `json:"aktif" binding:"required"`
	AllAccess      bool     `json:"all_access"`
	KodePermission []string `json:"permission"`
}

func (h *Handler) Daftar(c *gin.Context) {
	data, err := h.repositori.Daftar(c.Request.Context())
	if err != nil {
		httpresponse.Error(c, http.StatusServiceUnavailable, "USER_MANAGEMENT_UNAVAILABLE", "Data user management tidak dapat dibaca")
		return
	}

	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) CariPegawai(c *gin.Context) {
	kataKunci := strings.TrimSpace(c.Query("q"))
	if len(kataKunci) < 2 {
		httpresponse.Success(c, http.StatusOK, []manajemen_pengguna.Pegawai{})
		return
	}

	pegawai, err := h.repositori.CariPegawai(c.Request.Context(), kataKunci)
	if err != nil {
		httpresponse.Error(c, http.StatusServiceUnavailable, "SIMRS_EMPLOYEE_UNAVAILABLE", "Data pegawai SIMRS tidak dapat dibaca")
		return
	}

	httpresponse.Success(c, http.StatusOK, pegawai)
}

func (h *Handler) Tambah(c *gin.Context) {
	var request tambahPenggunaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Pegawai, status aktif, dan permission wajib diisi")
		return
	}

	kodePermission := permissionEfektif(request.AllAccess, request.KodePermission)
	if len(kodePermission) == 0 {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Minimal pilih satu permission")
		return
	}

	pengguna, err := h.repositori.Tambah(c.Request.Context(), manajemen_pengguna.InputTambahPengguna{
		Username:       strings.TrimSpace(request.Username),
		Nama:           strings.TrimSpace(request.Nama),
		Email:          strings.TrimSpace(request.Email),
		Aktif:          *request.Aktif,
		KodePermission: kodePermission,
	})
	if errors.Is(err, manajemen_pengguna.ErrUsernameAtauEmailSudahAda) {
		httpresponse.Error(c, http.StatusConflict, "USER_ALREADY_EXISTS", "Username atau email sudah digunakan")
		return
	}
	if errors.Is(err, manajemen_pengguna.ErrPermissionTidakValid) {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "INVALID_PERMISSION", "Permission yang dipilih tidak valid")
		return
	}
	if err != nil {
		httpresponse.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "User baru tidak dapat disimpan")
		return
	}

	httpresponse.Success(c, http.StatusCreated, gin.H{
		"message":  "User berhasil ditambahkan",
		"pengguna": pengguna,
	})
}

func (h *Handler) UbahAkses(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID == 0 {
		httpresponse.Error(c, http.StatusBadRequest, "INVALID_USER_ID", "ID user tidak valid")
		return
	}

	var request ubahAksesPenggunaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Status aktif dan permission wajib diisi")
		return
	}
	kodePermission := permissionEfektif(request.AllAccess, request.KodePermission)
	if len(kodePermission) == 0 {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Minimal pilih satu permission")
		return
	}

	err = h.repositori.UbahAkses(c.Request.Context(), userID, manajemen_pengguna.InputUbahAksesPengguna{
		Aktif:          *request.Aktif,
		KodePermission: kodePermission,
	})
	if errors.Is(err, manajemen_pengguna.ErrPenggunaTidakDitemukan) {
		httpresponse.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "User tidak ditemukan")
		return
	}
	if errors.Is(err, manajemen_pengguna.ErrPermissionTidakValid) {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "INVALID_PERMISSION", "Permission yang dipilih tidak valid")
		return
	}
	if err != nil {
		httpresponse.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Akses user tidak dapat disimpan")
		return
	}

	httpresponse.Success(c, http.StatusOK, gin.H{"message": "Akses user berhasil diperbarui"})
}

func permissionEfektif(allAccess bool, kodePermission []string) []string {
	if allAccess {
		return []string{"*"}
	}
	return kodePermission
}
