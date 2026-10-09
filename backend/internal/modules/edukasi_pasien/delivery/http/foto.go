package edukasipasienhttp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"simrs-backend/internal/modules/edukasi_pasien"
	"simrs-backend/internal/shared/berkasrawat"
)

const batasFoto = 10 << 20

type pengaturanFoto struct {
	uploadURL, prefix, webBaseURL string
	client                        *http.Client
}

func bacaInput(c *gin.Context) (edukasi_pasien.Input, *multipart.FileHeader, error) {
	var in edukasi_pasien.Input
	if c.ContentType() != "multipart/form-data" {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 128<<10)
		if err := c.ShouldBindJSON(&in); err != nil {
			return in, nil, edukasi_pasien.ErrValidasi
		}
		return in, nil, nil
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, batasFoto+(256<<10))
	if err := c.Request.ParseMultipartForm(1 << 20); err != nil {
		return in, nil, fmt.Errorf("%w: foto maksimal 10 MB dan formulir harus lengkap", edukasi_pasien.ErrValidasi)
	}
	if c.Request.Method == http.MethodDelete {
		return in, nil, edukasi_pasien.ErrValidasi
	}
	payload := c.PostForm("payload")
	if len(payload) > 128<<10 || json.Unmarshal([]byte(payload), &in) != nil {
		return in, nil, edukasi_pasien.ErrValidasi
	}
	files := c.Request.MultipartForm.File["foto"]
	if len(files) != 1 || len(c.Request.MultipartForm.File) != 1 {
		return in, nil, fmt.Errorf("%w: pilih satu foto JPG atau PNG", edukasi_pasien.ErrValidasi)
	}
	if files[0].Size <= 0 || files[0].Size > batasFoto {
		return in, nil, fmt.Errorf("%w: foto maksimal 10 MB", edukasi_pasien.ErrValidasi)
	}
	return in, files[0], nil
}

func siapkanFoto(file *multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("%w: foto tidak dapat dibaca", edukasi_pasien.ErrValidasi)
	}
	defer f.Close()
	cfg, format, err := image.DecodeConfig(f)
	if err != nil || (format != "jpeg" && format != "png") || cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > 40000000 {
		return "", fmt.Errorf("%w: isi foto harus JPG/PNG yang valid, maksimal 40 megapiksel", edukasi_pasien.ErrValidasi)
	}
	ext := ".jpg"
	if format == "png" {
		ext = ".png"
	}
	acak := make([]byte, 16)
	if _, err = rand.Read(acak); err != nil {
		return "", err
	}
	return "edukasi-" + hex.EncodeToString(acak) + ext, nil
}

func (h *Handler) unggahFoto(ctx context.Context, file *multipart.FileHeader, nama, noRawat string) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err = berkasrawat.Kirim(ctx, h.foto.client, h.foto.uploadURL, f, nama, noRawat, "edukasi"); err != nil {
		return "", fmt.Errorf("%w: %v", errUploadFoto, err)
	}
	return h.foto.prefix + "/" + nama, nil
}

func (h *Handler) urlFoto(lokasi string) string {
	if lokasi == "" || lokasi == "-" || strings.ContainsAny(lokasi, "\\?#") {
		return ""
	}
	for _, bagian := range strings.Split(lokasi, "/") {
		if bagian == ".." {
			return ""
		}
	}
	if strings.HasPrefix(lokasi, "/") || strings.Contains(lokasi, ":") {
		return ""
	}
	base, err := url.Parse(h.foto.webBaseURL)
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" {
		return ""
	}
	base.Path = path.Join(base.Path, "berkasrawat", lokasi)
	return base.String()
}
