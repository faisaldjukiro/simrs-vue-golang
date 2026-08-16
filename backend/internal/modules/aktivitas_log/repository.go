package aktivitas_log

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Catatan struct {
	ID             uint64          `json:"id"`
	UserID         *uint64         `json:"user_id"`
	RequestID      string          `json:"request_id"`
	Username       string          `json:"username"`
	NamaUser       string          `json:"nama_user"`
	Aksi           string          `json:"aksi"`
	Modul          string          `json:"modul"`
	Method         string          `json:"method"`
	Endpoint       string          `json:"endpoint"`
	TabelTarget    string          `json:"tabel_target"`
	TargetID       string          `json:"target_id"`
	QueryParams    json.RawMessage `json:"query_params"`
	RequestData    json.RawMessage `json:"request_data"`
	DataSebelum    json.RawMessage `json:"data_sebelum"`
	DataSesudah    json.RawMessage `json:"data_sesudah"`
	ResponseStatus int             `json:"response_status"`
	Berhasil       bool            `json:"berhasil"`
	DurasiMS       uint64          `json:"durasi_ms"`
	PesanError     string          `json:"pesan_error"`
	IPAddress      string          `json:"ip_address"`
	UserAgent      string          `json:"user_agent"`
	CreatedAt      time.Time       `json:"created_at"`
}

type Filter struct {
	Halaman        int
	Batas          int
	KataKunci      string
	Username       string
	Aksi           string
	Modul          string
	Berhasil       string
	TanggalMulai   string
	TanggalSelesai string
}

type HasilDaftar struct {
	Data         []Catatan `json:"data"`
	Halaman      int       `json:"halaman"`
	Batas        int       `json:"batas"`
	Total        int64     `json:"total"`
	TotalHalaman int       `json:"total_halaman"`
}

type Repositori struct {
	db      *sql.DB
	simrsDB *sql.DB
}

func NewRepositori(db, simrsDB *sql.DB) *Repositori {
	return &Repositori{db: db, simrsDB: simrsDB}
}

func (r *Repositori) Simpan(ctx context.Context, catatan Catatan) error {
	if catatan.Username == "" {
		catatan.Username = "anonim"
	}

	if len(catatan.DataSebelum) == 0 && catatan.TargetID != "" && (catatan.Aksi == "UBAH" || catatan.Aksi == "HAPUS") {
		var sebelumnya []byte
		err := r.db.QueryRowContext(ctx, `
			SELECT data_sesudah
			FROM aktivitas_log
			WHERE modul = ? AND target_id = ? AND berhasil = TRUE AND data_sesudah IS NOT NULL
			ORDER BY id DESC
			LIMIT 1
		`, catatan.Modul, catatan.TargetID).Scan(&sebelumnya)
		if err == nil {
			catatan.DataSebelum = sebelumnya
		} else if err != sql.ErrNoRows {
			return fmt.Errorf("baca snapshot aktivitas sebelumnya: %w", err)
		}
	}
	if catatan.Aksi == "HAPUS" && len(catatan.DataSebelum) == 0 {
		catatan.DataSebelum = catatan.RequestData
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO aktivitas_log (
			user_id, request_id, username, aksi, modul, method, endpoint,
			tabel_target, target_id, query_params, request_data, data_sebelum,
			data_sesudah, response_status, berhasil, durasi_ms, pesan_error,
			ip_address, user_agent, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, nullableUint64(catatan.UserID), nullableString(catatan.RequestID), catatan.Username,
		catatan.Aksi, catatan.Modul, catatan.Method, catatan.Endpoint,
		nullableString(catatan.TabelTarget), nullableString(catatan.TargetID), nullableJSON(catatan.QueryParams),
		nullableJSON(catatan.RequestData), nullableJSON(catatan.DataSebelum), nullableJSON(catatan.DataSesudah),
		catatan.ResponseStatus, catatan.Berhasil, catatan.DurasiMS, nullableString(catatan.PesanError),
		nullableString(catatan.IPAddress), nullableString(catatan.UserAgent))
	if err != nil {
		return fmt.Errorf("simpan aktivitas log: %w", err)
	}
	return nil
}

func (r *Repositori) Daftar(ctx context.Context, filter Filter) (HasilDaftar, error) {
	if filter.Halaman < 1 {
		filter.Halaman = 1
	}
	if filter.Batas < 1 || filter.Batas > 100 {
		filter.Batas = 25
	}

	where, args := susunFilter(filter)
	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM aktivitas_log"+where, args...).Scan(&total); err != nil {
		return HasilDaftar{}, fmt.Errorf("hitung aktivitas log: %w", err)
	}

	queryArgs := append(append([]any{}, args...), filter.Batas, (filter.Halaman-1)*filter.Batas)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, COALESCE(request_id, ''), username, aksi, modul,
			COALESCE(method, ''), COALESCE(endpoint, ''), COALESCE(tabel_target, ''),
			COALESCE(target_id, ''), query_params, request_data, data_sebelum, data_sesudah,
			COALESCE(response_status, 0), berhasil, COALESCE(durasi_ms, 0),
			COALESCE(pesan_error, ''), COALESCE(ip_address, ''), COALESCE(user_agent, ''), created_at
		FROM aktivitas_log`+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return HasilDaftar{}, fmt.Errorf("baca aktivitas log: %w", err)
	}
	defer rows.Close()

	data := make([]Catatan, 0)
	for rows.Next() {
		var item Catatan
		var userID sql.NullInt64
		var queryParams, requestData, sebelum, sesudah []byte
		if err := rows.Scan(&item.ID, &userID, &item.RequestID, &item.Username, &item.Aksi, &item.Modul,
			&item.Method, &item.Endpoint, &item.TabelTarget, &item.TargetID, &queryParams, &requestData,
			&sebelum, &sesudah, &item.ResponseStatus, &item.Berhasil, &item.DurasiMS,
			&item.PesanError, &item.IPAddress, &item.UserAgent, &item.CreatedAt); err != nil {
			return HasilDaftar{}, fmt.Errorf("scan aktivitas log: %w", err)
		}
		if userID.Valid {
			id := uint64(userID.Int64)
			item.UserID = &id
		}
		item.QueryParams = queryParams
		item.RequestData = requestData
		item.DataSebelum = sebelum
		item.DataSesudah = sesudah
		data = append(data, item)
	}
	if err := rows.Err(); err != nil {
		return HasilDaftar{}, fmt.Errorf("iterasi aktivitas log: %w", err)
	}
	r.lengkapiNamaPegawai(ctx, data)

	totalHalaman := 0
	if total > 0 {
		totalHalaman = int((total + int64(filter.Batas) - 1) / int64(filter.Batas))
	}
	return HasilDaftar{Data: data, Halaman: filter.Halaman, Batas: filter.Batas, Total: total, TotalHalaman: totalHalaman}, nil
}

// lengkapiNamaPegawai membaca nama dari SIMRS Khanza tanpa mengubah data di sana.
// Jika koneksi SIMRS sedang tidak tersedia, daftar log tetap dapat ditampilkan
// menggunakan username/NIK yang sudah tersimpan di database aplikasi.
func (r *Repositori) lengkapiNamaPegawai(ctx context.Context, data []Catatan) {
	if r.simrsDB == nil || len(data) == 0 {
		return
	}

	usernameUnik := make([]string, 0, len(data))
	sudahAda := make(map[string]struct{}, len(data))
	for i := range data {
		username := strings.TrimSpace(data[i].Username)
		data[i].NamaUser = username
		if username == "" || username == "anonim" {
			continue
		}
		if _, ada := sudahAda[username]; ada {
			continue
		}
		sudahAda[username] = struct{}{}
		usernameUnik = append(usernameUnik, username)
	}
	if len(usernameUnik) == 0 {
		return
	}

	placeholder := strings.TrimSuffix(strings.Repeat("?,", len(usernameUnik)), ",")
	args := make([]any, len(usernameUnik))
	for i, username := range usernameUnik {
		args[i] = username
	}
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT nik, COALESCE(nama, '')
		FROM pegawai
		WHERE nik IN (`+placeholder+`)
	`, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	namaPegawai := make(map[string]string, len(usernameUnik))
	for rows.Next() {
		var nik, nama string
		if err := rows.Scan(&nik, &nama); err != nil {
			return
		}
		if nama = strings.TrimSpace(nama); nama != "" {
			namaPegawai[nik] = nama
		}
	}
	for i := range data {
		if nama := namaPegawai[data[i].Username]; nama != "" {
			data[i].NamaUser = nama
		}
	}
}

func (r *Repositori) BolehMelihat(ctx context.Context, userID uint64) (bool, error) {
	var jumlah int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM user_permissions
		INNER JOIN permissions ON permissions.id = user_permissions.permission_id
		WHERE user_permissions.user_id = ? AND permissions.code IN ('*', 'sistem.audit_log')
	`, userID).Scan(&jumlah)
	return jumlah > 0, err
}

func susunFilter(filter Filter) (string, []any) {
	kondisi := make([]string, 0, 7)
	args := make([]any, 0, 16)
	tambah := func(sqlKondisi string, nilai any) {
		kondisi = append(kondisi, sqlKondisi)
		args = append(args, nilai)
	}
	if filter.Username != "" {
		tambah("username LIKE ?", "%"+filter.Username+"%")
	}
	if filter.KataKunci != "" {
		pola := "%" + filter.KataKunci + "%"
		kondisi = append(kondisi, `(
			username LIKE ? OR modul LIKE ? OR COALESCE(endpoint, '') LIKE ? OR
			COALESCE(tabel_target, '') LIKE ? OR COALESCE(target_id, '') LIKE ? OR
			CAST(query_params AS CHAR) LIKE ? OR CAST(request_data AS CHAR) LIKE ? OR
			CAST(data_sebelum AS CHAR) LIKE ? OR CAST(data_sesudah AS CHAR) LIKE ? OR
			COALESCE(pesan_error, '') LIKE ?
		)`)
		for i := 0; i < 10; i++ {
			args = append(args, pola)
		}
	}
	if filter.Aksi != "" {
		tambah("aksi = ?", filter.Aksi)
	}
	if filter.Modul != "" {
		tambah("modul = ?", filter.Modul)
	}
	if filter.Berhasil == "1" || filter.Berhasil == "0" {
		tambah("berhasil = ?", filter.Berhasil == "1")
	}
	if filter.TanggalMulai != "" {
		tambah("created_at >= ?", filter.TanggalMulai+" 00:00:00")
	}
	if filter.TanggalSelesai != "" {
		tambah("created_at < DATE_ADD(?, INTERVAL 1 DAY)", filter.TanggalSelesai)
	}
	if len(kondisi) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(kondisi, " AND "), args
}

func nullableJSON(value json.RawMessage) any {
	if len(value) == 0 || string(value) == "null" || string(value) == "{}" {
		return nil
	}
	return []byte(value)
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableUint64(value *uint64) any {
	if value == nil {
		return nil
	}
	return *value
}
