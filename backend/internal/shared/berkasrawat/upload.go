// Package berkasrawat mengirim file ke endpoint upload Berkas Digital SIMRS.
package berkasrawat

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

func Kirim(ctx context.Context, client *http.Client, alamat string, file io.Reader, nama, noRawat, kode string) error {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	if err := w.WriteField("no_rawat", noRawat); err != nil {
		return err
	}
	if err := w.WriteField("kode", kode); err != nil {
		return err
	}
	part, err := w.CreateFormFile("file", nama)
	if err != nil {
		return err
	}
	if _, err = io.Copy(part, file); err != nil {
		return fmt.Errorf("file upload tidak dapat dibaca")
	}
	if err = w.Close(); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, alamat, &body)
	if err != nil {
		return fmt.Errorf("alamat upload SIMRS tidak valid")
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Accept", "text/plain")
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("server berkas SIMRS tidak dapat dihubungi")
	}
	defer res.Body.Close()
	isi, err := io.ReadAll(io.LimitReader(res.Body, 2048))
	if err != nil {
		return fmt.Errorf("respons server berkas SIMRS tidak dapat dibaca")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("server berkas SIMRS menolak upload (HTTP %d)", res.StatusCode)
	}
	if strings.TrimSpace(string(isi)) != "UPLOAD_BERHASIL" {
		return fmt.Errorf("server berkas SIMRS tidak mengonfirmasi keberhasilan upload")
	}
	return nil
}
