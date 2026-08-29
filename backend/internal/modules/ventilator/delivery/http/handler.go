package ventilatorhttp

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/modules/ventilator"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct{ repo *ventilator.Repositori }

func NewHandler(repo *ventilator.Repositori) *Handler { return &Handler{repo: repo} }
func (h *Handler) RegisterMaster(g *gin.RouterGroup) {
	g.GET("", h.master)
	g.POST("", h.simpanMaster)
	g.PUT("/:kode", h.ubahMaster)
	g.DELETE("/:kode", h.hapusMaster)
}
func (h *Handler) RegisterPasien(g *gin.RouterGroup) {
	g.GET("", h.pasien)
	g.POST("/pemakaian", h.simpanPemakaian)
	g.POST("/setting", h.simpanSetting)
	g.POST("/monitoring", h.simpanMonitoring)
	g.POST("/checklist-vap", h.simpanVAP)
}
func (h *Handler) master(c *gin.Context) { d, e := h.repo.DaftarMaster(c); h.hasil(c, d, e) }
func validMaster(x ventilator.Master) bool {
	return strings.TrimSpace(x.Nama) != "" && strings.TrimSpace(x.Status) != ""
}
func (h *Handler) simpanMaster(c *gin.Context) {
	var x ventilator.Master
	if c.ShouldBindJSON(&x) != nil || !validMaster(x) {
		httpresponse.Error(c, 422, "VENTILATOR_VALIDATION_ERROR", "Nama dan status ventilator wajib diisi")
		return
	}
	kode, e := h.repo.SimpanMaster(c, x)
	if e != nil {
		h.err(c, e)
		return
	}
	httpresponse.Success(c, http.StatusCreated, gin.H{"pesan": "Master ventilator berhasil disimpan", "kode_ventilator": kode})
}
func (h *Handler) ubahMaster(c *gin.Context) {
	var x ventilator.Master
	if c.ShouldBindJSON(&x) != nil || strings.TrimSpace(x.Nama) == "" {
		httpresponse.Error(c, 422, "VENTILATOR_VALIDATION_ERROR", "Data master ventilator tidak lengkap")
		return
	}
	e := h.repo.UbahMaster(c, c.Param("kode"), x)
	h.pesan(c, 200, "Master ventilator berhasil diperbarui", e)
}
func (h *Handler) hapusMaster(c *gin.Context) {
	h.pesan(c, 200, "Master ventilator berhasil dihapus", h.repo.HapusMaster(c, c.Param("kode")))
}
func (h *Handler) pasien(c *gin.Context) {
	no := strings.TrimSpace(c.Query("no_rawat"))
	if no == "" {
		httpresponse.Error(c, 422, "VENTILATOR_VALIDATION_ERROR", "Nomor rawat wajib diisi")
		return
	}
	d, e := h.repo.DataPasien(c, no)
	h.hasil(c, d, e)
}
func (h *Handler) simpanPemakaian(c *gin.Context) {
	var x ventilator.Pemakaian
	if c.ShouldBindJSON(&x) != nil || x.NoRawat == "" || x.KodeVentilator == "" || x.TanggalMulai == "" {
		httpresponse.Error(c, 422, "VENTILATOR_VALIDATION_ERROR", "Nomor rawat, ventilator, dan tanggal mulai wajib diisi")
		return
	}
	x.TanggalMulai = waktuMySQL(x.TanggalMulai)
	if nilai := strings.TrimSpace(x.KedalamanJalanNapas); nilai != "" {
		kedalaman, err := strconv.ParseFloat(strings.ReplaceAll(nilai, ",", "."), 64)
		if err != nil || kedalaman < 0 || kedalaman > 999.99 {
			httpresponse.Error(c, 422, "VENTILATOR_VALIDATION_ERROR", "Kedalaman jalan napas harus berupa angka antara 0 sampai 999,99")
			return
		}
		x.KedalamanJalanNapas = strconv.FormatFloat(kedalaman, 'f', -1, 64)
	}
	if x.PetugasPemasangan == "" {
		x.PetugasPemasangan = username(c)
	}
	h.pesan(c, 201, "Pemakaian ventilator berhasil dimulai", h.repo.SimpanPemakaian(c, x))
}
func (h *Handler) simpanSetting(c *gin.Context) {
	var x ventilator.Setting
	if c.ShouldBindJSON(&x) != nil || x.IDPemakaian == 0 || x.Waktu == "" || x.Mode == "" {
		httpresponse.Error(c, 422, "VENTILATOR_VALIDATION_ERROR", "Pemakaian, waktu, dan mode ventilator wajib diisi")
		return
	}
	x.Waktu = waktuMySQL(x.Waktu)
	if x.Petugas == "" {
		x.Petugas = username(c)
	}
	h.pesan(c, 201, "Setting ventilator berhasil disimpan", h.repo.SimpanSetting(c, x))
}
func (h *Handler) simpanMonitoring(c *gin.Context) {
	var x ventilator.Monitoring
	if c.ShouldBindJSON(&x) != nil || x.IDPemakaian == 0 || x.Waktu == "" {
		httpresponse.Error(c, 422, "VENTILATOR_VALIDATION_ERROR", "Pemakaian dan waktu monitoring wajib diisi")
		return
	}
	x.Waktu = waktuMySQL(x.Waktu)
	if x.Petugas == "" {
		x.Petugas = username(c)
	}
	h.pesan(c, 201, "Monitoring ventilator berhasil disimpan", h.repo.SimpanMonitoring(c, x))
}
func (h *Handler) simpanVAP(c *gin.Context) {
	var x ventilator.ChecklistVAP
	if c.ShouldBindJSON(&x) != nil || x.IDPemakaian == 0 || x.Waktu == "" {
		httpresponse.Error(c, 422, "VENTILATOR_VALIDATION_ERROR", "Pemakaian dan waktu checklist wajib diisi")
		return
	}
	x.Waktu = waktuMySQL(x.Waktu)
	if x.Petugas == "" {
		x.Petugas = username(c)
	}
	h.pesan(c, 201, "Checklist VAP berhasil disimpan", h.repo.SimpanVAP(c, x))
}
func (h *Handler) hasil(c *gin.Context, d any, e error) {
	if e != nil {
		h.err(c, e)
		return
	}
	httpresponse.Success(c, 200, d)
}
func (h *Handler) pesan(c *gin.Context, status int, p string, e error) {
	if e != nil {
		h.err(c, e)
		return
	}
	httpresponse.Success(c, status, gin.H{"pesan": p})
}
func (h *Handler) err(c *gin.Context, e error) {
	log.Printf("ventilator gagal diproses: %v", e)
	if e == sql.ErrNoRows {
		httpresponse.Error(c, 404, "VENTILATOR_NOT_FOUND", "Data ventilator tidak ditemukan")
		return
	}
	httpresponse.Error(c, 503, "VENTILATOR_SIMRS_UNAVAILABLE", "Data ventilator tidak dapat diproses pada SIMRS Khanza")
}

func waktuMySQL(value string) string { return strings.Replace(strings.TrimSpace(value), "T", " ", 1) }
func username(c *gin.Context) string {
	value, ok := c.Get("authenticated_user")
	if !ok {
		return ""
	}
	user, ok := value.(autentikasi.Pengguna)
	if !ok {
		return ""
	}
	return strings.TrimSpace(user.Username)
}
