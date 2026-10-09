package rujukaninternalhttp

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/modules/rujukan_internal"
	"simrs-backend/internal/shared/httpresponse"
)

func (h *Handler) konsultasi(c *gin.Context) {
	v, _ := c.Get("authenticated_user")
	u, ok := v.(autentikasi.Pengguna)
	if !ok || u.ID == 0 {
		httpresponse.Error(c, 401, "UNAUTHENTICATED", "Sesi login tidak valid")
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
	if c.Request.Method == http.MethodGet {
		var data any
		var err error
		if strings.HasSuffix(c.FullPath(), "/asal") {
			data, err = h.repositori.ReferensiAsalKonsul(ctx, h.jenis, c.Query("q"))
		} else {
			data, err = h.repositori.DaftarKonsul(ctx, h.jenis, c.Query("no_rawat"), u.Username, admin)
		}
		if err != nil {
			gagalKonsul(c, err)
			return
		}
		httpresponse.Success(c, 200, data)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 131072)
	var in rujukan_internal.InputKonsul
	if c.ShouldBindJSON(&in) != nil {
		httpresponse.Error(c, 422, "KONSULTASI_INVALID", "Format konsultasi tidak sesuai")
		return
	}
	aksi := "baru"
	if c.Request.Method == http.MethodPut {
		aksi = "ubah"
	}
	if strings.HasSuffix(c.FullPath(), "/jawaban") {
		aksi = "jawab"
	}
	nomor, err := h.repositori.SimpanKonsul(ctx, h.jenis, u.Username, admin, aksi, in)
	if err != nil {
		gagalKonsul(c, err)
		return
	}
	pesan := "Surat konsultasi dan rujukan berhasil disimpan di SIMRS."
	if aksi == "ubah" {
		pesan = "Isi surat konsultasi berhasil diperbarui."
	}
	if aksi == "jawab" {
		pesan = "Jawaban konsultasi tersimpan. Status surat menjadi Selesai."
	}
	httpresponse.Success(c, 200, gin.H{"no_surat": nomor, "pesan": pesan})
}

func gagalKonsul(c *gin.Context, err error) {
	var dbErr *mysql.MySQLError
	status, pesan := 503, "Konsultasi belum dapat diproses. Periksa koneksi, izin CRUD, dan struktur tabel surat_konsul di SIMRS."
	switch {
	case errors.Is(err, rujukan_internal.ErrInput):
		status, pesan = 422, err.Error()
	case errors.Is(err, rujukan_internal.ErrAksesKonsul):
		status, pesan = 403, err.Error()
	case errors.Is(err, rujukan_internal.ErrTidakAda):
		status, pesan = 404, err.Error()
	case errors.Is(err, rujukan_internal.ErrBerubah), errors.Is(err, rujukan_internal.ErrDuplikat), errors.Is(err, rujukan_internal.ErrBilling), errors.Is(err, rujukan_internal.ErrKunjungan):
		status, pesan = 409, err.Error()
	case errors.As(err, &dbErr):
		switch dbErr.Number {
		case 1062:
			status, pesan = 409, "Nomor surat atau rujukan sudah digunakan. Muat ulang lalu coba kembali."
		case 1406, 1366:
			status, pesan = 422, "Isi konsultasi melampaui panjang kolom atau mengandung karakter yang tidak didukung SIMRS."
		case 1292:
			status, pesan = 422, "SIMRS menolak tanggal surat atau tanggal jawaban kosong (0000-00-00). Struktur surat_konsul lama memerlukan tanggal kosong sebelum dijawab; periksa kompatibilitas SQL mode bersama administrator. Tidak ada data yang disimpan."
		}
	}
	httpresponse.Error(c, status, "KONSULTASI_ERROR", pesan)
}
