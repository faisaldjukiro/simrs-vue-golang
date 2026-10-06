package data_klaim

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"time"
)

type DokterSIMRS struct {
	NoRawat    string `json:"no_rawat"`
	Nama       string `json:"nama_dokter"`
	Sumber     string `json:"sumber"`
	Keterangan string `json:"keterangan"`
}

type RepositoriDokter struct{ db *sql.DB }

func NewRepositoriDokter(db *sql.DB) *RepositoriDokter { return &RepositoriDokter{db: db} }

type relasiDokter struct {
	noRawat string
	jenis   string
	nama    string
}

func cocokkanDokter(relasi []relasiDokter, jenis string) DokterSIMRS {
	if jenis != "Rawat Inap" && jenis != "Rawat Jalan" {
		return DokterSIMRS{Keterangan: "Jenis pelayanan klaim tidak dikenali"}
	}
	if len(relasi) == 0 {
		return DokterSIMRS{Keterangan: "SEP belum ditemukan di SIMRS"}
	}
	rawat := map[string]bool{}
	nama := map[string]bool{}
	kodeJenis := "2"
	sumber := "Dokter registrasi rawat jalan"
	if jenis == "Rawat Inap" {
		kodeJenis, sumber = "1", "DPJP rawat inap"
	}
	for _, r := range relasi {
		if r.jenis != kodeJenis {
			return DokterSIMRS{Keterangan: "Jenis pelayanan SEP berbeda; perlu pemeriksaan"}
		}
		rawat[r.noRawat] = true
		if r.nama != "" && r.nama != "-" {
			nama[r.nama] = true
		}
	}
	if len(rawat) != 1 {
		return DokterSIMRS{Keterangan: "SEP terkait beberapa nomor rawat; perlu pemeriksaan"}
	}
	out := DokterSIMRS{Sumber: sumber}
	for no := range rawat {
		out.NoRawat = no
	}
	if out.NoRawat == "" {
		out.Keterangan = "Nomor rawat pada SEP belum tersedia"
		return out
	}
	names := make([]string, 0, len(nama))
	for n := range nama {
		names = append(names, n)
	}
	sort.Strings(names)
	out.Nama = strings.Join(names, "; ")
	if out.Nama == "" {
		out.Keterangan = "Dokter belum tercatat atau referensi dokter tidak ditemukan"
	}
	return out
}

// Lengkapi membaca satu kumpulan SEP sekaligus. Tidak menggandakan baris atau nilai klaim.
func (r *RepositoriDokter) Lengkapi(ctx context.Context, hasil *Hasil) {
	klaim, err := ekstrakKlaim(hasil.Response)
	if err != nil {
		return
	}
	unik := map[string]bool{}
	for _, item := range klaim {
		if obj, ok := item.(map[string]any); ok {
			obj["dokter_simrs"] = DokterSIMRS{Keterangan: "Nomor SEP tidak tersedia"}
			if sep, ok := obj["noSEP"].(string); ok && strings.TrimSpace(sep) != "" {
				unik[strings.TrimSpace(sep)] = true
			}
		}
	}
	if len(unik) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	seps := make([]string, 0, len(unik))
	for sep := range unik {
		seps = append(seps, sep)
	}
	semua := map[string][]relasiDokter{}
	for awal := 0; awal < len(seps); awal += 400 {
		akhir := awal + 400
		if akhir > len(seps) {
			akhir = len(seps)
		}
		args := make([]any, 0, akhir-awal)
		for _, sep := range seps[awal:akhir] {
			args = append(args, sep)
		}
		data, e := r.baca(ctx, args)
		if e != nil {
			hasil.PeringatanSIMRS = "Data klaim tersedia, tetapi nama dokter gagal dibaca dari SIMRS. Periksa koneksi dan izin SELECT, lalu tarik ulang data."
			for _, item := range klaim {
				if obj, ok := item.(map[string]any); ok {
					obj["dokter_simrs"] = DokterSIMRS{Keterangan: "Pencocokan SIMRS gagal"}
				}
			}
			return
		}
		for sep, rows := range data {
			semua[sep] = rows
		}
	}
	for _, item := range klaim {
		if obj, ok := item.(map[string]any); ok {
			sep, _ := obj["noSEP"].(string)
			jenis, _ := obj["jenisPelayanan"].(string)
			if strings.TrimSpace(sep) != "" {
				obj["dokter_simrs"] = cocokkanDokter(semua[strings.TrimSpace(sep)], jenis)
			}
		}
	}
}

func (r *RepositoriDokter) baca(ctx context.Context, args []any) (map[string][]relasiDokter, error) {
	query := `SELECT DISTINCT s.no_sep, COALESCE(s.no_rawat,''), COALESCE(s.jnspelayanan,''),
		COALESCE(CASE WHEN s.jnspelayanan='1' THEN di.nm_dokter ELSE dj.nm_dokter END,'')
		FROM bridging_sep s
		LEFT JOIN reg_periksa rp ON rp.no_rawat=s.no_rawat
		LEFT JOIN dokter dj ON dj.kd_dokter=rp.kd_dokter
		LEFT JOIN dpjp_ranap dp ON dp.no_rawat=s.no_rawat AND s.jnspelayanan='1'
		LEFT JOIN dokter di ON di.kd_dokter=dp.kd_dokter
		WHERE s.no_sep IN (` + strings.TrimSuffix(strings.Repeat("?,", len(args)), ",") + `)`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hasil := map[string][]relasiDokter{}
	for rows.Next() {
		var sep string
		var row relasiDokter
		if err := rows.Scan(&sep, &row.noRawat, &row.jenis, &row.nama); err != nil {
			return nil, err
		}
		sep = strings.TrimSpace(sep)
		row.noRawat, row.jenis, row.nama = strings.TrimSpace(row.noRawat), strings.TrimSpace(row.jenis), strings.TrimSpace(row.nama)
		hasil[sep] = append(hasil[sep], row)
	}
	return hasil, rows.Err()
}
