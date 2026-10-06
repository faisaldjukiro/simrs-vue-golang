package kardeks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrValidasi = errors.New("filter kardeks tidak valid")
var errBatas = errors.New("jumlah catatan melebihi batas tampilan")

type Bagian struct {
	Kode  string              `json:"kode"`
	Nama  string              `json:"nama"`
	Baris []map[string]string `json:"baris"`
	Error string              `json:"error,omitempty"`
}
type Hasil struct {
	Mulai   string   `json:"mulai"`
	Selesai string   `json:"selesai"`
	Bagian  []Bagian `json:"bagian"`
}
type Repositori struct{ simrs *sql.DB }

func NewRepositori(db *sql.DB) *Repositori { return &Repositori{simrs: db} }
func Periode(no, tanggal, jam string) (time.Time, time.Time, error) {
	if no == "" || len(no) > 17 {
		return time.Time{}, time.Time{}, ErrValidasi
	}
	awal, err := time.ParseInLocation("2006-01-02 15:04", tanggal+" "+jam, time.FixedZone("WITA", 8*3600))
	if err != nil {
		return awal, awal, fmt.Errorf("%w: tanggal atau jam mulai tidak valid", ErrValidasi)
	}
	return awal, awal.Add(24 * time.Hour), nil
}

var sumber = []struct{ kode, nama, query string }{
	{"cairan", "Cairan masuk dan keluar (volume aktual mL)", `SELECT CAST(id AS CHAR) id, DATE_FORMAT(waktu_mulai,'%Y-%m-%d %H:%i:%s') waktu_mulai, DATE_FORMAT(waktu_selesai,'%Y-%m-%d %H:%i:%s') waktu_selesai, jenis, kategori, COALESCE(rincian,'') rincian, CAST(volume_ml AS CHAR) volume_ml, petugas, COALESCE(catatan,'') catatan FROM catatan_cairan_ranap WHERE no_rawat=? AND waktu_selesai>? AND waktu_mulai<? ORDER BY waktu_mulai,id`},
	{"setting_ventilator", "Pengaturan Ventilator (nilai pada waktu pencatatan)", `SELECT DATE_FORMAT(s.waktu_setting,'%Y-%m-%d %H:%i:%s') waktu,CAST(s.id_pemakaian AS CHAR) id_pemakaian,p.kode_ventilator,s.mode_ventilator,COALESCE(CAST(s.fio2_persen AS CHAR),'') fio2_persen,COALESCE(CAST(s.peep_cmh2o AS CHAR),'') peep_cmh2o,COALESCE(CAST(s.tidal_volume_ml AS CHAR),'') tidal_volume_ml,COALESCE(CAST(s.frekuensi_set AS CHAR),'') frekuensi_set,COALESCE(CAST(s.pressure_control_cmh2o AS CHAR),'') pressure_control_cmh2o,COALESCE(CAST(s.pressure_support_cmh2o AS CHAR),'') pressure_support_cmh2o,s.petugas,COALESCE(s.catatan,'') catatan FROM setting_ventilator s JOIN pemakaian_ventilator p ON p.id=s.id_pemakaian WHERE p.no_rawat=? AND s.waktu_setting>=? AND s.waktu_setting<? ORDER BY s.waktu_setting,s.id`},
	{"monitoring_ventilator", "Monitoring Ventilator", `SELECT DATE_FORMAT(s.waktu_monitoring,'%Y-%m-%d %H:%i:%s') waktu,CAST(s.id_pemakaian AS CHAR) id_pemakaian,p.kode_ventilator,COALESCE(s.tekanan_darah,'') td,COALESCE(CAST(s.nadi_per_menit AS CHAR),'') nadi,COALESCE(CAST(s.respirasi_per_menit AS CHAR),'') respirasi,COALESCE(CAST(s.suhu_celsius AS CHAR),'') suhu,COALESCE(CAST(s.spo2_persen AS CHAR),'') spo2,COALESCE(s.kesadaran,'') kesadaran,COALESCE(s.tahap,'') tahap,s.petugas,COALESCE(s.catatan,'') catatan FROM monitoring_ventilator s JOIN pemakaian_ventilator p ON p.id=s.id_pemakaian WHERE p.no_rawat=? AND s.waktu_monitoring>=? AND s.waktu_monitoring<? ORDER BY s.waktu_monitoring,s.id`},
	{"observasi", "Observasi Rawat Inap", `SELECT DATE_FORMAT(TIMESTAMP(o.tgl_perawatan,o.jam_rawat),'%Y-%m-%d %H:%i:%s') waktu,COALESCE(o.td,'') td,COALESCE(o.hr,'') nadi,COALESCE(o.rr,'') respirasi,COALESCE(o.suhu,'') suhu,COALESCE(o.spo2,'') spo2,COALESCE(o.gcs,'') gcs,o.nip petugas FROM catatan_observasi_ranap o WHERE o.no_rawat=? AND TIMESTAMP(o.tgl_perawatan,o.jam_rawat)>=? AND TIMESTAMP(o.tgl_perawatan,o.jam_rawat)<? ORDER BY o.tgl_perawatan,o.jam_rawat`},
	{"pemeriksaan", "Pemeriksaan Rawat Inap / CPPT", `SELECT DATE_FORMAT(TIMESTAMP(o.tgl_perawatan,o.jam_rawat),'%Y-%m-%d %H:%i:%s') waktu,COALESCE(o.tensi,'') td,COALESCE(o.nadi,'') nadi,COALESCE(o.respirasi,'') respirasi,COALESCE(o.suhu_tubuh,'') suhu,COALESCE(o.spo2,'') spo2,COALESCE(o.gcs,'') gcs,COALESCE(o.kesadaran,'') kesadaran,COALESCE(o.instruksi,'') instruksi,COALESCE(o.penilaian,'') penilaian,o.nip petugas FROM pemeriksaan_ranap o WHERE o.no_rawat=? AND TIMESTAMP(o.tgl_perawatan,o.jam_rawat)>=? AND TIMESTAMP(o.tgl_perawatan,o.jam_rawat)<? ORDER BY o.tgl_perawatan,o.jam_rawat`},
	{"resep", "Resep Dokter (bukan bukti pemberian)", `SELECT DATE_FORMAT(TIMESTAMP(o.tgl_peresepan,o.jam_peresepan),'%Y-%m-%d %H:%i:%s') waktu,o.no_resep, d.kode_brng,COALESCE(b.nama_brng,d.kode_brng) obat,COALESCE(CAST(d.jml AS CHAR),'') jumlah,COALESCE(b.kode_sat,'') satuan,COALESCE(d.aturan_pakai,'') aturan_pakai,o.kd_dokter dokter FROM resep_obat o JOIN resep_dokter d ON d.no_resep=o.no_resep LEFT JOIN databarang b ON b.kode_brng=d.kode_brng WHERE o.no_rawat=? AND o.status='ranap' AND TIMESTAMP(o.tgl_peresepan,o.jam_peresepan)>=? AND TIMESTAMP(o.tgl_peresepan,o.jam_peresepan)<? ORDER BY o.tgl_peresepan,o.jam_peresepan,o.no_resep,d.kode_brng`},
	{"obat", "Transaksi Obat (waktu transaksi, bukan konfirmasi pemberian)", `SELECT DATE_FORMAT(TIMESTAMP(o.tgl_perawatan,o.jam),'%Y-%m-%d %H:%i:%s') waktu,o.kode_brng,COALESCE(b.nama_brng,o.kode_brng) obat,CAST(o.jml AS CHAR) jumlah,COALESCE(b.kode_sat,'') satuan,COALESCE(o.kd_bangsal,'') bangsal,o.no_batch,o.no_faktur FROM detail_pemberian_obat o LEFT JOIN databarang b ON b.kode_brng=o.kode_brng WHERE o.no_rawat=? AND o.status='Ranap' AND TIMESTAMP(o.tgl_perawatan,o.jam)>=? AND TIMESTAMP(o.tgl_perawatan,o.jam)<? ORDER BY o.tgl_perawatan,o.jam,o.kode_brng,o.no_batch,o.no_faktur`},
	{"instruksi", "Instruksi SBAR", `SELECT DATE_FORMAT(TIMESTAMP(o.tgl_perawatan,o.jam_rawat),'%Y-%m-%d %H:%i:%s') waktu,o.instruksi,o.nip petugas,COALESCE(DATE_FORMAT(TIMESTAMP(o.tgl_validasi,o.jam_validasi),'%Y-%m-%d %H:%i:%s'),'') waktu_validasi FROM sbar_instruksi o WHERE o.no_rawat=? AND TIMESTAMP(o.tgl_perawatan,o.jam_rawat)>=? AND TIMESTAMP(o.tgl_perawatan,o.jam_rawat)<? ORDER BY o.tgl_perawatan,o.jam_rawat`},
	{"catatan", "Catatan Perawatan Dokter", `SELECT DATE_FORMAT(TIMESTAMP(o.tanggal,o.jam),'%Y-%m-%d %H:%i:%s') waktu,COALESCE(o.catatan,'') catatan,COALESCE(o.kd_dokter,'') dokter FROM catatan_perawatan o WHERE o.no_rawat=? AND TIMESTAMP(o.tanggal,o.jam)>=? AND TIMESTAMP(o.tanggal,o.jam)<? ORDER BY o.tanggal,o.jam`},
	{"lab", "Hasil Laboratorium", `SELECT DATE_FORMAT(TIMESTAMP(o.tgl_periksa,o.jam),'%Y-%m-%d %H:%i:%s') waktu,o.kd_jenis_prw,CAST(o.id_template AS CHAR) id_template,COALESCE(t.Pemeriksaan,'') pemeriksaan,o.nilai,COALESCE(t.satuan,'') satuan,o.nilai_rujukan,o.keterangan FROM detail_periksa_lab o LEFT JOIN template_laboratorium t ON t.id_template=o.id_template WHERE o.no_rawat=? AND TIMESTAMP(o.tgl_periksa,o.jam)>=? AND TIMESTAMP(o.tgl_periksa,o.jam)<? ORDER BY o.tgl_periksa,o.jam,o.kd_jenis_prw,o.id_template`},
	{"radiologi", "Hasil Radiologi", `SELECT DATE_FORMAT(TIMESTAMP(o.tgl_periksa,o.jam),'%Y-%m-%d %H:%i:%s') waktu,o.hasil,COALESCE(DATE_FORMAT(TIMESTAMP(o.tgl_hasil,o.jam_hasil),'%Y-%m-%d %H:%i:%s'),'') waktu_hasil FROM hasil_radiologi o WHERE o.no_rawat=? AND TIMESTAMP(o.tgl_periksa,o.jam)>=? AND TIMESTAMP(o.tgl_periksa,o.jam)<? ORDER BY o.tgl_periksa,o.jam`},
}

func (r *Repositori) Daftar(ctx context.Context, no, tanggal, jam string) (Hasil, error) {
	h := Hasil{Bagian: []Bagian{}}
	awal, akhir, err := Periode(no, tanggal, jam)
	if err != nil {
		return h, err
	}
	h.Mulai = awal.Format("2006-01-02 15:04:05")
	h.Selesai = akhir.Format("2006-01-02 15:04:05")
	for _, s := range sumber {
		b := Bagian{Kode: s.kode, Nama: s.nama, Baris: []map[string]string{}}
		rows, err := r.baca(ctx, s.query, no, h.Mulai, h.Selesai)
		if err != nil {
			b.Error = "Sumber belum dapat dibaca. Periksa koneksi, tabel, dan izin SELECT SIMRS."
			if errors.Is(err, errBatas) {
				b.Error = "Sumber memiliki lebih dari 2000 catatan pada periode ini. Data tidak ditampilkan sebagian; hubungi administrator."
			}
		} else {
			b.Baris = rows
		}
		h.Bagian = append(h.Bagian, b)
	}
	return h, nil
}
func (r *Repositori) baca(ctx context.Context, query, no, awal, akhir string) ([]map[string]string, error) {
	rows, err := r.simrs.QueryContext(ctx, query+" LIMIT 2001", no, awal, akhir)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	kolom, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	hasil := []map[string]string{}
	for rows.Next() {
		nilai := make([]sql.NullString, len(kolom))
		dest := make([]any, len(kolom))
		for i := range nilai {
			dest[i] = &nilai[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		b := map[string]string{}
		for i, k := range kolom {
			b[k] = nilai[i].String
		}
		hasil = append(hasil, b)
		if len(hasil) > 2000 {
			return nil, errBatas
		}
	}
	return hasil, rows.Err()
}
