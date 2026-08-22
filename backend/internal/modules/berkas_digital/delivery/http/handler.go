package berkasdigitalhttp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/berkas_digital"
	"simrs-backend/internal/shared/httpresponse"
)

const batasUpload = 25 << 20

type Handler struct {
	layanan      *berkas_digital.Layanan
	uploadURL    string
	lokasiPrefix string
	client       *http.Client
}

func NewHandler(layanan *berkas_digital.Layanan, uploadURL, lokasiPrefix string) *Handler {
	uploadURL = strings.TrimSpace(uploadURL)
	if uploadURL == "" {
		uploadURL = "http://127.0.0.1/webapps/berkasrawat/uploadsep.php"
	}
	lokasiPrefix = strings.Trim(strings.TrimSpace(lokasiPrefix), "/")
	if lokasiPrefix == "" {
		lokasiPrefix = "pages/upload"
	}
	return &Handler{
		layanan:      layanan,
		uploadURL:    uploadURL,
		lokasiPrefix: lokasiPrefix,
		client:       &http.Client{Timeout: 60 * time.Second},
	}
}

func (h *Handler) Register(group *gin.RouterGroup) {
	group.GET("", h.Data)
	group.POST("", h.Upload)
	group.DELETE("", h.Hapus)
}

func (h *Handler) Data(c *gin.Context) {
	data, err := h.layanan.Data(c.Request.Context(), c.Query("no_rawat"))
	if err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, data)
}

func (h *Handler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, batasUpload)

	noRawat := strings.TrimSpace(c.PostForm("no_rawat"))
	kode := strings.TrimSpace(c.PostForm("kode"))
	file, err := c.FormFile("file")
	if err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "BERKAS_DIGITAL_VALIDATION_ERROR", "File berkas digital wajib dipilih.")
		return
	}
	ekstensi := strings.ToLower(filepath.Ext(file.Filename))
	if !ekstensiDiizinkan(ekstensi) {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "BERKAS_DIGITAL_FILE_TYPE_ERROR", "Format file hanya boleh PDF, JPG, JPEG, atau PNG.")
		return
	}
	if err := h.layanan.PastikanBisaUpload(c.Request.Context(), noRawat, kode); err != nil {
		h.tulisError(c, err)
		return
	}

	namaFile := h.namaFileAman(noRawat, kode, file.Filename)
	lokasiFile := h.lokasiPrefix + "/" + namaFile
	if err := h.kirimKeServerKhanza(c, file, namaFile, noRawat, kode); err != nil {
		httpresponse.Error(c, http.StatusBadGateway, "BERKAS_DIGITAL_UPLOAD_SERVER_ERROR", err.Error())
		return
	}

	input := berkas_digital.Input{NoRawat: noRawat, Kode: kode, LokasiFile: lokasiFile}
	if err := h.layanan.Simpan(c.Request.Context(), input); err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusCreated, gin.H{
		"pesan":       "Berkas digital berhasil diupload.",
		"lokasi_file": lokasiFile,
	})
}

func (h *Handler) Hapus(c *gin.Context) {
	var kunci berkas_digital.Kunci
	if err := c.ShouldBindJSON(&kunci); err != nil {
		httpresponse.Error(c, http.StatusUnprocessableEntity, "BERKAS_DIGITAL_VALIDATION_ERROR", "Data berkas digital yang akan dihapus tidak lengkap.")
		return
	}
	if err := h.layanan.Hapus(c.Request.Context(), kunci); err != nil {
		h.tulisError(c, err)
		return
	}
	httpresponse.Success(c, http.StatusOK, gin.H{"pesan": "Berkas digital berhasil dihapus dari perawatan."})
}

func (h *Handler) tulisError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, berkas_digital.ErrInputTidakValid):
		httpresponse.Error(c, http.StatusUnprocessableEntity, "BERKAS_DIGITAL_VALIDATION_ERROR", err.Error())
	case errors.Is(err, berkas_digital.ErrMasterTidakDitemukan):
		httpresponse.Error(c, http.StatusUnprocessableEntity, "BERKAS_DIGITAL_MASTER_NOT_FOUND", err.Error())
	case errors.Is(err, berkas_digital.ErrBillingTerkunci):
		httpresponse.Error(c, http.StatusConflict, "BERKAS_DIGITAL_BILLING_LOCKED", err.Error())
	case errors.Is(err, berkas_digital.ErrBerkasTidakDitemukan):
		httpresponse.Error(c, http.StatusNotFound, "BERKAS_DIGITAL_NOT_FOUND", err.Error())
	default:
		httpresponse.Error(c, http.StatusServiceUnavailable, "BERKAS_DIGITAL_SIMRS_UNAVAILABLE", "Berkas digital tidak dapat diproses pada SIMRS Khanza.")
	}
}

func (h *Handler) kirimKeServerKhanza(c *gin.Context, fileHeader *multipart.FileHeader, namaFile, noRawat, kode string) error {
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("file upload tidak dapat dibaca")
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("no_rawat", noRawat); err != nil {
		return fmt.Errorf("payload upload tidak dapat disiapkan")
	}
	if err := writer.WriteField("kode", kode); err != nil {
		return fmt.Errorf("payload upload tidak dapat disiapkan")
	}
	part, err := writer.CreateFormFile("file", namaFile)
	if err != nil {
		return fmt.Errorf("payload file tidak dapat disiapkan")
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("file upload tidak dapat dibaca")
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("payload upload tidak dapat ditutup")
	}

	request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, h.uploadURL, &body)
	if err != nil {
		return fmt.Errorf("request upload ke SIMRS Khanza tidak valid")
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Accept", "text/plain")

	response, err := h.client.Do(request)
	if err != nil {
		return fmt.Errorf("upload ke server berkas SIMRS Khanza gagal: %w", err)
	}
	defer response.Body.Close()
	responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
	pesan := strings.TrimSpace(string(responseBody))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("server berkas SIMRS Khanza menolak upload: HTTP %d %s", response.StatusCode, pesan)
	}
	if pesan != "UPLOAD_BERHASIL" {
		if pesan == "" {
			pesan = "respon kosong"
		}
		return fmt.Errorf("server berkas SIMRS Khanza gagal upload: %s", pesan)
	}
	return nil
}

func (h *Handler) namaFileAman(noRawat, kode, namaAsli string) string {
	ekstensi := strings.ToLower(filepath.Ext(namaAsli))
	namaDasar := strings.TrimSuffix(filepath.Base(namaAsli), filepath.Ext(namaAsli))
	namaDasar = karakterAman(namaDasar)
	if namaDasar == "" {
		namaDasar = "berkas"
	}
	kunci := karakterAman(strings.ReplaceAll(noRawat, "/", "-") + "-" + kode)
	return kunci + "-" + time.Now().Format("20060102150405") + "-" + namaDasar + ekstensi
}

func karakterAman(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-' || r == '_':
			builder.WriteRune(r)
		case r == ' ' || r == '.':
			builder.WriteRune('-')
		}
	}
	return strings.Trim(builder.String(), "-")
}

func ekstensiDiizinkan(ekstensi string) bool {
	switch ekstensi {
	case ".pdf", ".jpg", ".jpeg", ".png":
		return true
	default:
		return false
	}
}
