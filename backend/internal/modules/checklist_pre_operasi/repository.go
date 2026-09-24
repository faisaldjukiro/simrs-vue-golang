package checklist_pre_operasi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"simrs-backend/internal/shared/khanzamutasi"

	"github.com/go-sql-driver/mysql"
)

var ErrValidasi = errors.New("validasi checklist")
var ErrKonflik = errors.New("Data berubah, waktu checklist sudah digunakan, atau akses perubahan tidak tersedia. Muat ulang riwayat.")

func snapshot(input Input) ([]string, []any) {
	kolom := []string{"no_rawat", "tanggal"}
	nilai := []any{input.NoRawat, input.Tanggal}
	for _, b := range BidangForm {
		kolom = append(kolom, b.Kode)
		nilai = append(nilai, input.Data[b.Kode])
	}
	return kolom, nilai
}

func (r *Repositori) mutasiKhanza(ctx context.Context, input Input, hapus bool) error {
	if input.Asli == nil || input.NoRawat == "" || input.Asli.NoRawat != input.NoRawat || input.Asli.Tanggal == "" {
		return ErrValidasi
	}
	for _, b := range BidangForm {
		if _, ok := input.Asli.Data[b.Kode]; !ok {
			return ErrValidasi
		}
	}
	kolom, lama := snapshot(*input.Asli)
	_, baru := snapshot(input)
	err := khanzamutasi.Jalankan(ctx, r.simrs, "checklist_pre_operasi", kolom, lama, baru, hapus)
	if errors.Is(err, khanzamutasi.ErrKonflik) {
		return ErrKonflik
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrKonflik
	}
	return err
}

type Bidang struct {
	Kode, Label string
	Maksimal    int
	Pilihan     []string
}

var BidangForm = []Bidang{
	{"sncn", "SN/CN", 25, nil},
	{"tindakan", "Tindakan", 50, nil},
	{"kd_dokter_bedah", "Dokter Bedah", 20, nil},
	{"kd_dokter_anestesi", "Dokter Anestesi", 20, nil},
	{"identitas", "Identitas", 0, []string{"Ya", "Tidak"}},
	{"surat_ijin_bedah", "Surat Izin Bedah", 0, []string{"Ada", "Tidak Ada"}},
	{"surat_ijin_anestesi", "Surat Izin Anestesi", 0, []string{"Ada", "Tidak Ada"}},
	{"surat_ijin_transfusi", "Surat Izin Transfusi", 0, []string{"Ada", "Tidak Ada", "Tidak Diperlukan"}},
	{"penandaan_area_operasi", "Penandaan Area Operasi", 0, []string{"Ada", "Tidak Ada", "Tidak Diperlukan"}},
	{"keadaan_umum", "Keadaan Umum", 0, []string{"Baik", "Sedang", "Lemah"}},
	{"pemeriksaan_penunjang_rontgen", "RONTGEN", 0, []string{"Ada", "Tidak Ada", "Tidak Diperlukan"}},
	{"keterangan_pemeriksaan_penunjang_rontgen", "Keterangan RONTGEN", 20, nil},
	{"pemeriksaan_penunjang_ekg", "EKG", 0, []string{"Ada", "Tidak Ada", "Tidak Diperlukan"}},
	{"keterangan_pemeriksaan_penunjang_ekg", "Keterangan EKG", 20, nil},
	{"pemeriksaan_penunjang_usg", "USG", 0, []string{"Ada", "Tidak Ada", "Tidak Diperlukan"}},
	{"keterangan_pemeriksaan_penunjang_usg", "Keterangan USG", 20, nil},
	{"pemeriksaan_penunjang_ctscan", "CTSCAN", 0, []string{"Ada", "Tidak Ada", "Tidak Diperlukan"}},
	{"keterangan_pemeriksaan_penunjang_ctscan", "Keterangan CTSCAN", 20, nil},
	{"pemeriksaan_penunjang_mri", "MRI", 0, []string{"Ada", "Tidak Ada", "Tidak Diperlukan"}},
	{"keterangan_pemeriksaan_penunjang_mri", "Keterangan MRI", 20, nil},
	{"persiapan_darah", "Persiapan Darah", 0, []string{"Ada", "Tidak Ada", "Tidak Diperlukan"}},
	{"keterangan_persiapan_darah", "Keterangan Persiapan Darah", 20, nil},
	{"perlengkapan_khusus", "Perlengkapan Khusus", 0, []string{"Ada", "Tidak Ada", "Tidak Diperlukan"}},
	{"nip_petugas_ruangan", "Petugas Ruangan", 20, nil},
	{"nip_perawat_ok", "Petugas OK", 20, nil},
}

type Input struct {
	Sumber  string            `json:"sumber"`
	Asli    *Input            `json:"asli"`
	ID      uint64            `json:"id"`
	Versi   int               `json:"versi"`
	NoRawat string            `json:"no_rawat"`
	Tanggal string            `json:"tanggal"`
	Data    map[string]string `json:"data"`
}

type Catatan struct {
	Input
	Sumber   string `json:"sumber"`
	BisaUbah bool   `json:"bisa_ubah"`
	Pembuat  string `json:"pembuat"`
}

type Referensi struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}

type Hasil struct {
	Catatan    []Catatan `json:"catatan"`
	Peringatan string    `json:"peringatan"`
}

type Repositori struct{ lokal, simrs *sql.DB }

func NewRepositori(lokal, simrs *sql.DB) *Repositori { return &Repositori{lokal: lokal, simrs: simrs} }

func Validasi(input *Input) error {
	input.NoRawat = strings.TrimSpace(input.NoRawat)
	if input.NoRawat == "" || len(input.NoRawat) > 17 {
		return fmt.Errorf("%w: nomor rawat tidak valid", ErrValidasi)
	}
	if _, err := time.Parse("2006-01-02 15:04:05", input.Tanggal); err != nil {
		return fmt.Errorf("%w: tanggal dan jam wajib diisi lengkap", ErrValidasi)
	}
	data := map[string]string{}
	for _, bidang := range BidangForm {
		nilai := strings.TrimSpace(input.Data[bidang.Kode])
		if bidang.Pilihan != nil {
			cocok := false
			for _, opsi := range bidang.Pilihan {
				if nilai == opsi {
					cocok = true
				}
			}
			if !cocok {
				return fmt.Errorf("%w: pilih %s", ErrValidasi, bidang.Label)
			}
		} else if utf8.RuneCountInString(nilai) > bidang.Maksimal {
			return fmt.Errorf("%w: %s maksimal %d karakter", ErrValidasi, bidang.Label, bidang.Maksimal)
		}
		if !strings.HasPrefix(bidang.Kode, "keterangan_") && nilai == "" {
			return fmt.Errorf("%w: %s wajib diisi", ErrValidasi, bidang.Label)
		}
		data[bidang.Kode] = nilai
	}
	input.Data = data
	return nil
}

func (r *Repositori) Referensi(ctx context.Context, jenis, q string) ([]Referensi, error) {
	tabel, kode, nama := "dokter", "kd_dokter", "nm_dokter"
	if jenis == "petugas" {
		tabel, kode, nama = "petugas", "nip", "nama"
	} else if jenis != "dokter" {
		return nil, fmt.Errorf("%w: jenis referensi tidak valid", ErrValidasi)
	}
	rows, err := r.simrs.QueryContext(ctx, "SELECT "+kode+", "+nama+" FROM "+tabel+" WHERE ("+kode+" LIKE ? OR "+nama+" LIKE ?) ORDER BY "+nama+" LIMIT 30", "%"+q+"%", "%"+q+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hasil := []Referensi{}
	for rows.Next() {
		var item Referensi
		if err := rows.Scan(&item.Kode, &item.Nama); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *Repositori) lengkapiReferensi(ctx context.Context, input *Input) error {
	var ada int
	if err := r.simrs.QueryRowContext(ctx, "SELECT COUNT(*) FROM reg_periksa WHERE no_rawat = ?", input.NoRawat).Scan(&ada); err != nil {
		return err
	}
	if ada == 0 {
		return fmt.Errorf("%w: kunjungan pasien tidak ditemukan", ErrValidasi)
	}
	for _, ref := range []struct{ field, tabel, kode, nama string }{
		{"kd_dokter_bedah", "dokter", "kd_dokter", "nm_dokter"},
		{"kd_dokter_anestesi", "dokter", "kd_dokter", "nm_dokter"},
		{"nip_petugas_ruangan", "petugas", "nip", "nama"},
		{"nip_perawat_ok", "petugas", "nip", "nama"},
	} {
		var nama string
		err := r.simrs.QueryRowContext(ctx, "SELECT "+ref.nama+" FROM "+ref.tabel+" WHERE "+ref.kode+" = ?", input.Data[ref.field]).Scan(&nama)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: kode petugas/dokter tidak ditemukan", ErrValidasi)
		}
		if err != nil {
			return err
		}
		input.Data[ref.field+"_nama"] = nama
	}
	return nil
}

func (r *Repositori) Simpan(ctx context.Context, input Input, user uint64, admin bool) error {
	if err := Validasi(&input); err != nil {
		return err
	}
	if err := r.lengkapiReferensi(ctx, &input); err != nil {
		return err
	}
	if input.Sumber == "Khanza" {
		return r.mutasiKhanza(ctx, input, false)
	}
	payload, err := json.Marshal(input.Data)
	if err != nil {
		return err
	}
	var hasil sql.Result
	if input.ID == 0 {
		// Izin eksplisit user: catatan baru langsung ke tabel Khanza.
		// Nama kolom dari daftar tetap, bukan input client; sesuai 27 field referensi.
		kolom := []string{"no_rawat", "tanggal"}
		args := []any{input.NoRawat, input.Tanggal}
		for _, bidang := range BidangForm {
			kolom = append(kolom, bidang.Kode)
			args = append(args, input.Data[bidang.Kode])
		}
		placeholder := strings.TrimSuffix(strings.Repeat("?,", len(kolom)), ",")
		hasil, err = r.simrs.ExecContext(ctx,
			"INSERT INTO checklist_pre_operasi ("+strings.Join(kolom, ",")+") VALUES ("+placeholder+")",
			args...)
	} else {
		hasil, err = r.lokal.ExecContext(ctx, "UPDATE sirapi_checklist_pre_operasi SET tanggal=?,data_checklist=?,versi=versi+1,diubah_oleh=? WHERE id=? AND no_rawat=? AND versi=? AND deleted_at IS NULL AND (dibuat_oleh=? OR ?)", input.Tanggal, payload, user, input.ID, input.NoRawat, input.Versi, user, admin)
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrKonflik
	}
	if err != nil {
		return err
	}
	jumlah, err := hasil.RowsAffected()
	if err == nil && jumlah == 0 {
		return ErrKonflik
	}
	return err
}

func (r *Repositori) Hapus(ctx context.Context, input Input, user uint64, admin bool) error {
	if input.Sumber == "Khanza" {
		return r.mutasiKhanza(ctx, input, true)
	}
	if input.ID == 0 || input.NoRawat == "" || input.Versi < 1 {
		return fmt.Errorf("%w: pilih catatan yang akan dihapus", ErrValidasi)
	}
	hasil, err := r.lokal.ExecContext(ctx, "UPDATE sirapi_checklist_pre_operasi SET deleted_at=NOW(),diubah_oleh=?,versi=versi+1 WHERE id=? AND no_rawat=? AND versi=? AND deleted_at IS NULL AND (dibuat_oleh=? OR ?)", user, input.ID, input.NoRawat, input.Versi, user, admin)
	if err != nil {
		return err
	}
	jumlah, err := hasil.RowsAffected()
	if err == nil && jumlah == 0 {
		return ErrKonflik
	}
	return err
}

func (r *Repositori) Daftar(ctx context.Context, noRawat string, user uint64, admin bool) (Hasil, error) {
	hasil := Hasil{Catatan: []Catatan{}}
	if strings.TrimSpace(noRawat) == "" || len(noRawat) > 17 {
		return hasil, fmt.Errorf("%w: nomor rawat wajib diisi", ErrValidasi)
	}
	rows, err := r.lokal.QueryContext(ctx, "SELECT id,no_rawat,DATE_FORMAT(tanggal,'%Y-%m-%d %H:%i:%s'),data_checklist,dibuat_oleh,versi FROM sirapi_checklist_pre_operasi WHERE no_rawat=? AND deleted_at IS NULL ORDER BY tanggal DESC", noRawat)
	if err != nil {
		return hasil, err
	}
	for rows.Next() {
		var item Catatan
		var payload []byte
		var pembuat uint64
		if err := rows.Scan(&item.ID, &item.NoRawat, &item.Tanggal, &payload, &pembuat, &item.Versi); err != nil {
			rows.Close()
			return hasil, err
		}
		if err := json.Unmarshal(payload, &item.Data); err != nil {
			rows.Close()
			return hasil, err
		}
		item.Sumber = "SIRAPI"
		item.BisaUbah = admin || pembuat == user
		item.Pembuat = fmt.Sprint(pembuat)
		hasil.Catatan = append(hasil.Catatan, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return hasil, err
	}
	lama, err := r.riwayatLama(ctx, noRawat)
	if err != nil {
		// Jangan samarkan kegagalan riwayat sebagai daftar kosong.
		hasil.Peringatan = "Riwayat Khanza belum dapat dibaca. Yang ditampilkan hanya catatan SIRAPI."
	} else {
		hasil.Catatan = append(hasil.Catatan, lama...)
	}
	sort.SliceStable(hasil.Catatan, func(i, j int) bool { return hasil.Catatan[i].Tanggal > hasil.Catatan[j].Tanggal })
	return hasil, nil
}

func (r *Repositori) riwayatLama(ctx context.Context, noRawat string) ([]Catatan, error) {
	kolom := []string{"DATE_FORMAT(c.tanggal,'%Y-%m-%d %H:%i:%s')"}
	for _, bidang := range BidangForm {
		kolom = append(kolom, "COALESCE(c."+bidang.Kode+",'')")
	}
	kolom = append(kolom, "COALESCE(db.nm_dokter,'')", "COALESCE(da.nm_dokter,'')", "COALESCE(pr.nama,'')", "COALESCE(po.nama,'')")
	rows, err := r.simrs.QueryContext(ctx, "SELECT "+strings.Join(kolom, ",")+" FROM checklist_pre_operasi c LEFT JOIN dokter db ON db.kd_dokter=c.kd_dokter_bedah LEFT JOIN dokter da ON da.kd_dokter=c.kd_dokter_anestesi LEFT JOIN petugas pr ON pr.nip=c.nip_petugas_ruangan LEFT JOIN petugas po ON po.nip=c.nip_perawat_ok WHERE c.no_rawat=? ORDER BY c.tanggal DESC", noRawat)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hasil := []Catatan{}
	for rows.Next() {
		values := make([]string, len(kolom))
		tujuan := make([]any, len(kolom))
		for i := range values {
			tujuan[i] = &values[i]
		}
		if err := rows.Scan(tujuan...); err != nil {
			return nil, err
		}
		item := Catatan{Input: Input{NoRawat: noRawat, Tanggal: values[0], Data: map[string]string{}}, Sumber: "Khanza", BisaUbah: true}
		for i, bidang := range BidangForm {
			item.Data[bidang.Kode] = values[i+1]
		}
		for i, kode := range []string{"kd_dokter_bedah", "kd_dokter_anestesi", "nip_petugas_ruangan", "nip_perawat_ok"} {
			item.Data[kode+"_nama"] = values[len(BidangForm)+1+i]
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}
