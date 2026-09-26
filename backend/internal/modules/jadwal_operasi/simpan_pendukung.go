package jadwal_operasi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
	"simrs-backend/internal/shared/khanzamutasi"
)

type InputPendukung struct {
	Jenis   string            `json:"jenis"`
	NoRawat string            `json:"no_rawat"`
	Jadwal  Input             `json:"jadwal"`
	Data    map[string]string `json:"data"`
	Asli    map[string]string `json:"asli"`
}

var ErrAksesPendukung = errors.New("hanya tenaga medis/petugas yang tercatat atau administrator yang boleh mengubah catatan ini")

func BolehUbahPendukung(jenis, username string, admin bool, data map[string]string) bool {
	if admin {
		return true
	}
	if username == "" {
		return false
	}
	var kode []string
	switch jenis {
	case "signin_sebelum_anestesi", "timeout_sebelum_insisi":
		kode = []string{"kd_dokter_anestesi", "nip_perawat_ok"}
	case "signout_sebelum_menutup_luka":
		kode = []string{"kd_dokter_bedah", "kd_dokter_anestesi", "nip_perawat_ok"}
	case "checklist_post_operasi":
		kode = []string{"nip_perawat_ok", "nip_perawat_anestesi"}
	case "transfer_pasien_antar_ruang":
		kode = []string{"nip_menyerahkan", "nip_menerima"}
	case "penilaian_pre_operasi", "penilaian_pre_anestesi", "penilaian_pre_induksi":
		kode = []string{"kd_dokter"}
	case "skor_aldrette_pasca_anestesi", "skor_steward_pasca_anestesi", "skor_bromage_pasca_anestesi":
		kode = []string{"nip"}
	case "laporan_operasi":
		// Dialog laporan mengikuti akses modul Jadwal Operasi, tanpa kolom pemilik.
		return true
	}
	for _, k := range kode {
		if data[k] == username {
			return true
		}
	}
	return false
}

func validasiPendukung(in *InputPendukung) error {
	bidang := FormPendukung(in.Jenis)
	if len(bidang) == 0 || in.NoRawat == "" || len(in.NoRawat) > 17 || in.Jadwal.NoRawat != in.NoRawat {
		return ErrValidasi
	}
	if _, err := time.Parse("2006-01-02", in.Jadwal.Tanggal); err != nil {
		return ErrValidasi
	}
	if in.Data == nil {
		return ErrValidasi
	}
	// Kateter tidak ada menggunakan sentinel Khanza, bukan waktu rekaan.
	if in.Jenis == "checklist_post_operasi" && in.Data["kateter_urine"] == "Tidak Ada" {
		in.Data["tanggal_pemasangan_kateter"] = "0000-00-00 00:00:00"
	}
	total := 0
	for _, b := range bidang {
		if b.Jenis == "computed" {
			continue
		}
		v := strings.TrimSpace(in.Data[b.Kode])
		if (b.Wajib && v == "") || utf8.RuneCountInString(v) > b.Batas {
			return fmt.Errorf("%w: %s wajib diisi sesuai batas %d karakter", ErrValidasi, b.Label, b.Batas)
		}
		if b.Jenis == "select" {
			index := -1
			for i, opsi := range b.Pilihan {
				if v == opsi {
					index = i
					break
				}
			}
			if index < 0 {
				return fmt.Errorf("%w: pilihan %s tidak valid", ErrValidasi, b.Label)
			}
			if b.NilaiKe != "" {
				in.Data[b.NilaiKe] = strconv.Itoa(index)
				total += index
			}
		}
		if b.Jenis == "date" || b.Jenis == "datetime-local" {
			zeroKateter := b.Kode == "tanggal_pemasangan_kateter" && in.Data["kateter_urine"] == "Tidak Ada"
			if !zeroKateter {
				layout := "2006-01-02"
				if b.Jenis == "datetime-local" {
					layout += " 15:04:05"
					v = strings.ReplaceAll(v, "T", " ")
					if len(v) == 16 {
						v += ":00"
					}
				}
				if _, err := time.Parse(layout, v); err != nil {
					return fmt.Errorf("%w: %s tidak valid", ErrValidasi, b.Label)
				}
			}
		}
		in.Data[b.Kode] = v
	}
	for _, b := range bidang {
		if b.Kode == "penilaian_totalnilai" {
			in.Data[b.Kode] = strconv.Itoa(total)
		}
	}
	if in.Jenis == "laporan_operasi" {
		if !strings.HasPrefix(in.Data["tanggal"], in.Jadwal.Tanggal+" ") || in.Data["selesaioperasi"] < in.Data["tanggal"] {
			return fmt.Errorf("%w: mulai harus sesuai tanggal jadwal dan selesai tidak boleh sebelum mulai", ErrValidasi)
		}
		if len(in.Data["laporan_operasi"]) > 60000 {
			return fmt.Errorf("%w: laporan terlalu panjang", ErrValidasi)
		}
	}
	if in.Jenis == "transfer_pasien_antar_ruang" && in.Data["tanggal_pindah"] < in.Data["tanggal_masuk"] {
		return fmt.Errorf("%w: tanggal pindah tidak boleh sebelum tanggal masuk", ErrValidasi)
	}
	return nil
}

func (r *Repositori) SimpanPendukung(ctx context.Context, in InputPendukung, metode, username string, admin bool) (err error) {
	bidang := FormPendukung(in.Jenis)
	if len(bidang) == 0 || in.NoRawat == "" || len(in.NoRawat) > 17 || in.Jadwal.NoRawat != in.NoRawat {
		return ErrValidasi
	}
	if metode != "POST" && metode != "PUT" && metode != "DELETE" {
		return ErrValidasi
	}
	if metode != "DELETE" {
		if err = validasiPendukung(&in); err != nil {
			return err
		}
	}
	if metode != "POST" {
		for _, b := range bidang {
			if _, ok := in.Asli[b.Kode]; !ok {
				return fmt.Errorf("%w: snapshot tidak lengkap", ErrValidasi)
			}
		}
		if in.Jenis == "laporan_operasi" && !strings.HasPrefix(in.Asli["tanggal"], in.Jadwal.Tanggal+" ") {
			return ErrValidasi
		}
		if !BolehUbahPendukung(in.Jenis, username, admin, in.Asli) {
			return ErrAksesPendukung
		}
	}
	if metode == "POST" && in.Jenis == "transfer_pasien_antar_ruang" && !BolehUbahPendukung(in.Jenis, username, admin, in.Data) {
		return ErrAksesPendukung
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	tx, err := r.simrs.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	defer func() {
		var sqlErr *mysql.MySQLError
		if errors.As(err, &sqlErr) && sqlErr.Number == 1062 {
			err = fmt.Errorf("%w: catatan pada waktu tersebut sudah ada", ErrKonflik)
		}
	}()
	var status string
	// Serialisasi penulisan SIRAPI per kunjungan, juga melindungi insert duplikat.
	err = tx.QueryRowContext(ctx, "SELECT stts FROM reg_periksa WHERE no_rawat=? FOR UPDATE", in.NoRawat).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrValidasi
	}
	if err != nil {
		return err
	}
	var terkunci bool
	if err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM billing WHERE no_rawat=?)", in.NoRawat).Scan(&terkunci); err != nil {
		return err
	}
	if terkunci || status == "Batal" {
		return ErrTerkunci
	}
	var jadwalAda bool
	if err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM booking_operasi WHERE no_rawat=? AND tanggal=? AND kode_paket=? AND jam_mulai=? AND jam_selesai=? AND kd_dokter=? AND kd_ruang_ok=?)",
		in.NoRawat, in.Jadwal.Tanggal, in.Jadwal.KodePaket, in.Jadwal.JamMulai, in.Jadwal.JamSelesai, in.Jadwal.KdDokter, in.Jadwal.KdRuangOK).Scan(&jadwalAda); err != nil {
		return err
	}
	if !jadwalAda {
		return fmt.Errorf("%w: jadwal SIMRS berubah atau tidak ditemukan", ErrKonflik)
	}
	kolom := []string{"no_rawat"}
	baru := []any{in.NoRawat}
	lama := []any{in.NoRawat}
	for _, b := range bidang {
		kolom = append(kolom, b.Kode)
		baru = append(baru, in.Data[b.Kode])
		lama = append(lama, in.Asli[b.Kode])
		if metode != "DELETE" && (b.Jenis == "dokter" || b.Jenis == "petugas") {
			tabel, kode := "dokter", "kd_dokter"
			if b.Jenis == "petugas" {
				tabel, kode = "petugas", "nip"
			}
			var ada bool
			if err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM "+tabel+" WHERE "+kode+"=?)", in.Data[b.Kode]).Scan(&ada); err != nil {
				return err
			}
			if !ada {
				return fmt.Errorf("%w: %s tidak ditemukan", ErrValidasi, b.Label)
			}
		}
	}
	// Tabel hanya dari kontrak statis; parameter request tidak menjadi identifier SQL.
	tabel := sumberPendukung[in.Jenis].tabel
	where, args := khanzamutasi.Kondisi(kolom, lama)
	if metode != "POST" {
		rows, e := tx.QueryContext(ctx, "SELECT 1 FROM "+tabel+" WHERE "+where+" LIMIT 2 FOR UPDATE", args...)
		if e != nil {
			return e
		}
		jumlah := 0
		for rows.Next() {
			jumlah++
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return e
		}
		if jumlah != 1 {
			return ErrKonflik
		}
	}
	if metode != "DELETE" {
		kunciTanggal := "tanggal"
		if in.Jenis == "transfer_pasien_antar_ruang" {
			kunciTanggal = "tanggal_masuk"
		}
		duplikat := "SELECT COUNT(*) FROM " + tabel + " WHERE no_rawat=? AND " + kunciTanggal + "=?"
		a := []any{in.NoRawat, in.Data[kunciTanggal]}
		if in.Jenis == "laporan_operasi" {
			duplikat = "SELECT COUNT(*) FROM laporan_operasi WHERE no_rawat=? AND DATE(tanggal)=?"
			a[1] = in.Jadwal.Tanggal
		}
		if metode == "PUT" {
			duplikat += " AND NOT (" + where + ")"
			a = append(a, args...)
		}
		var jumlah int
		if err = tx.QueryRowContext(ctx, duplikat, a...).Scan(&jumlah); err != nil {
			return err
		}
		if jumlah > 0 {
			return fmt.Errorf("%w: catatan sudah ada, gunakan Edit pada riwayat", ErrKonflik)
		}
	}
	switch metode {
	case "POST":
		placeholder := strings.TrimSuffix(strings.Repeat("?,", len(kolom)), ",")
		_, err = tx.ExecContext(ctx, "INSERT INTO "+tabel+" ("+strings.Join(kolom, ",")+") VALUES ("+placeholder+")", baru...)
	case "PUT":
		set := make([]string, len(kolom))
		for i, k := range kolom {
			set[i] = k + "=?"
		}
		_, err = tx.ExecContext(ctx, "UPDATE "+tabel+" SET "+strings.Join(set, ",")+" WHERE "+where+" LIMIT 1", append(baru, args...)...)
	case "DELETE":
		_, err = tx.ExecContext(ctx, "DELETE FROM "+tabel+" WHERE "+where+" LIMIT 1", args...)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}
