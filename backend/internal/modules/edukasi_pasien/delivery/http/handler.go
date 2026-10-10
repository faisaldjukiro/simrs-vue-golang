package edukasipasienhttp

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"simrs-backend/internal/modules/autentikasi"
	"simrs-backend/internal/modules/edukasi_pasien"
	"simrs-backend/internal/shared/httpresponse"
	"strings"
	"time"
)

type Handler struct {
	repo *edukasi_pasien.Repositori
	foto pengaturanFoto
}

var errUploadFoto = errors.New("foto gagal diupload")

func NewHandler(repo *edukasi_pasien.Repositori, uploadURL, prefix, webBaseURL string) *Handler {
	return &Handler{repo: repo, foto: pengaturanFoto{uploadURL: uploadURL, prefix: strings.Trim(prefix, "/"), webBaseURL: webBaseURL, client: &http.Client{Timeout: 60 * time.Second}}}
}
func (h *Handler) Register(g *gin.RouterGroup) {
	g.GET("", h.proses)
	g.GET("/referensi", h.proses)
	g.POST("", h.proses)
	g.PUT("", h.proses)
	g.DELETE("", h.proses)
}
func (h *Handler) proses(c *gin.Context) {
	v, ada := c.Get("authenticated_user")
	u, ok := v.(autentikasi.Pengguna)
	if !ada || !ok || u.ID == 0 {
		httpresponse.Error(c, 401, "UNAUTHENTICATED", "Sesi login tidak valid")
		return
	}
	admin := false
	for _, p := range u.Permissions {
		if p == "*" {
			admin = true
		}
	}
	durasi := 20 * time.Second
	if c.ContentType() == "multipart/form-data" {
		durasi = 150 * time.Second
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), durasi)
	defer cancel()
	if c.Request.Method == http.MethodGet && c.FullPath() == "/api/edukasi-pasien/referensi" {
		hasil, err := h.repo.Referensi(ctx, c.Query("jenis"), c.Query("q"), u.Username, admin)
		if err != nil {
			tulisError(c, err)
			return
		}
		httpresponse.Success(c, 200, hasil)
		return
	}
	if c.Request.Method == http.MethodGet {
		hasil, err := h.repo.Daftar(ctx, c.Query("no_rawat"), u.Username, admin)
		if err != nil {
			tulisError(c, err)
			return
		}
		for i := range hasil.Catatan {
			hasil.Catatan[i].FotoURL = h.urlFoto(hasil.Catatan[i].Data["foto"])
			hasil.Catatan[i].FotoPenerimaURL = h.urlFoto(hasil.Catatan[i].Data["foto_penerima"])
		}
		httpresponse.Success(c, 200, hasil)
		return
	}
	defer func() {
		if c.Request.MultipartForm != nil {
			c.Request.MultipartForm.RemoveAll()
		}
	}()
	in, files, err := bacaInput(c)
	if err != nil {
		tulisError(c, err)
		return
	}
	unggah := map[string]func() (string, error){}
	for key, file := range files {
		nama, err := siapkanFoto(file)
		if err != nil {
			tulisError(c, err)
			return
		}
		if len(h.foto.prefix)+1+len(nama) > 255 {
			tulisError(c, fmt.Errorf("%w: konfigurasi lokasi foto melebihi 255 karakter", edukasi_pasien.ErrValidasi))
			return
		}
		unggah[key] = func() (string, error) { return h.unggahFoto(ctx, file, nama, in.NoRawat) }
	}
	if err := h.repo.MutasiDenganBukti(ctx, in, c.Request.Method, u.Username, admin, unggah); err != nil {
		tulisError(c, err)
		return
	}
	pesan := "Catatan edukasi pasien berhasil disimpan di SIMRS"
	if c.Request.Method == http.MethodDelete {
		pesan = "Catatan edukasi pasien berhasil dihapus dari SIMRS"
	}
	httpresponse.Success(c, 200, gin.H{"pesan": pesan})
}
func tulisError(c *gin.Context, err error) {
	status, kode, pesan := 500, "EDUKASI_PASIEN_ERROR", "Catatan edukasi pasien belum dapat diproses. Periksa koneksi, izin CRUD, dan tabel SIMRS, lalu muat ulang riwayat."
	switch {
	case errors.Is(err, edukasi_pasien.ErrValidasi):
		status, kode, pesan = 422, "VALIDATION_ERROR", err.Error()
	case errors.Is(err, edukasi_pasien.ErrAkses):
		status, kode, pesan = 403, "FORBIDDEN", err.Error()
	case errors.Is(err, edukasi_pasien.ErrKonflik):
		status, kode, pesan = 409, "CONFLICT", err.Error()
	case errors.Is(err, edukasi_pasien.ErrDuplikat):
		status, kode, pesan = 409, "EDUKASI_WAKTU_DUPLIKAT", err.Error()
	case errors.Is(err, errUploadFoto):
		status, kode, pesan = 502, "EDUKASI_FOTO_UPLOAD_ERROR", err.Error()+". Perubahan catatan dibatalkan."
	}
	if status == 500 {
		var dbErr *mysql.MySQLError
		if errors.As(err, &dbErr) {
			// Jangan mencatat SQL, nilai input, kredensial, atau identitas pasien.
			log.Printf("edukasi_pasien method=%s mysql_code=%d", c.Request.Method, dbErr.Number)
			switch dbErr.Number {
			case 1044, 1045, 1142, 1143:
				pesan = "Akun koneksi SIMRS tidak memiliki izin database yang diperlukan. Periksa izin SELECT/INSERT/UPDATE/DELETE."
			case 1146:
				pesan = "Tabel yang diperlukan Edukasi Pasien tidak ditemukan pada koneksi SIMRS. Periksa database tujuan dan tabel catatan_edukasi, petugas, ruangan, serta reg_periksa."
			case 1054, 1136, 1364:
				pesan = "Struktur tabel SIMRS tidak sesuai kolom Edukasi Pasien. Periksa SHOW CREATE TABLE catatan_edukasi; jangan membuat tabel pengganti."
			case 1406, 1265, 1366, 1292:
				status = 422
				pesan = "Isi catatan tidak sesuai tipe atau panjang kolom SIMRS. Periksa durasi, tanggal/jam, metode, dan panjang teks terhadap struktur catatan_edukasi."
			case 1452:
				status = 422
				pesan = "Referensi catatan ditolak oleh SIMRS. Periksa nomor rawat, petugas, dan ruangan yang dipilih."
			}
			kode = fmt.Sprintf("EDUKASI_MYSQL_%d", dbErr.Number)
			pesan += fmt.Sprintf(" (MySQL %d)", dbErr.Number)
		} else if errors.Is(err, context.DeadlineExceeded) {
			status, kode, pesan = 504, "EDUKASI_TIMEOUT", "SIMRS belum merespons dalam batas waktu. Muat ulang riwayat sebelum mencoba simpan kembali."
		}
	}
	if errors.Is(err, edukasi_pasien.ErrFotoTerunggah) {
		pesan += ". " + edukasi_pasien.ErrFotoTerunggah.Error()
	}
	httpresponse.Error(c, status, kode, pesan)
}
