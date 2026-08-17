package berandahttp

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/modules/beranda"
	"simrs-backend/internal/shared/httpresponse"
)

type Handler struct {
	repositori *beranda.Repositori
}

func NewHandler(repositori *beranda.Repositori) *Handler {
	return &Handler{repositori: repositori}
}

func (h *Handler) Beranda(c *gin.Context) {
	pengguna, ada := c.Get(autentikasi.ContextKeyPengguna)
	user, valid := pengguna.(autentikasi.Pengguna)
	if !ada || !valid {
		httpresponse.Error(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Pengguna tidak terautentikasi")
		return
	}

	startedAt := time.Now()
	filter := bacaFilterBeranda(c)
	data, err := h.repositori.BacaBeranda(c.Request.Context(), filter, user.ID)
	if err != nil {
		httpresponse.Error(c, http.StatusServiceUnavailable, "SIMRS_UNAVAILABLE", "Database SIMRS lama tidak dapat dibaca")
		return
	}

	httpresponse.Success(c, http.StatusOK, gin.H{
		"koneksi_database": gin.H{
			"terhubung":  true,
			"latensi_ms": time.Since(startedAt).Milliseconds(),
		},
		"ringkasan":      data.Ringkasan,
		"registrasi":     data.Registrasi,
		"rawat_jalan":    data.RawatJalan,
		"igd":            data.IGD,
		"rawat_inap":     data.RawatInap,
		"paginasi":       data.Paginasi,
		"poliklinik":     data.Poliklinik,
		"dokter":         data.Dokter,
		"pilihan_status": data.PilihanStatus,
		"sidebar_pasien": data.SidebarPasien,
		"filter": gin.H{
			"rawat_jalan": filter.RawatJalan,
			"igd":         filter.IGD,
			"rawat_inap":  filter.RawatInap,
		},
	})
}

func bacaFilterBeranda(c *gin.Context) beranda.FilterBeranda {
	hariIni := time.Now().Format("2006-01-02")
	tabAktif := c.Query("tab")

	return beranda.FilterBeranda{
		RawatJalan: bacaFilterPasien(c, "rj", hariIni, tabAktif == "Rawat Jalan", false),
		IGD:        bacaFilterPasien(c, "igd", hariIni, tabAktif == "IGD/UGD", false),
		RawatInap:  bacaFilterPasien(c, "ri", hariIni, tabAktif == "Rawat Inap", true),
	}
}

func bacaFilterPasien(c *gin.Context, prefix string, tanggalDefault string, pakaiFilterUmum bool, defaultBelumPulang bool) beranda.FilterPasien {
	tanggalMulai, manualMulai := nilaiTanggal(c, prefix+"_date_from", "")
	tanggalSelesai, manualSelesai := nilaiTanggal(c, prefix+"_date_to", "")

	if tanggalMulai == "" && pakaiFilterUmum {
		tanggalMulai, manualMulai = nilaiTanggal(c, "date_from", "")
	}
	if tanggalSelesai == "" && pakaiFilterUmum {
		tanggalSelesai, manualSelesai = nilaiTanggal(c, "date_to", "")
	}
	if tanggalMulai == "" {
		tanggalMulai = tanggalDefault
	}
	if tanggalSelesai == "" || tanggalSelesai < tanggalMulai {
		tanggalSelesai = tanggalMulai
	}

	status := c.Query(prefix + "_status")
	statusBayar := c.Query(prefix + "_status_bayar")
	dokter := c.Query(prefix + "_dokter")
	poliklinik := c.Query(prefix + "_poly")
	pencarian := c.Query(prefix + "_search")
	if pakaiFilterUmum {
		if status == "" {
			status = c.Query("status")
		}
		if statusBayar == "" {
			statusBayar = c.Query("status_bayar")
		}
		if dokter == "" {
			dokter = c.Query("dokter")
		}
		if poliklinik == "" {
			poliklinik = c.Query("poly")
		}
		if pencarian == "" {
			pencarian = c.Query("search")
		}
	}

	belumPulang := defaultBelumPulang
	if nilai, ada := c.GetQuery(prefix + "_belum_pulang"); ada {
		belumPulang = nilai == "1" || nilai == "true" || nilai == "on"
	}

	return beranda.FilterPasien{
		TanggalMulai:         tanggalMulai,
		TanggalSelesai:       tanggalSelesai,
		Status:               status,
		StatusBayar:          statusBayar,
		Poliklinik:           poliklinik,
		Dokter:               dokter,
		Pencarian:            pencarian,
		BelumPulang:          belumPulang,
		TanggalDipilihManual: manualMulai || manualSelesai,
		Halaman:              nilaiInt(c, prefix+"_page", nilaiInt(c, "page", 1)),
		Batas:                nilaiInt(c, prefix+"_limit", nilaiInt(c, "limit", 100)),
	}
}

func nilaiTanggal(c *gin.Context, nama string, fallback string) (string, bool) {
	nilai := c.Query(nama)
	if nilai == "" {
		return fallback, false
	}
	if _, err := time.Parse("2006-01-02", nilai); err != nil {
		return fallback, false
	}
	return nilai, true
}

func nilaiInt(c *gin.Context, nama string, fallback int) int {
	nilai := c.Query(nama)
	if nilai == "" {
		return fallback
	}
	angka, err := strconv.Atoi(nilai)
	if err != nil {
		return fallback
	}
	return angka
}
