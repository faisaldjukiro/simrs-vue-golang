package rujukan_internal

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrAksesKonsul = errors.New("konsultasi hanya dapat diubah oleh dokter perujuk dan dijawab oleh dokter tujuan atau administrator")

var TujuanKonsul = []string{
	"Konsultasi/tindakan medik saat ini",
	"Perawatan bersama untuk selanjutnya",
	"Alih rawat kasus ini untuk selanjutnya",
}

type SuratKonsul struct {
	NoRawat          string `json:"no_rawat"`
	NoSurat          string `json:"no_surat"`
	Tanggal          string `json:"tanggalsurat"`
	DokterAsal       string `json:"kd_dokter_perujuk"`
	NamaDokterAsal   string `json:"dokterperujuk"`
	PoliAsal         string `json:"kd_poli_perujuk"`
	NamaPoliAsal     string `json:"poliperujuk"`
	DokterTujuan     string `json:"kd_dokter_tujuan"`
	NamaDokterTujuan string `json:"doktertujuan"`
	PoliTujuan       string `json:"kd_poli_tujuan"`
	NamaPoliTujuan   string `json:"politujuan"`
	Umur             string `json:"umur"`
	Tujuan           string `json:"tujuan_konsul"`
	Status           string `json:"status"`
	Isi              string `json:"isikonsul"`
	TanggalJawaban   string `json:"tanggaljawaban"`
	Jawaban          string `json:"jawabankonsul"`
}

type BarisKonsul struct {
	SuratKonsul
	Revisi    string `json:"revisi"`
	BisaUbah  bool   `json:"bisa_ubah"`
	BisaJawab bool   `json:"bisa_jawab"`
}

type InputKonsul struct {
	SuratKonsul
	Revisi string `json:"revisi"`
}

type DataKonsul struct {
	Daftar      []BarisKonsul `json:"daftar"`
	DokterLogin Referensi     `json:"dokter_login"`
	Asal        Referensi     `json:"asal"`
	Admin       bool          `json:"admin"`
	BolehSimpan bool          `json:"boleh_simpan"`
	Pesan       string        `json:"pesan"`
}

const kolomKonsul = `COALESCE(no_rawat,''),no_surat,DATE_FORMAT(tanggalsurat,'%Y-%m-%d %H:%i:%s'),
kd_dokter_perujuk,dokterperujuk,kd_poli_perujuk,poliperujuk,kd_dokter_tujuan,doktertujuan,
kd_poli_tujuan,politujuan,umur,tujuan_konsul,status,isikonsul,
COALESCE(DATE_FORMAT(tanggaljawaban,'%Y-%m-%d %H:%i:%s'),''),jawabankonsul`

type pemindaiKonsul interface{ Scan(...any) error }

func pindaiKonsul(row pemindaiKonsul, s *SuratKonsul) error {
	return row.Scan(&s.NoRawat, &s.NoSurat, &s.Tanggal, &s.DokterAsal, &s.NamaDokterAsal,
		&s.PoliAsal, &s.NamaPoliAsal, &s.DokterTujuan, &s.NamaDokterTujuan, &s.PoliTujuan,
		&s.NamaPoliTujuan, &s.Umur, &s.Tujuan, &s.Status, &s.Isi, &s.TanggalJawaban, &s.Jawaban)
}

func revisiKonsul(s SuratKonsul) string {
	b, _ := json.Marshal(s)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

func konsulSelesai(s SuratKonsul) bool {
	return s.Status != "Menunggu" || strings.TrimSpace(s.Jawaban) != "" ||
		(s.TanggalJawaban != "" && s.TanggalJawaban != "0000-00-00 00:00:00")
}

func dokterKonsul(ctx context.Context, db pembaca, username string) (Referensi, error) {
	var d Referensi
	err := db.QueryRowContext(ctx, `SELECT kd_dokter,nm_dokter FROM dokter WHERE kd_dokter=? AND status='1'`, username).Scan(&d.Kode, &d.Nama)
	if errors.Is(err, sql.ErrNoRows) {
		return d, nil
	}
	return d, err
}

func (r *Repositori) DaftarKonsul(ctx context.Context, jenis, noRawat, username string, admin bool) (DataKonsul, error) {
	d := DataKonsul{Daftar: []BarisKonsul{}, Admin: admin}
	if !validJenis(jenis) || !validNoRawat(noRawat) {
		return d, ErrInput
	}
	err := periksaKunjungan(ctx, r.simrsDB, noRawat, jenis, false)
	if err != nil && !errors.Is(err, ErrBilling) && !errors.Is(err, ErrKunjungan) {
		return d, err
	}
	if err != nil {
		d.Pesan = err.Error()
	}
	d.DokterLogin, err = dokterKonsul(ctx, r.simrsDB, username)
	if err != nil {
		return d, err
	}
	d.BolehSimpan = d.Pesan == "" && (admin || d.DokterLogin.Kode != "")
	if d.Pesan == "" && !d.BolehSimpan {
		d.Pesan = "Akun belum terhubung dengan dokter aktif. Riwayat hanya dapat dilihat."
	}
	query := `SELECT p.kd_poli,p.nm_poli FROM reg_periksa r JOIN poliklinik p ON p.kd_poli=r.kd_poli WHERE r.no_rawat=?`
	if jenis == Ranap {
		query = `SELECT k.kd_kamar,b.nm_bangsal FROM kamar_inap ki JOIN kamar k ON k.kd_kamar=ki.kd_kamar JOIN bangsal b ON b.kd_bangsal=k.kd_bangsal WHERE ki.no_rawat=? ORDER BY ki.tgl_masuk DESC,ki.jam_masuk DESC LIMIT 1`
	}
	err = r.simrsDB.QueryRowContext(ctx, query, noRawat).Scan(&d.Asal.Kode, &d.Asal.Nama)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return d, err
	}
	rows, err := r.simrsDB.QueryContext(ctx, "SELECT "+kolomKonsul+" FROM surat_konsul WHERE no_rawat=? ORDER BY tanggalsurat DESC,no_surat DESC", noRawat)
	if err != nil {
		return d, err
	}
	defer rows.Close()
	for rows.Next() {
		var b BarisKonsul
		if err = pindaiKonsul(rows, &b.SuratKonsul); err != nil {
			return d, err
		}
		b.Revisi = revisiKonsul(b.SuratKonsul)
		b.BisaUbah = d.BolehSimpan && !konsulSelesai(b.SuratKonsul) && (admin || d.DokterLogin.Kode == b.DokterAsal)
		b.BisaJawab = d.BolehSimpan && !konsulSelesai(b.SuratKonsul) && (admin || d.DokterLogin.Kode == b.DokterTujuan)
		d.Daftar = append(d.Daftar, b)
	}
	return d, rows.Err()
}

func (r *Repositori) ReferensiAsalKonsul(ctx context.Context, jenis, q string) ([]Referensi, error) {
	if jenis != Ranap {
		return r.CariReferensi(ctx, "poli", q)
	}
	d := []Referensi{}
	if len(q) < 2 {
		return d, nil
	}
	if len(q) > 100 {
		return nil, ErrInput
	}
	rows, err := r.simrsDB.QueryContext(ctx, `SELECT k.kd_kamar,CONCAT(b.nm_bangsal,' · ',k.kd_kamar) FROM kamar k JOIN bangsal b ON b.kd_bangsal=k.kd_bangsal WHERE k.statusdata='1' AND b.status='1' AND (k.kd_kamar LIKE ? OR b.nm_bangsal LIKE ?) ORDER BY b.nm_bangsal,k.kd_kamar LIMIT 30`, "%"+q+"%", "%"+q+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var v Referensi
		if err = rows.Scan(&v.Kode, &v.Nama); err != nil {
			return nil, err
		}
		d = append(d, v)
	}
	return d, rows.Err()
}

func validTeksKonsul(label, teks string, batas int, wajib bool) error {
	if (wajib && strings.TrimSpace(teks) == "") || utf8.RuneCountInString(teks) > batas {
		return fmt.Errorf("%w: %s wajib sesuai batas %d karakter", ErrInput, label, batas)
	}
	return nil
}

func validIsiKonsul(isi, tujuan string) error {
	if err := validTeksKonsul("isi konsultasi", isi, 10000, true); err != nil {
		return err
	}
	for _, t := range TujuanKonsul {
		if tujuan == t {
			return nil
		}
	}
	return fmt.Errorf("%w: tujuan konsultasi tidak tersedia", ErrInput)
}

// Seluruh mutasi memakai SIMRS; tidak membuat tabel maupun salinan transaksi lokal.
func (r *Repositori) SimpanKonsul(ctx context.Context, jenis, username string, admin bool, aksi string, in InputKonsul) (string, error) {
	if !validJenis(jenis) || !validNoRawat(in.NoRawat) {
		return "", ErrInput
	}
	if aksi != "baru" && aksi != "ubah" && aksi != "jawab" {
		return "", ErrInput
	}
	if aksi == "jawab" {
		if err := validTeksKonsul("jawaban konsultasi", in.Jawaban, 10000, true); err != nil {
			return "", err
		}
	} else if err := validIsiKonsul(in.Isi, in.Tujuan); err != nil {
		return "", err
	}
	// Penomoran harian dan INSERT rujukan menggunakan koneksi serta transaksi yang sama.
	conn, err := r.simrsDB.Conn(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	if aksi == "baru" {
		var dapat int
		if err = conn.QueryRowContext(ctx, `SELECT GET_LOCK('sirapi_surat_konsul_nomor',5)`).Scan(&dapat); err != nil {
			return "", err
		}
		if dapat != 1 {
			return "", ErrBerubah
		}
		defer func() {
			cleanup, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			conn.ExecContext(cleanup, `DO RELEASE_LOCK('sirapi_surat_konsul_nomor')`)
		}()
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	// Hindari sukses parsial bila instalasi lama menggunakan engine non-transaksional.
	refTabel, _ := targetKhanza(jenis)
	for _, tabel := range []string{"surat_konsul", "reg_periksa", refTabel} {
		var engine string
		if err = tx.QueryRowContext(ctx, `SELECT ENGINE FROM information_schema.TABLES WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=?`, tabel).Scan(&engine); err != nil {
			return "", err
		}
		if !strings.EqualFold(engine, "InnoDB") {
			return "", fmt.Errorf("%w: tabel %s harus mendukung transaksi InnoDB", ErrInput, tabel)
		}
	}
	if err = periksaKunjungan(ctx, tx, in.NoRawat, jenis, true); err != nil {
		return "", err
	}
	dokter, err := dokterKonsul(ctx, tx, username)
	if err != nil {
		return "", err
	}
	if !admin && dokter.Kode == "" {
		return "", ErrAksesKonsul
	}
	if aksi != "baru" {
		var aktual SuratKonsul
		err = pindaiKonsul(tx.QueryRowContext(ctx, "SELECT "+kolomKonsul+" FROM surat_konsul WHERE no_rawat=? AND no_surat=? FOR UPDATE", in.NoRawat, in.NoSurat), &aktual)
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrBerubah
		}
		if err != nil {
			return "", err
		}
		if konsulSelesai(aktual) {
			return "", fmt.Errorf("%w: surat konsultasi sudah dijawab atau ditutup dan tidak dapat diubah", ErrBerubah)
		}
		if in.Revisi != revisiKonsul(aktual) {
			return "", ErrBerubah
		}
		if aksi == "jawab" {
			if !admin && dokter.Kode != aktual.DokterTujuan {
				return "", ErrAksesKonsul
			}
			waktu := time.Now().In(time.FixedZone("WITA", 8*3600)).Format("2006-01-02 15:04:05")
			// Tidak menerima perubahan isi/perujuk/tujuan dari payload jawaban.
			_, err = tx.ExecContext(ctx, `UPDATE surat_konsul SET jawabankonsul=?,tanggaljawaban=?,status='Selesai' WHERE no_rawat=? AND no_surat=?`, in.Jawaban, waktu, in.NoRawat, in.NoSurat)
		} else {
			if !admin && dokter.Kode != aktual.DokterAsal {
				return "", ErrAksesKonsul
			}
			// Tujuan rujukan tetap; koreksi hanya isi serta tujuan klinis surat sebelum dijawab.
			_, err = tx.ExecContext(ctx, `UPDATE surat_konsul SET isikonsul=?,tujuan_konsul=? WHERE no_rawat=? AND no_surat=?`, in.Isi, in.Tujuan, in.NoRawat, in.NoSurat)
		}
		if err != nil {
			return "", err
		}
		return in.NoSurat, tx.Commit()
	}
	if !admin && in.DokterAsal != dokter.Kode {
		return "", ErrAksesKonsul
	}
	waktu, err := time.Parse("2006-01-02 15:04:05", in.Tanggal)
	if err != nil || waktu.Year() < 1900 {
		return "", fmt.Errorf("%w: tanggal/jam surat tidak valid", ErrInput)
	}
	for label, v := range map[string]string{"kode dokter perujuk": in.DokterAsal, "kode dokter tujuan": in.DokterTujuan, "kode asal": in.PoliAsal, "kode poli tujuan": in.PoliTujuan} {
		if err = validTeksKonsul(label, v, 10, true); err != nil {
			return "", err
		}
	}
	d, err := dokterKonsul(ctx, tx, in.DokterAsal)
	if err != nil {
		return "", err
	}
	if d.Kode == "" {
		return "", fmt.Errorf("%w: dokter perujuk tidak aktif", ErrInput)
	}
	in.NamaDokterAsal = d.Nama
	ref, err := Validasi(jenis, Input{NoRawat: in.NoRawat, KodeDokter: in.DokterTujuan, KodePoli: in.PoliTujuan, Tanggal: in.Tanggal[:10], Jam: in.Tanggal[11:]})
	if err != nil {
		return "", err
	}
	in.NamaDokterTujuan, in.NamaPoliTujuan, err = periksaReferensi(ctx, tx, ref)
	if err != nil {
		return "", err
	}
	query := `SELECT nm_poli FROM poliklinik WHERE kd_poli=? AND status='1'`
	if jenis == Ranap {
		query = `SELECT b.nm_bangsal FROM kamar k JOIN bangsal b ON b.kd_bangsal=k.kd_bangsal WHERE k.kd_kamar=? AND k.statusdata='1' AND b.status='1'`
	}
	err = tx.QueryRowContext(ctx, query, in.PoliAsal).Scan(&in.NamaPoliAsal)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w: ruangan/poli asal tidak ditemukan", ErrInput)
	}
	if err != nil {
		return "", err
	}
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(p.umur,'') FROM reg_periksa r JOIN pasien p ON p.no_rkm_medis=r.no_rkm_medis WHERE r.no_rawat=?`, in.NoRawat).Scan(&in.Umur)
	if err != nil {
		return "", err
	}
	for _, v := range []struct {
		label, teks string
		batas       int
	}{{"nama dokter perujuk", in.NamaDokterAsal, 50}, {"nama ruangan/poli asal", in.NamaPoliAsal, 50}, {"nama dokter tujuan", in.NamaDokterTujuan, 100}, {"nama poli tujuan", in.NamaPoliTujuan, 100}, {"umur", in.Umur, 100}} {
		if err = validTeksKonsul(v.label, v.teks, v.batas, false); err != nil {
			return "", err
		}
	}
	var urut int
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(CAST(RIGHT(no_surat,3) AS UNSIGNED)),0) FROM surat_konsul WHERE tanggalsurat>=? AND tanggalsurat<DATE_ADD(?,INTERVAL 1 DAY)`, in.Tanggal[:10], in.Tanggal[:10]).Scan(&urut)
	if err != nil {
		return "", err
	}
	if urut >= 999 {
		return "", fmt.Errorf("%w: nomor konsultasi harian sudah mencapai 999", ErrInput)
	}
	in.NoSurat = fmt.Sprintf("SK%s%03d", waktu.Format("20060102"), urut+1)
	// Rujukan identik digunakan kembali; konflik tidak boleh menimpa rujukan lama.
	aktual, err := bacaAsal(ctx, tx, jenis, Rujukan{Input: ref, Sumber: "Khanza"})
	if errors.Is(err, sql.ErrNoRows) {
		if err = insertKhanza(ctx, tx, jenis, ref); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	} else if aktual != ref {
		return "", ErrDuplikat
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO surat_konsul
	(no_rawat,no_surat,tanggalsurat,kd_dokter_perujuk,dokterperujuk,kd_poli_perujuk,poliperujuk,kd_dokter_tujuan,doktertujuan,kd_poli_tujuan,politujuan,umur,tujuan_konsul,status,isikonsul,tanggaljawaban,jawabankonsul)
	VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,'Menunggu',?,'0000-00-00 00:00:00','')`,
		in.NoRawat, in.NoSurat, in.Tanggal, in.DokterAsal, in.NamaDokterAsal, in.PoliAsal, in.NamaPoliAsal, in.DokterTujuan, in.NamaDokterTujuan, in.PoliTujuan, in.NamaPoliTujuan, in.Umur, in.Tujuan, in.Isi)
	if err != nil {
		return "", errorSQL(err)
	}
	return in.NoSurat, tx.Commit()
}
