package aktivitas_log

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/modules/autentikasi"
)

const batasResponseError = 16 * 1024

type responseWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (w *responseWriter) Write(data []byte) (int, error) {
	if w.body.Len() < batasResponseError {
		sisa := batasResponseError - w.body.Len()
		if len(data) < sisa {
			sisa = len(data)
		}
		_, _ = w.body.Write(data[:sisa])
	}
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteString(data string) (int, error) {
	if w.body.Len() < batasResponseError {
		sisa := batasResponseError - w.body.Len()
		if len(data) < sisa {
			sisa = len(data)
		}
		_, _ = w.body.WriteString(data[:sisa])
	}
	return w.ResponseWriter.WriteString(data)
}

func Middleware(repositori *Repositori) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		mulai := time.Now()
		requestID := buatRequestID()
		c.Header("X-Request-ID", requestID)

		requestData := bacaRequestData(c.Request)
		writer := &responseWriter{ResponseWriter: c.Writer}
		c.Writer = writer
		c.Next()

		pengguna, _ := c.Get(autentikasi.ContextKeyPengguna)
		user, _ := pengguna.(autentikasi.Pengguna)
		username := user.Username
		var userID *uint64
		if user.ID > 0 {
			id := user.ID
			userID = &id
		}
		if username == "" {
			username = c.GetString("audit_username")
		}

		status := c.Writer.Status()
		aksi := namaAksi(c.Request.Method, c.Request.URL.Path)
		module := namaModul(c.Request.URL.Path)
		targetID := cariTargetID(c, requestData)
		queryParams := marshalAman(sanitasiQuery(c.Request.URL.Query()))
		requestJSON := marshalAman(requestData)
		var dataSesudah json.RawMessage
		if status < http.StatusBadRequest && (aksi == "TAMBAH" || aksi == "UBAH" || aksi == "PROSES") {
			dataSesudah = requestJSON
		}

		catatan := Catatan{
			UserID:         userID,
			RequestID:      requestID,
			Username:       username,
			Aksi:           aksi,
			Modul:          module,
			Method:         c.Request.Method,
			Endpoint:       endpoint(c),
			TabelTarget:    tabelTarget(module),
			TargetID:       targetID,
			QueryParams:    queryParams,
			RequestData:    requestJSON,
			DataSesudah:    dataSesudah,
			ResponseStatus: status,
			Berhasil:       status < http.StatusBadRequest,
			DurasiMS:       uint64(time.Since(mulai).Milliseconds()),
			PesanError:     bacaPesanError(writer.body.Bytes()),
			IPAddress:      c.ClientIP(),
			UserAgent:      potong(c.Request.UserAgent(), 255),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := repositori.Simpan(ctx, catatan); err != nil {
			log.Printf("audit trail tidak dapat disimpan: %v", err)
		}
	}
}

func bacaRequestData(request *http.Request) map[string]any {
	if request.Body == nil || !strings.Contains(strings.ToLower(request.Header.Get("Content-Type")), "application/json") {
		return nil
	}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil
	}
	request.Body = io.NopCloser(bytes.NewReader(body))
	if len(body) == 0 {
		return nil
	}
	var data map[string]any
	if json.Unmarshal(body, &data) != nil {
		return nil
	}
	return sanitasiMap(data)
}

func sanitasiMap(data map[string]any) map[string]any {
	if len(data) == 0 {
		return nil
	}
	hasil := make(map[string]any, len(data))
	for key, value := range data {
		if rahasia(key) {
			hasil[key] = "[DISAMARKAN]"
			continue
		}
		hasil[key] = sanitasiNilai(value)
	}
	return hasil
}

func sanitasiNilai(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return sanitasiMap(typed)
	case []any:
		hasil := make([]any, len(typed))
		for i := range typed {
			hasil[i] = sanitasiNilai(typed[i])
		}
		return hasil
	default:
		return value
	}
}

func sanitasiQuery(query url.Values) map[string]any {
	hasil := make(map[string]any, len(query))
	for key, value := range query {
		if rahasia(key) {
			hasil[key] = "[DISAMARKAN]"
		} else if len(value) == 1 {
			hasil[key] = value[0]
		} else {
			hasil[key] = value
		}
	}
	return hasil
}

func rahasia(key string) bool {
	kunci := strings.ToLower(strings.ReplaceAll(key, "-", "_"))
	for _, bagian := range []string{"password", "token", "authorization", "secret", "signature", "user_key", "consumer_key"} {
		if strings.Contains(kunci, bagian) {
			return true
		}
	}
	return false
}

func namaAksi(method, path string) string {
	if strings.HasSuffix(path, "/auth/login") {
		return "LOGIN"
	}
	if strings.HasSuffix(path, "/auth/logout") {
		return "LOGOUT"
	}
	switch method {
	case http.MethodGet, http.MethodHead:
		return "LIHAT"
	case http.MethodPost:
		if strings.Contains(path, "/proses") || strings.Contains(path, "/group") || strings.Contains(path, "/final") {
			return "PROSES"
		}
		return "TAMBAH"
	case http.MethodPut, http.MethodPatch:
		return "UBAH"
	case http.MethodDelete:
		return "HAPUS"
	default:
		return strings.ToUpper(method)
	}
}

func namaModul(path string) string {
	bagian := strings.Split(strings.Trim(path, "/"), "/")
	if len(bagian) >= 2 && bagian[0] == "api" {
		if bagian[1] == "bpjs" && len(bagian) >= 4 {
			return strings.Join(bagian[1:4], "/")
		}
		return bagian[1]
	}
	return "sistem"
}

func tabelTarget(modul string) string {
	return map[string]string{
		"auth":                      "access_tokens",
		"user-management":           "users",
		"cppt":                      "pemeriksaan_ranap/pemeriksaan_ralan",
		"diagnosa-pasien":           "diagnosa_pasien/prosedur_pasien",
		"penanganan-dokter-petugas": "rawat_inap_drpr/rawat_jl_drpr",
		"permintaan-radiologi":      "permintaan_radiologi",
		"resume-pasien-ranap":       "resume_pasien",
		"triase-igd":                "data_triase_igd",
		"awal-keperawatan-igd":      "penilaian_awal_keperawatan_igd",
		"bpjs/monitoring/klaim":     "bpjs_vclaim",
	}[modul]
}

func cariTargetID(c *gin.Context, data map[string]any) string {
	for _, key := range []string{"id", "no_rawat", "nomor", "no_sep", "username", "kode"} {
		if value := strings.TrimSpace(c.Param(key)); value != "" {
			return potong(value, 191)
		}
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			return potong(value, 191)
		}
		if value, ok := data[key]; ok {
			if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
				return potong(strings.TrimSpace(text), 191)
			}
		}
	}
	return ""
}

func endpoint(c *gin.Context) string {
	if route := c.FullPath(); route != "" {
		return route
	}
	return c.Request.URL.Path
}

func bacaPesanError(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var response struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &response) == nil {
		return potong(response.Error.Message, 500)
	}
	return ""
}

func marshalAman(value any) json.RawMessage {
	if value == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil || string(data) == "{}" {
		return nil
	}
	return data
}

func buatRequestID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405.000000000")))[:32]
	}
	return hex.EncodeToString(buffer)
}

func potong(value string, batas int) string {
	if len(value) <= batas {
		return value
	}
	return value[:batas]
}
