package kardeks

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"
)

type InputCairan struct {
	NoRawat  string   `json:"no_rawat"`
	Mulai    string   `json:"waktu_mulai"`
	Selesai  string   `json:"waktu_selesai"`
	Jenis    string   `json:"jenis"`
	Kategori string   `json:"kategori"`
	Rincian  string   `json:"rincian"`
	Volume   *float64 `json:"volume_ml"`
	Catatan  string   `json:"catatan"`
}

func validasiCairan(in InputCairan) error {
	invalid := func(s string) error { return fmt.Errorf("%w: %s", ErrValidasi, s) }
	if strings.TrimSpace(in.NoRawat) == "" || len(in.NoRawat) > 17 {
		return invalid("nomor rawat tidak valid")
	}
	mulai, e1 := time.Parse("2006-01-02 15:04:05", in.Mulai)
	selesai, e2 := time.Parse("2006-01-02 15:04:05", in.Selesai)
	if e1 != nil || e2 != nil || !selesai.After(mulai) {
		return invalid("waktu selesai harus setelah waktu mulai, format tanggal/jam tidak valid")
	}
	if in.Volume == nil || math.IsNaN(*in.Volume) || math.IsInf(*in.Volume, 0) || *in.Volume < 0 || *in.Volume > 9999999999.99 || math.Abs(*in.Volume*100-math.Round(*in.Volume*100)) > 0.0001 {
		return invalid("volume wajib diisi, tidak negatif, maksimal dua angka desimal")
	}
	masuk := map[string]bool{"Infus": true, "Transfusi": true, "Oral": true, "NGT": true, "Lainnya": true}
	keluar := map[string]bool{"Urine": true, "Drain": true, "Cairan Lambung": true, "Lainnya": true}
	if !((in.Jenis == "Masuk" && masuk[in.Kategori]) || (in.Jenis == "Keluar" && keluar[in.Kategori])) {
		return invalid("kategori tidak sesuai jenis cairan")
	}
	if utf8.RuneCountInString(in.Rincian) > 100 || len(in.Catatan) > 16000 {
		return invalid("rincian maksimal 100 karakter dan catatan maksimal 16000 byte")
	}
	return nil
}

func (r *Repositori) SimpanCairan(ctx context.Context, in InputCairan, petugas string) error {
	if err := validasiCairan(in); err != nil {
		return err
	}
	tx, err := r.simrs.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var bayar string
	err = tx.QueryRowContext(ctx, "SELECT COALESCE(status_bayar,'') FROM reg_periksa WHERE no_rawat=? FOR UPDATE", in.NoRawat).Scan(&bayar)
	if err == sql.ErrNoRows {
		return fmt.Errorf("%w: kunjungan tidak ditemukan", ErrValidasi)
	}
	if err != nil {
		return err
	}
	if strings.EqualFold(bayar, "Sudah Bayar") {
		return fmt.Errorf("%w: billing sudah selesai; pencatatan terkunci", ErrValidasi)
	}
	var ada bool
	if err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM petugas WHERE nip=?)", petugas).Scan(&ada); err != nil {
		return err
	}
	if !ada || len(petugas) > 20 {
		return fmt.Errorf("%w: akun login belum terhubung dengan NIP petugas SIMRS", ErrValidasi)
	}
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM catatan_cairan_ranap WHERE no_rawat=? AND waktu_mulai=? AND waktu_selesai=? AND jenis=? AND kategori=? AND COALESCE(rincian,'')=? AND volume_ml=?)`, in.NoRawat, in.Mulai, in.Selesai, in.Jenis, in.Kategori, in.Rincian, *in.Volume).Scan(&ada); err != nil {
		return err
	}
	if ada {
		return fmt.Errorf("%w: catatan identik sudah tersimpan; muat ulang sebelum mencoba lagi", ErrValidasi)
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO catatan_cairan_ranap(no_rawat,waktu_mulai,waktu_selesai,jenis,kategori,rincian,volume_ml,petugas,catatan) VALUES(?,?,?,?,?,?,?,?,?)`, in.NoRawat, in.Mulai, in.Selesai, in.Jenis, in.Kategori, in.Rincian, *in.Volume, petugas, in.Catatan)
	if err != nil {
		return err
	}
	return tx.Commit()
}
