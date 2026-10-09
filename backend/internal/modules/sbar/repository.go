package sbar

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrValidasi = errors.New("data SBAR tidak valid")
	ErrAkses    = errors.New("akses tindakan SBAR tidak diizinkan")
	ErrKonflik  = errors.New("catatan berubah, sudah diverifikasi, atau tidak ditemukan; muat ulang riwayat")
)

type Kunci struct {
	Tanggal string `json:"tgl_perawatan"`
	Jam     string `json:"jam_rawat"`
}

type Data struct {
	Kunci
	NIP            string `json:"nip"`
	Situation      string `json:"situation"`
	Background     string `json:"background"`
	Assesment      string `json:"assesment"`
	Recommendation string `json:"recommendation"`
	Instruksi      string `json:"instruksi"`
}

type Input struct {
	NoRawat   string `json:"no_rawat"`
	Data      Data   `json:"data"`
	Asli      Kunci  `json:"asli"`
	Revisi    string `json:"revisi"`
	Validator string `json:"validator"`
}

type Pilihan struct {
	Kode    string `json:"kode"`
	Nama    string `json:"nama"`
	Jabatan string `json:"jabatan"`
}

type Catatan struct {
	Data            Data   `json:"data"`
	NamaPetugas     string `json:"nama_petugas"`
	Jabatan         string `json:"jabatan"`
	Status          string `json:"status"`
	Validator       string `json:"validator"`
	NamaValidator   string `json:"nama_validator"`
	TanggalValidasi string `json:"tanggal_validasi"`
	JamValidasi     string `json:"jam_validasi"`
	Revisi          string `json:"revisi"`
	BisaUbah        bool   `json:"bisa_ubah"`
	BisaVerifikasi  bool   `json:"bisa_verifikasi"`
	Terkunci        bool   `json:"terkunci"`
	InstruksiAda    bool   `json:"-"`
	ValidasiAda     bool   `json:"-"`
	WaktuInstruksi  string `json:"-"`
}

type Hasil struct {
	Catatan        []Catatan `json:"catatan"`
	Petugas        Pilihan   `json:"petugas_login"`
	Dokter         Pilihan   `json:"dokter_login"`
	DPJP           []Pilihan `json:"dpjp"`
	Admin          bool      `json:"boleh_pilih_petugas"`
	BatasInstruksi int       `json:"batas_instruksi"`
}

type Repositori struct{ db *sql.DB }

func NewRepositori(db *sql.DB) *Repositori { return &Repositori{db: db} }

func (r *Repositori) batasInstruksi(ctx context.Context) (int, error) {
	var panjang int64
	err := r.db.QueryRowContext(ctx, `SELECT CHARACTER_MAXIMUM_LENGTH FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='sbar_instruksi' AND COLUMN_NAME='instruksi'`).Scan(&panjang)
	if err != nil {
		return 0, err
	}
	if panjang < 1 {
		return 0, fmt.Errorf("struktur instruksi tidak sesuai")
	}
	if panjang > 2000 {
		panjang = 2000
	}
	return int(panjang), nil
}

const pilihCatatan = `SELECT DATE_FORMAT(s.tgl_perawatan,'%Y-%m-%d'), TIME_FORMAT(s.jam_rawat,'%H:%i:%s'),
	s.nip, COALESCE(s.situation,''), COALESCE(s.background,''), s.assesment, s.recommendation,
	COALESCE(i.instruksi,''), COALESCE(p.nama,s.nip), COALESCE(p.jbtn,''),
	COALESCE(v.status_validasi,''), COALESCE(v.nik_validator,''), COALESCE(d.nm_dokter,v.nik_validator,''),
	COALESCE(DATE_FORMAT(v.tgl_validasi,'%Y-%m-%d'),''), COALESCE(TIME_FORMAT(v.jam_validasi,'%H:%i:%s'),''),
	i.no_rawat IS NOT NULL, v.no_rawat IS NOT NULL,
	COALESCE(DATE_FORMAT(i.tgl_validasi,'%Y-%m-%d'),'')
	FROM pemeriksaan_ranap_sbar s
	LEFT JOIN sbar_instruksi i ON i.no_rawat=s.no_rawat AND i.tgl_perawatan=s.tgl_perawatan
		AND i.jam_rawat=s.jam_rawat AND i.nip=s.nip
	LEFT JOIN pegawai p ON p.nik=s.nip
	LEFT JOIN validasi_pemeriksaan_sbar v ON v.no_rawat=s.no_rawat
		AND v.tgl_perawatan=s.tgl_perawatan AND v.jam_rawat=s.jam_rawat
	LEFT JOIN dokter d ON d.kd_dokter=v.nik_validator
	WHERE s.no_rawat=?`

func bacaCatatan(rows *sql.Rows) ([]Catatan, error) {
	defer rows.Close()
	hasil := []Catatan{}
	for rows.Next() {
		var c Catatan
		err := rows.Scan(&c.Data.Tanggal, &c.Data.Jam, &c.Data.NIP, &c.Data.Situation,
			&c.Data.Background, &c.Data.Assesment, &c.Data.Recommendation, &c.Data.Instruksi,
			&c.NamaPetugas, &c.Jabatan, &c.Status, &c.Validator, &c.NamaValidator,
			&c.TanggalValidasi, &c.JamValidasi, &c.InstruksiAda, &c.ValidasiAda, &c.WaktuInstruksi)
		if err != nil {
			return nil, err
		}
		snapshot, err := json.Marshal([]any{c.Data, c.Status, c.Validator, c.TanggalValidasi,
			c.JamValidasi, c.InstruksiAda, c.ValidasiAda, c.WaktuInstruksi})
		if err != nil {
			return nil, err
		}
		hash := sha256.Sum256(snapshot)
		c.Revisi = hex.EncodeToString(hash[:])
		hasil = append(hasil, c)
	}
	return hasil, rows.Err()
}

func terkunci(c Catatan) bool {
	return c.ValidasiAda || (c.WaktuInstruksi != "" && c.WaktuInstruksi != "0000-00-00")
}

func validasiKunci(k Kunci) bool {
	tanggal, err := time.Parse("2006-01-02", k.Tanggal)
	if err != nil || tanggal.Year() < 1 {
		return false
	}
	_, err = time.Parse("15:04:05", k.Jam)
	return err == nil
}

func Validasi(in Input, aksi string) error {
	if strings.TrimSpace(in.NoRawat) == "" || len(in.NoRawat) > 17 {
		return ErrValidasi
	}
	if aksi != "POST" && aksi != "PUT" && aksi != "DELETE" && aksi != "verifikasi" {
		return ErrValidasi
	}
	if aksi != "POST" && (!validasiKunci(in.Asli) || len(in.Revisi) != 64) {
		return fmt.Errorf("%w: pilih catatan dari riwayat terbaru", ErrValidasi)
	}
	if aksi == "DELETE" || aksi == "verifikasi" {
		return nil
	}
	if !validasiKunci(in.Data.Kunci) || in.Data.NIP == "" || utf8.RuneCountInString(in.Data.NIP) > 20 {
		return fmt.Errorf("%w: tanggal, jam, dan petugas wajib valid", ErrValidasi)
	}
	isi := false
	for _, teks := range []string{in.Data.Situation, in.Data.Background, in.Data.Assesment, in.Data.Recommendation} {
		if utf8.RuneCountInString(teks) > 2000 {
			return fmt.Errorf("%w: setiap kolom SBAR maksimal 2000 karakter", ErrValidasi)
		}
		isi = isi || strings.TrimSpace(teks) != ""
	}
	if !isi {
		return fmt.Errorf("%w: isi minimal salah satu S/B/A/R", ErrValidasi)
	}
	if utf8.RuneCountInString(in.Data.Instruksi) > 2000 {
		return fmt.Errorf("%w: instruksi maksimal 2000 karakter", ErrValidasi)
	}
	return nil
}

func (r *Repositori) Referensi(ctx context.Context, jenis, q string) ([]Pilihan, error) {
	query := `SELECT nik,nama,COALESCE(jbtn,'') FROM pegawai WHERE nik LIKE ? OR nama LIKE ? ORDER BY nama LIMIT 30`
	if jenis == "dokter" {
		query = `SELECT kd_dokter,nm_dokter,'' FROM dokter WHERE kd_dokter LIKE ? OR nm_dokter LIKE ? ORDER BY nm_dokter LIMIT 30`
	} else if jenis != "petugas" {
		return nil, ErrValidasi
	}
	rows, err := r.db.QueryContext(ctx, query, "%"+q+"%", "%"+q+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Pilihan{}
	for rows.Next() {
		var p Pilihan
		if err := rows.Scan(&p.Kode, &p.Nama, &p.Jabatan); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repositori) Daftar(ctx context.Context, no, user string, admin bool) (Hasil, error) {
	h := Hasil{Catatan: []Catatan{}, DPJP: []Pilihan{}, Admin: admin}
	if strings.TrimSpace(no) == "" || len(no) > 17 {
		return h, ErrValidasi
	}
	batas, err := r.batasInstruksi(ctx)
	if err != nil {
		return h, err
	}
	h.BatasInstruksi = batas
	err = r.db.QueryRowContext(ctx, `SELECT nik,nama,COALESCE(jbtn,'') FROM pegawai WHERE nik=?`, user).Scan(&h.Petugas.Kode, &h.Petugas.Nama, &h.Petugas.Jabatan)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return h, err
	}
	err = r.db.QueryRowContext(ctx, `SELECT kd_dokter,nm_dokter FROM dokter WHERE kd_dokter=?`, user).Scan(&h.Dokter.Kode, &h.Dokter.Nama)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return h, err
	}
	dpjp, err := r.db.QueryContext(ctx, `SELECT dp.kd_dokter,COALESCE(d.nm_dokter,dp.kd_dokter) FROM dpjp_ranap dp LEFT JOIN dokter d ON d.kd_dokter=dp.kd_dokter WHERE dp.no_rawat=? ORDER BY d.nm_dokter`, no)
	if err != nil {
		return h, err
	}
	for dpjp.Next() {
		var p Pilihan
		if err = dpjp.Scan(&p.Kode, &p.Nama); err != nil {
			dpjp.Close()
			return h, err
		}
		h.DPJP = append(h.DPJP, p)
	}
	err = dpjp.Err()
	dpjp.Close()
	if err != nil {
		return h, err
	}
	rows, err := r.db.QueryContext(ctx, pilihCatatan+` ORDER BY s.tgl_perawatan DESC,s.jam_rawat DESC`, no)
	if err != nil {
		return h, err
	}
	h.Catatan, err = bacaCatatan(rows)
	for i := range h.Catatan {
		c := &h.Catatan[i]
		c.Terkunci = terkunci(*c)
		c.BisaUbah = !terkunci(*c) && (admin || c.Data.NIP == user)
		c.BisaVerifikasi = !terkunci(*c) && (admin || h.Dokter.Kode != "")
	}
	return h, err
}

func (r *Repositori) Mutasi(ctx context.Context, in Input, aksi, user string, admin bool) error {
	if err := Validasi(in, aksi); err != nil {
		return err
	}
	var tabelTransaksional int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.TABLES
		WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME IN ('pemeriksaan_ranap_sbar','sbar_instruksi','validasi_pemeriksaan_sbar')
		AND ENGINE='InnoDB'`).Scan(&tabelTransaksional)
	if err != nil {
		return err
	}
	if tabelTransaksional != 3 {
		return fmt.Errorf("%w: ketiga tabel SBAR harus tersedia dan memakai InnoDB agar simpan/verifikasi aman", ErrValidasi)
	}
	if aksi == "POST" || aksi == "PUT" {
		batas, e := r.batasInstruksi(ctx)
		if e != nil {
			return e
		}
		if utf8.RuneCountInString(in.Data.Instruksi) > batas {
			return fmt.Errorf("%w: instruksi maksimal %d karakter sesuai kolom SIMRS", ErrValidasi, batas)
		}
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Seluruh mutasi untuk kunjungan yang sama diserialkan, lalu snapshot dicek.
	var layanan string
	err = tx.QueryRowContext(ctx, `SELECT status_lanjut FROM reg_periksa WHERE no_rawat=? FOR UPDATE`, in.NoRawat).Scan(&layanan)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrKonflik
	}
	if err != nil {
		return err
	}
	if layanan != "Ranap" {
		return fmt.Errorf("%w: SBAR ini khusus kunjungan rawat inap", ErrValidasi)
	}
	var lama Catatan
	if aksi != "POST" {
		rows, e := tx.QueryContext(ctx, pilihCatatan+` AND s.tgl_perawatan=? AND s.jam_rawat=? FOR UPDATE`, in.NoRawat, in.Asli.Tanggal, in.Asli.Jam)
		if e != nil {
			return e
		}
		daftar, e := bacaCatatan(rows)
		if e != nil {
			return e
		}
		if len(daftar) != 1 || daftar[0].Revisi != in.Revisi || terkunci(daftar[0]) {
			return ErrKonflik
		}
		lama = daftar[0]
		if aksi != "verifikasi" && !admin && lama.Data.NIP != user {
			return ErrAkses
		}
		if aksi == "PUT" && (in.Data.Kunci != lama.Data.Kunci || in.Data.NIP != lama.Data.NIP) {
			return fmt.Errorf("%w: waktu dan pencatat merupakan identitas catatan dan tidak dapat diganti", ErrValidasi)
		}
	}
	if aksi == "POST" {
		var ada bool
		err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM validasi_pemeriksaan_sbar WHERE no_rawat=? AND tgl_perawatan=? AND jam_rawat=?)
			OR EXISTS(SELECT 1 FROM sbar_instruksi WHERE no_rawat=? AND tgl_perawatan=? AND jam_rawat=?)`,
			in.NoRawat, in.Data.Tanggal, in.Data.Jam, in.NoRawat, in.Data.Tanggal, in.Data.Jam).Scan(&ada)
		if err != nil {
			return err
		}
		if ada {
			return ErrKonflik
		}
	}
	if aksi == "verifikasi" {
		if !admin && in.Validator != user {
			return ErrAkses
		}
		var dokter string
		err = tx.QueryRowContext(ctx, `SELECT kd_dokter FROM dokter WHERE kd_dokter=?`, in.Validator).Scan(&dokter)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAkses
		}
		if err != nil {
			return err
		}
		if strings.TrimSpace(lama.Data.Situation) == "" || strings.TrimSpace(lama.Data.Background) == "" {
			return fmt.Errorf("%w: Situation dan Background wajib terisi sebelum verifikasi", ErrValidasi)
		}
		waktu := time.Now().In(time.FixedZone("WITA", 8*3600))
		tgl, jam := waktu.Format("2006-01-02"), waktu.Format("15:04:05")
		d := lama.Data
		_, err = tx.ExecContext(ctx, `INSERT INTO validasi_pemeriksaan_sbar
			(no_rawat,tgl_perawatan,jam_rawat,situation,background,assesment,recommendation,nik,nik_validator,tgl_validasi,jam_validasi,status_validasi)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,'Validasi')`, in.NoRawat, d.Tanggal, d.Jam, d.Situation, d.Background, d.Assesment, d.Recommendation, d.NIP, dokter, tgl, jam)
		if err != nil {
			return err
		}
		if lama.InstruksiAda {
			_, err = tx.ExecContext(ctx, `UPDATE sbar_instruksi SET tgl_validasi=?,jam_validasi=? WHERE no_rawat=? AND tgl_perawatan=? AND jam_rawat=? AND nip=?`, tgl, jam, in.NoRawat, d.Tanggal, d.Jam, d.NIP)
		} else {
			_, err = tx.ExecContext(ctx, `INSERT INTO sbar_instruksi (no_rawat,tgl_perawatan,jam_rawat,instruksi,nip,tgl_validasi,jam_validasi) VALUES(?,?,?,?,?,?,?)`, in.NoRawat, d.Tanggal, d.Jam, d.Instruksi, d.NIP, tgl, jam)
		}
	} else if aksi == "DELETE" {
		_, err = tx.ExecContext(ctx, `DELETE FROM sbar_instruksi WHERE no_rawat=? AND tgl_perawatan=? AND jam_rawat=? AND nip=?`, in.NoRawat, lama.Data.Tanggal, lama.Data.Jam, lama.Data.NIP)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `DELETE FROM pemeriksaan_ranap_sbar WHERE no_rawat=? AND tgl_perawatan=? AND jam_rawat=?`, in.NoRawat, lama.Data.Tanggal, lama.Data.Jam)
	} else {
		d := in.Data
		if !admin && d.NIP != user {
			return ErrAkses
		}
		var pegawai string
		err = tx.QueryRowContext(ctx, `SELECT nik FROM pegawai WHERE nik=?`, d.NIP).Scan(&pegawai)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: petugas tidak ditemukan", ErrValidasi)
		}
		if err != nil {
			return err
		}
		if aksi == "POST" {
			_, err = tx.ExecContext(ctx, `INSERT INTO pemeriksaan_ranap_sbar
				(no_rawat,tgl_perawatan,jam_rawat,situation,background,assesment,recommendation,nip) VALUES(?,?,?,?,?,?,?,?)`,
				in.NoRawat, d.Tanggal, d.Jam, d.Situation, d.Background, d.Assesment, d.Recommendation, d.NIP)
		} else {
			_, err = tx.ExecContext(ctx, `UPDATE pemeriksaan_ranap_sbar SET tgl_perawatan=?,jam_rawat=?,situation=?,background=?,assesment=?,recommendation=?,nip=? WHERE no_rawat=? AND tgl_perawatan=? AND jam_rawat=?`,
				d.Tanggal, d.Jam, d.Situation, d.Background, d.Assesment, d.Recommendation, d.NIP, in.NoRawat, lama.Data.Tanggal, lama.Data.Jam)
		}
		if err != nil {
			return err
		}
		if aksi == "PUT" && lama.InstruksiAda {
			_, err = tx.ExecContext(ctx, `UPDATE sbar_instruksi SET tgl_perawatan=?,jam_rawat=?,instruksi=?,nip=? WHERE no_rawat=? AND tgl_perawatan=? AND jam_rawat=? AND nip=?`,
				d.Tanggal, d.Jam, d.Instruksi, d.NIP, in.NoRawat, lama.Data.Tanggal, lama.Data.Jam, lama.Data.NIP)
		} else {
			_, err = tx.ExecContext(ctx, `INSERT INTO sbar_instruksi (no_rawat,tgl_perawatan,jam_rawat,instruksi,nip,tgl_validasi,jam_validasi) VALUES(?,?,?,?,?,NULL,NULL)`, in.NoRawat, d.Tanggal, d.Jam, d.Instruksi, d.NIP)
		}
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}
