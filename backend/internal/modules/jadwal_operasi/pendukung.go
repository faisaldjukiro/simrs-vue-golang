package jadwal_operasi

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Daftar tetap dari dialog yang dibuka DlgBookingOperasi.java.
// Endpoint ini hanya membaca; nama tabel/kolom tidak berasal dari parameter HTTP.
var sumberPendukung = map[string]struct{ tabel, tanggal string }{
	"kamar_inap":                   {"kamar_inap", "tgl_masuk"},
	"penilaian_pre_induksi":        {"penilaian_pre_induksi", "tanggal"},
	"signin_sebelum_anestesi":      {"signin_sebelum_anestesi", "tanggal"},
	"timeout_sebelum_insisi":       {"timeout_sebelum_insisi", "tanggal"},
	"signout_sebelum_menutup_luka": {"signout_sebelum_menutup_luka", "tanggal"},
	"checklist_post_operasi":       {"checklist_post_operasi", "tanggal"},
	"penilaian_pre_operasi":        {"penilaian_pre_operasi", "tanggal"},
	"penilaian_pre_anestesi":       {"penilaian_pre_anestesi", "tanggal"},
	"laporan_operasi":              {"laporan_operasi", "tanggal"},
	"tagihan_operasi":              {"operasi", "tgl_operasi"},
	"transfer_pasien_antar_ruang":  {"transfer_pasien_antar_ruang", "tanggal_pindah"},
	"skor_aldrette_pasca_anestesi": {"skor_aldrette_pasca_anestesi", "tanggal"},
	"skor_steward_pasca_anestesi":  {"skor_steward_pasca_anestesi", "tanggal"},
	"skor_bromage_pasca_anestesi":  {"skor_bromage_pasca_anestesi", "tanggal"},
}

type RiwayatPendukung struct {
	Kolom      []string            `json:"kolom"`
	Catatan    []map[string]string `json:"catatan"`
	Form       []BidangOperasi     `json:"form"`
	PesanKunci string              `json:"pesan_kunci"`
}

func (r *Repositori) Pendukung(ctx context.Context, jenis, no, tanggal string) (RiwayatPendukung, error) {
	hasil := RiwayatPendukung{Kolom: []string{}, Catatan: []map[string]string{}}
	sumber, ok := sumberPendukung[jenis]
	if !ok || strings.TrimSpace(no) == "" || len(no) > 17 {
		return hasil, ErrValidasi
	}
	if _, err := time.Parse("2006-01-02", tanggal); err != nil {
		return hasil, ErrValidasi
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	var ada bool
	if err := r.simrs.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM reg_periksa WHERE no_rawat=?)", no).Scan(&ada); err != nil {
		return hasil, err
	}
	if !ada {
		return hasil, ErrValidasi
	}
	hasil.Form = FormPendukung(jenis)
	if len(hasil.Form) > 0 {
		pesan, err := r.kunci(ctx, no)
		if err != nil {
			return hasil, err
		}
		hasil.PesanKunci = pesan
	}
	query := "SELECT * FROM `" + sumber.tabel + "` WHERE no_rawat=?"
	args := []any{no}
	// Laporan operasi pada dialog lama dibatasi tanggal jadwal; modul lain
	// menampilkan riwayat kunjungan, termasuk penilaian sebelum hari operasi.
	if jenis == "laporan_operasi" {
		query += " AND DATE(tanggal)=?"
		args = append(args, tanggal)
	}
	query += " ORDER BY `" + sumber.tanggal + "` DESC"
	rows, err := r.simrs.QueryContext(ctx, query, args...)
	if err != nil {
		return hasil, fmt.Errorf("riwayat pendukung: %w", err)
	}
	defer rows.Close()
	kolom, err := rows.Columns()
	if err != nil {
		return hasil, err
	}
	for _, nama := range kolom {
		if nama != "no_rawat" {
			hasil.Kolom = append(hasil.Kolom, nama)
		}
	}
	for rows.Next() {
		nilai := make([]any, len(kolom))
		tujuan := make([]any, len(kolom))
		for i := range nilai {
			tujuan[i] = &nilai[i]
		}
		if err := rows.Scan(tujuan...); err != nil {
			return hasil, err
		}
		baris := map[string]string{}
		for i, nama := range kolom {
			if nama != "no_rawat" {
				switch v := nilai[i].(type) {
				case nil:
					baris[nama] = ""
				case time.Time:
					if v.IsZero() {
						baris[nama] = "0000-00-00 00:00:00"
					} else {
						baris[nama] = v.Format("2006-01-02 15:04:05")
					}
					if nama == "tanggal_steril" {
						baris[nama] = baris[nama][:10]
					}
				case []byte:
					baris[nama] = string(v)
				default:
					baris[nama] = fmt.Sprint(v)
				}
			}
		}
		hasil.Catatan = append(hasil.Catatan, baris)
	}
	return hasil, rows.Err()
}
