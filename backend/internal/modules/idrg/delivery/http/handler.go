package idrghttp

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/modules/idrg"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct {
	repositori *idrg.Repositori
	layanan    *idrg.Layanan
}

func NewHandler(repositori *idrg.Repositori, layanan *idrg.Layanan) *Handler {
	return &Handler{repositori: repositori, layanan: layanan}
}

func (h *Handler) Register(group *gin.RouterGroup) {
	group.GET("", h.DaftarPasien)
	group.GET("/diagnosa", h.AmbilDiagnosa)
	group.GET("/prosedur", h.AmbilProsedur)
	group.GET("/cari-diagnosa", h.CariDiagnosa)
	group.GET("/cari-prosedur", h.CariProsedur)
	group.POST("/proses", h.Proses)
}

func (h *Handler) DaftarPasien(c *gin.Context) {
	tanggal := tanggalQuery(c, "tanggal", time.Now().Format("2006-01-02"))
	filter := idrg.FilterPasien{
		Tanggal:    tanggal,
		JenisRawat: c.DefaultQuery("jenis_rawat", "ranap"),
		Pencarian:  c.Query("cari"),
		Halaman:    angkaQuery(c, "page", 1),
		Batas:      angkaQuery(c, "limit", 50),
	}

	data, err := h.repositori.DaftarPasien(c.Request.Context(), filter)
	if err != nil {
		httpresponse.Error(c, http.StatusServiceUnavailable, "IDRG_SIMRS_UNAVAILABLE", "Data pasien IDRG tidak dapat dibaca dari SIMRS lama")
		return
	}

	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) AmbilDiagnosa(c *gin.Context) {
	noRawat := strings.TrimSpace(c.Query("no_rawat"))
	if noRawat == "" {
		httpresponse.Error(c, http.StatusBadRequest, "NO_RAWAT_REQUIRED", "Nomor rawat wajib diisi")
		return
	}

	data, err := h.repositori.AmbilDiagnosa(c.Request.Context(), noRawat)
	if err != nil {
		httpresponse.Error(c, http.StatusServiceUnavailable, "IDRG_DIAGNOSA_UNAVAILABLE", "Diagnosa pasien tidak dapat dibaca")
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) AmbilProsedur(c *gin.Context) {
	noRawat := strings.TrimSpace(c.Query("no_rawat"))
	if noRawat == "" {
		httpresponse.Error(c, http.StatusBadRequest, "NO_RAWAT_REQUIRED", "Nomor rawat wajib diisi")
		return
	}

	data, err := h.repositori.AmbilProsedur(c.Request.Context(), noRawat)
	if err != nil {
		httpresponse.Error(c, http.StatusServiceUnavailable, "IDRG_PROSEDUR_UNAVAILABLE", "Prosedur pasien tidak dapat dibaca")
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) CariDiagnosa(c *gin.Context) {
	data, err := h.repositori.CariDiagnosa(c.Request.Context(), c.Query("q"))
	if err != nil {
		httpresponse.Error(c, http.StatusServiceUnavailable, "IDRG_DIAGNOSA_SEARCH_UNAVAILABLE", "Master diagnosa tidak dapat dibaca")
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) CariProsedur(c *gin.Context) {
	data, err := h.repositori.CariProsedur(c.Request.Context(), c.Query("q"))
	if err != nil {
		httpresponse.Error(c, http.StatusServiceUnavailable, "IDRG_PROSEDUR_SEARCH_UNAVAILABLE", "Master prosedur tidak dapat dibaca")
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

type prosesRequest struct {
	Aksi    string         `json:"aksi"`
	NoRawat string         `json:"no_rawat"`
	NoSEP   string         `json:"no_sep"`
	Pasien  map[string]any `json:"pasien"`
	Klaim   map[string]any `json:"klaim"`
}

func (h *Handler) Proses(c *gin.Context) {
	var request prosesRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Aksi dan data pasien wajib diisi")
		return
	}

	aksi := strings.TrimSpace(request.Aksi)
	if aksi == "" || strings.TrimSpace(request.NoRawat) == "" {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Aksi dan nomor rawat wajib diisi")
		return
	}

	data, err := h.layanan.Proses(c.Request.Context(), idrg.InputProses{
		Aksi:     request.Aksi,
		NoRawat:  request.NoRawat,
		NoSEP:    request.NoSEP,
		Username: usernameLogin(c),
		Pasien:   request.Pasien,
		Klaim:    request.Klaim,
	})
	if err != nil {
		switch {
		case errors.Is(err, idrg.ErrEKlaimBelumAktif):
			httpresponse.Error(c, http.StatusUnprocessableEntity, "EKLAIM_NOT_CONFIGURED", "Konfigurasi E-Klaim belum lengkap. Isi INACBG_URL_WS dan INACBG_KEY di .env backend.")
		case errors.Is(err, idrg.ErrInputTidakValid):
			httpresponse.Error(c, http.StatusUnprocessableEntity, "IDRG_VALIDATION_ERROR", err.Error())
		case errors.Is(err, idrg.ErrAksiBelumSiap):
			httpresponse.Error(c, http.StatusUnprocessableEntity, "IDRG_ACTION_NOT_READY", err.Error())
		default:
			httpresponse.Error(c, http.StatusBadGateway, "EKLAIM_PROCESS_FAILED", err.Error())
		}
		return
	}

	httpresponse.Success(c, http.StatusOK, data)
}

func usernameLogin(c *gin.Context) string {
	value, exists := c.Get("authenticated_user")
	if !exists {
		return ""
	}
	user, ok := value.(autentikasi.Pengguna)
	if !ok {
		return ""
	}
	return strings.TrimSpace(user.Username)
}

func tanggalQuery(c *gin.Context, nama string, fallback string) string {
	nilai := strings.TrimSpace(c.Query(nama))
	if nilai == "" {
		return fallback
	}
	if _, err := time.Parse("2006-01-02", nilai); err != nil {
		return fallback
	}
	return nilai
}

func angkaQuery(c *gin.Context, nama string, fallback int) int {
	nilai := strings.TrimSpace(c.Query(nama))
	if nilai == "" {
		return fallback
	}
	angka, err := strconv.Atoi(nilai)
	if err != nil {
		return fallback
	}
	return angka
}
