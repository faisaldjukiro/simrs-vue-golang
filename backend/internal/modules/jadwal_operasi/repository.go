package jadwal_operasi

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"simrs-backend/internal/shared/arsiplokal"
	"simrs-backend/internal/shared/khanzamutasi"

	"github.com/go-sql-driver/mysql"
)

var ErrValidasi = errors.New("data jadwal tidak valid")
var ErrKonflik = errors.New("jadwal berubah atau akses perubahan tidak tersedia; muat ulang riwayat")
var ErrBentrok = errors.New("jadwal bentrok atau sudah ada pada ruang dan waktu tersebut")
var ErrTerkunci = errors.New("billing sudah terverifikasi atau kunjungan dibatalkan; jadwal hanya dapat dilihat")

type Input struct {
	Sumber         string `json:"sumber"`
	Asli           *Input `json:"asli"`
	ID             uint64 `json:"id"`
	Versi          int    `json:"versi"`
	NoRawat        string `json:"no_rawat"`
	KodePaket      string `json:"kode_paket"`
	Tanggal        string `json:"tanggal"`
	JamMulai       string `json:"jam_mulai"`
	JamSelesai     string `json:"jam_selesai"`
	Status         string `json:"status"`
	KdDokter       string `json:"kd_dokter"`
	KdRuangOK      string `json:"kd_ruang_ok"`
	DokterAnestesi string `json:"dokteranastesi"`
	Perawat        string `json:"perawat"`
}
type Jadwal struct {
	Input
	NamaPaket  string `json:"nama_paket"`
	NamaDokter string `json:"nama_dokter"`
	NamaRuang  string `json:"nama_ruang"`
	Sumber     string `json:"sumber"`
	BisaUbah   bool   `json:"bisa_ubah"`
}
type Hasil struct {
	Jadwal     []Jadwal `json:"jadwal"`
	Peringatan string   `json:"peringatan"`
	PesanKunci string   `json:"pesan_kunci"`
}
type Referensi struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}
type Repositori struct{ lokal, simrs *sql.DB }

func NewRepositori(lokal, simrs *sql.DB) *Repositori { return &Repositori{lokal: lokal, simrs: simrs} }

func Validasi(in *Input) error {
	for _, v := range []struct {
		nilai *string
		batas int
	}{
		{&in.NoRawat, 17}, {&in.KodePaket, 15}, {&in.KdDokter, 20}, {&in.KdRuangOK, 3},
	} {
		*v.nilai = strings.TrimSpace(*v.nilai)
		if *v.nilai == "" || utf8.RuneCountInString(*v.nilai) > v.batas {
			return fmt.Errorf("%w: pasien, paket operasi, operator, dan ruang wajib dipilih", ErrValidasi)
		}
	}
	if _, err := time.Parse("2006-01-02", in.Tanggal); err != nil {
		return fmt.Errorf("%w: tanggal tidak valid", ErrValidasi)
	}
	for _, jam := range []string{in.JamMulai, in.JamSelesai} {
		if _, err := time.Parse("15:04:05", jam); err != nil {
			return fmt.Errorf("%w: jam tidak valid", ErrValidasi)
		}
	}
	switch in.Status {
	case "Permintaan", "Menunggu", "Proses Operasi", "Selesai":
	default:
		return fmt.Errorf("%w: status tidak valid", ErrValidasi)
	}
	in.DokterAnestesi = strings.TrimSpace(in.DokterAnestesi)
	in.Perawat = strings.TrimSpace(in.Perawat)
	if utf8.RuneCountInString(in.DokterAnestesi) > 50 || utf8.RuneCountInString(in.Perawat) > 50 {
		return fmt.Errorf("%w: nama anestesi/perawat maksimal 50 karakter", ErrValidasi)
	}
	return nil
}

func (r *Repositori) kunci(ctx context.Context, no string) (string, error) {
	var batal bool
	err := r.simrs.QueryRowContext(ctx, "SELECT stts='Batal' FROM reg_periksa WHERE no_rawat=?", no).Scan(&batal)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w: kunjungan tidak ditemukan", ErrValidasi)
	}
	if err != nil {
		return "", err
	}
	var billing bool
	if err = r.simrs.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM billing WHERE no_rawat=?)", no).Scan(&billing); err != nil {
		return "", err
	}
	if batal || billing {
		return ErrTerkunci.Error(), nil
	}
	return "", nil
}

// Filter paket mengikuti DlgCariDaftarOperasi.setBayar dan set_tarif.
func (r *Repositori) filterPaket(ctx context.Context, no string) (string, []any, error) {
	var penjamin, posisi string
	if err := r.simrs.QueryRowContext(ctx, "SELECT kd_pj,status_lanjut FROM reg_periksa WHERE no_rawat=?", no).Scan(&penjamin, &posisi); err != nil {
		return "", nil, err
	}
	bayar, kelasAktif := "Yes", "Yes"
	err := r.simrs.QueryRowContext(ctx, "SELECT cara_bayar_operasi,kelas_operasi FROM set_tarif LIMIT 1").Scan(&bayar, &kelasAktif)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", nil, err
	}
	kelas := "Rawat Jalan"
	if posisi == "Ranap" {
		induk := no
		var ibu string
		err = r.simrs.QueryRowContext(ctx, "SELECT no_rawat FROM ranap_gabung WHERE no_rawat2=? LIMIT 1", no).Scan(&ibu)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", nil, err
		}
		if ibu != "" {
			induk = ibu
		}
		kelas = ""
		err = r.simrs.QueryRowContext(ctx, "SELECT k.kelas FROM kamar k JOIN kamar_inap ki ON ki.kd_kamar=k.kd_kamar WHERE ki.no_rawat=? AND ki.stts_pulang='-' ORDER BY ki.tgl_masuk DESC,ki.jam_masuk DESC LIMIT 1", induk).Scan(&kelas)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", nil, err
		}
	}
	where := "status='1'"
	args := []any{}
	if bayar == "Yes" {
		where += " AND (kd_pj=? OR kd_pj='-')"
		args = append(args, penjamin)
	}
	if kelasAktif == "Yes" {
		where += " AND (kelas=? OR kelas='-')"
		args = append(args, kelas)
	}
	return where, args, nil
}

func (r *Repositori) Referensi(ctx context.Context, jenis, q, no string) ([]Referensi, error) {
	tabel, kode, nama, where := "dokter", "kd_dokter", "nm_dokter", "status='1'"
	args := []any{}
	switch jenis {
	case "dokter":
	case "ruang":
		tabel, kode, nama, where = "ruang_ok", "kd_ruang_ok", "nm_ruang_ok", "1=1"
	case "paket":
		tabel, kode, nama = "paket_operasi", "kode_paket", "nm_perawatan"
		var err error
		where, args, err = r.filterPaket(ctx, no)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrValidasi
	}
	args = append(args, "%"+strings.TrimSpace(q)+"%", "%"+strings.TrimSpace(q)+"%")
	rows, err := r.simrs.QueryContext(ctx, "SELECT "+kode+","+nama+" FROM "+tabel+" WHERE "+where+" AND ("+kode+" LIKE ? OR "+nama+" LIKE ?) ORDER BY "+nama+" LIMIT 30", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hasil := []Referensi{}
	for rows.Next() {
		var v Referensi
		if err := rows.Scan(&v.Kode, &v.Nama); err != nil {
			return nil, err
		}
		hasil = append(hasil, v)
	}
	return hasil, rows.Err()
}

func (r *Repositori) nama(ctx context.Context, in Input) (string, string, string, error) {
	var paket, dokter, ruang string
	where, args, err := r.filterPaket(ctx, in.NoRawat)
	if err != nil {
		return "", "", "", err
	}
	args = append(args, in.KodePaket)
	err = r.simrs.QueryRowContext(ctx, "SELECT nm_perawatan FROM paket_operasi WHERE "+where+" AND kode_paket=?", args...).Scan(&paket)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", "", fmt.Errorf("%w: paket tidak sesuai kelas/penjamin atau tidak aktif", ErrValidasi)
	}
	if err != nil {
		return "", "", "", err
	}
	err = r.simrs.QueryRowContext(ctx, "SELECT nm_dokter FROM dokter WHERE kd_dokter=? AND status='1'", in.KdDokter).Scan(&dokter)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", "", fmt.Errorf("%w: operator tidak ditemukan atau tidak aktif", ErrValidasi)
	}
	if err != nil {
		return "", "", "", err
	}
	err = r.simrs.QueryRowContext(ctx, "SELECT nm_ruang_ok FROM ruang_ok WHERE kd_ruang_ok=?", in.KdRuangOK).Scan(&ruang)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", "", fmt.Errorf("%w: ruang operasi tidak ditemukan", ErrValidasi)
	}
	return paket, dokter, ruang, err
}

// Simpan memakai tabel SIMRS saja. Named lock menggantikan tabel mutex lokal.
func (r *Repositori) Simpan(ctx context.Context, in Input, user uint64, admin bool) error {
	if in.Sumber != "Khanza" && (in.ID != 0 || in.Sumber == "SIRAPI") {
		return fmt.Errorf("%w: arsip lokal hanya baca; rekonsiliasi diperlukan", ErrValidasi)
	}
	if err := Validasi(&in); err != nil {
		return err
	}
	pesan, err := r.kunci(ctx, in.NoRawat)
	if err != nil {
		return err
	}
	if pesan != "" {
		return ErrTerkunci
	}
	if _, _, _, err = r.nama(ctx, in); err != nil {
		return err
	}
	conn, err := r.simrs.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	const lockName = "CONCAT('sirapi.booking.',LEFT(SHA2(DATABASE(),256),32))"
	var locked sql.NullInt64
	if err = conn.QueryRowContext(ctx, "SELECT GET_LOCK("+lockName+",10)").Scan(&locked); err != nil {
		return err
	}
	if !locked.Valid || locked.Int64 != 1 {
		return ErrKonflik
	}
	defer func() {
		release, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var freed sql.NullInt64
		if e := conn.QueryRowContext(release, "SELECT RELEASE_LOCK("+lockName+")").Scan(&freed); e != nil || !freed.Valid || freed.Int64 != 1 {
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
		}
	}()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	pengecualian := ""
	argsAsli := []any{}
	whereAsli := ""
	if in.Sumber == "Khanza" {
		if in.Asli == nil || in.Asli.NoRawat != in.NoRawat {
			return ErrValidasi
		}
		kolom, lama := snapshot(*in.Asli)
		whereAsli, argsAsli = khanzamutasi.Kondisi(kolom, lama)
		rows, e := tx.QueryContext(ctx, "SELECT 1 FROM booking_operasi WHERE "+whereAsli+" LIMIT 2 FOR UPDATE", argsAsli...)
		if e != nil {
			return e
		}
		count := 0
		for rows.Next() {
			count++
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return e
		}
		if count != 1 {
			return ErrKonflik
		}
		pengecualian = " AND NOT (" + whereAsli + ")"
	}
	var jumlah int
	args := append([]any{in.NoRawat, in.KodePaket, in.Tanggal, in.JamMulai}, argsAsli...)
	err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM booking_operasi WHERE no_rawat=? AND kode_paket=? AND tanggal=? AND jam_mulai=?"+pengecualian, args...).Scan(&jumlah)
	if err != nil {
		return err
	}
	if jumlah > 0 {
		return ErrBentrok
	}
	// Sesuai BtnSimpan/BtnEdit Java: jam mulai pasien lain dalam rentang inklusif.
	if in.JamMulai != "00:00:00" {
		err = tx.QueryRowContext(ctx, queryBentrokJava,
			in.Tanggal, in.KdRuangOK, in.NoRawat, in.JamMulai, in.JamSelesai).Scan(&jumlah)
		if err != nil {
			return err
		}
		if jumlah > 0 {
			return ErrBentrok
		}
	}
	kolom, baru := snapshot(in)
	if in.Sumber == "Khanza" {
		// Java hanya mengubah delapan kolom utama, bukan anestesi/perawat.
		kolom, baru = kolom[:8], baru[:8]
		set := make([]string, len(kolom))
		for i, k := range kolom {
			set[i] = k + "=?"
		}
		_, err = tx.ExecContext(ctx, "UPDATE booking_operasi SET "+strings.Join(set, ",")+" WHERE "+whereAsli+" LIMIT 1", append(baru, argsAsli...)...)
	} else {
		_, err = tx.ExecContext(ctx, "INSERT INTO booking_operasi ("+strings.Join(kolom, ",")+") VALUES (?,?,?,?,?,?,?,?,?,?)", baru...)
	}
	var e *mysql.MySQLError
	if errors.As(err, &e) && e.Number == 1062 {
		return ErrBentrok
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repositori) Hapus(ctx context.Context, in Input, user uint64, admin bool) error {
	if in.Sumber != "Khanza" {
		return fmt.Errorf("%w: arsip lokal hanya baca; rekonsiliasi diperlukan", ErrValidasi)
	}
	pesan, err := r.kunci(ctx, in.NoRawat)
	if err != nil {
		return err
	}
	if pesan != "" {
		return ErrTerkunci
	}
	return r.mutasiKhanza(ctx, in, true)
}

func (r *Repositori) Daftar(ctx context.Context, no string, user uint64, admin bool) (Hasil, error) {
	hasil := Hasil{Jadwal: []Jadwal{}}
	if strings.TrimSpace(no) == "" || len(no) > 17 {
		return hasil, ErrValidasi
	}
	var err error
	hasil.PesanKunci, err = r.kunci(ctx, no)
	if err != nil {
		return hasil, err
	}
	rows, err := r.lokal.QueryContext(ctx, "SELECT id,versi,no_rawat,kode_paket,DATE_FORMAT(tanggal,'%Y-%m-%d'),CAST(jam_mulai AS CHAR),CAST(jam_selesai AS CHAR),status,kd_dokter,kd_ruang_ok,dokteranastesi,perawat,nama_paket,nama_dokter,nama_ruang,dibuat_oleh FROM sirapi_jadwal_operasi WHERE no_rawat=? AND deleted_at IS NULL", no)
	if err != nil && !arsiplokal.TabelTidakAda(err) {
		return hasil, err
	}
	if err == nil {
		for rows.Next() {
			var v Jadwal
			var pembuat uint64
			if err = rows.Scan(&v.ID, &v.Versi, &v.NoRawat, &v.KodePaket, &v.Tanggal, &v.JamMulai, &v.JamSelesai, &v.Status, &v.KdDokter, &v.KdRuangOK, &v.DokterAnestesi, &v.Perawat, &v.NamaPaket, &v.NamaDokter, &v.NamaRuang, &pembuat); err != nil {
				rows.Close()
				return hasil, err
			}
			v.Sumber = "SIRAPI"
			v.BisaUbah = false
			hasil.Jadwal = append(hasil.Jadwal, v)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return hasil, err
		}
	}
	lama, err := r.riwayat(ctx, no)
	if err != nil {
		return hasil, err
	} else {
		for i := range lama {
			lama[i].BisaUbah = hasil.PesanKunci == ""
		}
		hasil.Jadwal = append(hasil.Jadwal, lama...)
	}
	if len(hasil.Jadwal) > len(lama) {
		hasil.Peringatan = "Ada arsip lokal yang belum direkonsiliasi. Arsip hanya baca; penyimpanan baru langsung ke SIMRS."
	}
	sort.SliceStable(hasil.Jadwal, func(i, j int) bool {
		return hasil.Jadwal[i].Tanggal+hasil.Jadwal[i].JamMulai > hasil.Jadwal[j].Tanggal+hasil.Jadwal[j].JamMulai
	})
	return hasil, nil
}

func (r *Repositori) riwayat(ctx context.Context, no string) ([]Jadwal, error) {
	rows, err := r.simrs.QueryContext(ctx, `SELECT b.no_rawat,b.kode_paket,DATE_FORMAT(b.tanggal,'%Y-%m-%d'),CAST(b.jam_mulai AS CHAR),CAST(b.jam_selesai AS CHAR),b.status,b.kd_dokter,b.kd_ruang_ok,COALESCE(b.dokteranastesi,''),COALESCE(b.perawat,''),COALESCE(p.nm_perawatan,''),COALESCE(d.nm_dokter,''),COALESCE(r.nm_ruang_ok,'')
 FROM booking_operasi b LEFT JOIN paket_operasi p ON p.kode_paket=b.kode_paket LEFT JOIN dokter d ON d.kd_dokter=b.kd_dokter LEFT JOIN ruang_ok r ON r.kd_ruang_ok=b.kd_ruang_ok WHERE b.no_rawat=?`, no)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hasil := []Jadwal{}
	for rows.Next() {
		var v Jadwal
		if err = rows.Scan(&v.NoRawat, &v.KodePaket, &v.Tanggal, &v.JamMulai, &v.JamSelesai, &v.Status, &v.KdDokter, &v.KdRuangOK, &v.DokterAnestesi, &v.Perawat, &v.NamaPaket, &v.NamaDokter, &v.NamaRuang); err != nil {
			return nil, err
		}
		v.Sumber = "Khanza"
		hasil = append(hasil, v)
	}
	return hasil, rows.Err()
}
