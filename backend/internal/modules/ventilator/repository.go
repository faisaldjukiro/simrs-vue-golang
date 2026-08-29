package ventilator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
)

type Repositori struct{ db *sql.DB }

func NewRepositori(db *sql.DB) *Repositori { return &Repositori{db: db} }

func (r *Repositori) DaftarMaster(ctx context.Context) ([]Master, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT kode_ventilator,nama_ventilator,COALESCE(merk,''),COALESCE(model,''),COALESCE(nomor_seri,''),COALESCE(ruangan,''),status,DATE_FORMAT(tanggal_pemeliharaan_terakhir,'%Y-%m-%d'),DATE_FORMAT(tanggal_pemeliharaan_berikutnya,'%Y-%m-%d'),COALESCE(keterangan,'') FROM master_ventilator ORDER BY nama_ventilator`)
	if err != nil {
		return nil, fmt.Errorf("baca master ventilator: %w", err)
	}
	defer rows.Close()
	data := []Master{}
	for rows.Next() {
		var x Master
		var a, b sql.NullString
		if err := rows.Scan(&x.KodeVentilator, &x.Nama, &x.Merk, &x.Model, &x.NomorSeri, &x.Ruangan, &x.Status, &a, &b, &x.Keterangan); err != nil {
			return nil, err
		}
		if a.Valid {
			x.TerakhirRawat = &a.String
		}
		if b.Valid {
			x.BerikutnyaRawat = &b.String
		}
		data = append(data, x)
	}
	return data, rows.Err()
}
func (r *Repositori) SimpanMaster(ctx context.Context, x Master) (string, error) {
	const jumlahPercobaan = 5
	for percobaan := 0; percobaan < jumlahPercobaan; percobaan++ {
		var nomorBerikutnya uint64
		err := r.db.QueryRowContext(ctx, `
			SELECT COALESCE(MAX(CAST(SUBSTRING(kode_ventilator, 6) AS UNSIGNED)), 0) + 1
			FROM master_ventilator
			WHERE kode_ventilator REGEXP '^VENT-[0-9]+$'
		`).Scan(&nomorBerikutnya)
		if err != nil {
			return "", fmt.Errorf("buat kode ventilator: %w", err)
		}

		x.KodeVentilator = fmt.Sprintf("VENT-%04d", nomorBerikutnya)
		_, err = r.db.ExecContext(ctx, `INSERT INTO master_ventilator(kode_ventilator,nama_ventilator,merk,model,nomor_seri,ruangan,status,tanggal_pemeliharaan_terakhir,tanggal_pemeliharaan_berikutnya,keterangan) VALUES(?,?,?,?,?,?,?,NULLIF(?,''),NULLIF(?,''),?)`, x.KodeVentilator, x.Nama, x.Merk, x.Model, x.NomorSeri, x.Ruangan, x.Status, nilai(x.TerakhirRawat), nilai(x.BerikutnyaRawat), x.Keterangan)
		if err == nil {
			return x.KodeVentilator, nil
		}
		var mysqlError *mysql.MySQLError
		if !errors.As(err, &mysqlError) || mysqlError.Number != 1062 {
			return "", err
		}
	}
	return "", fmt.Errorf("kode ventilator gagal dibuat setelah %d percobaan", jumlahPercobaan)
}
func (r *Repositori) UbahMaster(ctx context.Context, kode string, x Master) error {
	res, err := r.db.ExecContext(ctx, `UPDATE master_ventilator SET nama_ventilator=?,merk=?,model=?,nomor_seri=?,ruangan=?,status=?,tanggal_pemeliharaan_terakhir=NULLIF(?,''),tanggal_pemeliharaan_berikutnya=NULLIF(?,''),keterangan=? WHERE kode_ventilator=?`, x.Nama, x.Merk, x.Model, x.NomorSeri, x.Ruangan, x.Status, nilai(x.TerakhirRawat), nilai(x.BerikutnyaRawat), x.Keterangan, kode)
	if err == nil {
		n, _ := res.RowsAffected()
		if n == 0 {
			return sql.ErrNoRows
		}
	}
	return err
}
func (r *Repositori) HapusMaster(ctx context.Context, kode string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM master_ventilator WHERE kode_ventilator=?`, kode)
	return err
}
func nilai(x *string) string {
	if x == nil {
		return ""
	}
	return strings.TrimSpace(*x)
}

func (r *Repositori) DataPasien(ctx context.Context, noRawat string) (DataPasien, error) {
	master, err := r.DaftarMaster(ctx)
	if err != nil {
		return DataPasien{}, err
	}
	out := DataPasien{Master: master, Pemakaian: []Pemakaian{}, Setting: []Setting{}, Monitoring: []Monitoring{}, ChecklistVAP: []ChecklistVAP{}}
	rows, err := r.db.QueryContext(ctx, `SELECT p.id,p.no_rawat,p.kode_ventilator,m.nama_ventilator,DATE_FORMAT(p.tanggal_mulai,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(p.tanggal_selesai,'%Y-%m-%d %H:%i:%s'),COALESCE(p.ruangan,''),COALESCE(p.indikasi,''),COALESCE(p.jenis_jalan_napas,''),COALESCE(p.ukuran_jalan_napas,''),COALESCE(p.kedalaman_jalan_napas,''),p.status,COALESCE(p.dokter_penanggung_jawab,''),COALESCE(p.petugas_pemasangan,''),COALESCE(p.catatan,'') FROM pemakaian_ventilator p JOIN master_ventilator m ON m.kode_ventilator=p.kode_ventilator WHERE p.no_rawat=? ORDER BY p.tanggal_mulai DESC`, noRawat)
	if err != nil {
		return out, err
	}
	var ids []uint64
	var idPemakaianAktif uint64
	for rows.Next() {
		var x Pemakaian
		var selesai sql.NullString
		if err := rows.Scan(&x.ID, &x.NoRawat, &x.KodeVentilator, &x.NamaVentilator, &x.TanggalMulai, &selesai, &x.Ruangan, &x.Indikasi, &x.JenisJalanNapas, &x.UkuranJalanNapas, &x.KedalamanJalanNapas, &x.Status, &x.DokterPenanggungJawab, &x.PetugasPemasangan, &x.Catatan); err != nil {
			rows.Close()
			return out, err
		}
		if selesai.Valid {
			x.TanggalSelesai = &selesai.String
		}
		out.Pemakaian = append(out.Pemakaian, x)
		ids = append(ids, x.ID)
		if idPemakaianAktif == 0 && (strings.EqualFold(x.Status, "Aktif") || strings.EqualFold(x.Status, "Weaning")) {
			idPemakaianAktif = x.ID
		}
	}
	rows.Close()
	if len(ids) > 0 {
		id := idPemakaianAktif
		if id == 0 {
			id = ids[0]
		}
		out.Setting, err = r.daftarSetting(ctx, id)
		if err != nil {
			return out, fmt.Errorf("baca setting ventilator: %w", err)
		}
		out.Monitoring, err = r.daftarMonitoring(ctx, id)
		if err != nil {
			return out, fmt.Errorf("baca monitoring ventilator: %w", err)
		}
		out.ChecklistVAP, err = r.daftarVAP(ctx, id)
		if err != nil {
			return out, fmt.Errorf("baca checklist VAP: %w", err)
		}
	}
	var status string
	_ = r.db.QueryRowContext(ctx, `SELECT status_bayar FROM reg_periksa WHERE no_rawat=?`, noRawat).Scan(&status)
	out.BillingTerkunci = strings.EqualFold(status, "Sudah Bayar")
	return out, nil
}
func (r *Repositori) daftarSetting(ctx context.Context, id uint64) ([]Setting, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT id,id_pemakaian,DATE_FORMAT(waktu_setting,'%Y-%m-%d %H:%i:%s'),mode_ventilator,COALESCE(fio2_persen,0),COALESCE(peep_cmh2o,0),COALESCE(tidal_volume_ml,0),COALESCE(frekuensi_set,0),COALESCE(pressure_control_cmh2o,0),COALESCE(pressure_support_cmh2o,0),COALESCE(petugas,''),COALESCE(catatan,'') FROM setting_ventilator WHERE id_pemakaian=? ORDER BY waktu_setting DESC`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	d := []Setting{}
	for rows.Next() {
		var x Setting
		if e = rows.Scan(&x.ID, &x.IDPemakaian, &x.Waktu, &x.Mode, &x.FiO2, &x.PEEP, &x.TidalVolume, &x.Frekuensi, &x.PressureControl, &x.PressureSupport, &x.Petugas, &x.Catatan); e != nil {
			return nil, e
		}
		d = append(d, x)
	}
	return d, rows.Err()
}
func (r *Repositori) daftarMonitoring(ctx context.Context, id uint64) ([]Monitoring, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT id,id_pemakaian,DATE_FORMAT(waktu_monitoring,'%Y-%m-%d %H:%i:%s'),COALESCE(kesadaran,''),COALESCE(tekanan_darah,''),COALESCE(nadi_per_menit,0),COALESCE(respirasi_per_menit,0),COALESCE(suhu_celsius,0),COALESCE(spo2_persen,0),COALESCE(tahap,''),COALESCE(petugas,''),COALESCE(catatan,'') FROM monitoring_ventilator WHERE id_pemakaian=? ORDER BY waktu_monitoring DESC`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	d := []Monitoring{}
	for rows.Next() {
		var x Monitoring
		if e = rows.Scan(&x.ID, &x.IDPemakaian, &x.Waktu, &x.Kesadaran, &x.TekananDarah, &x.Nadi, &x.Respirasi, &x.Suhu, &x.SpO2, &x.Tahap, &x.Petugas, &x.Catatan); e != nil {
			return nil, e
		}
		d = append(d, x)
	}
	return d, rows.Err()
}
func (r *Repositori) daftarVAP(ctx context.Context, id uint64) ([]ChecklistVAP, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT id,id_pemakaian,DATE_FORMAT(waktu_pemeriksaan,'%Y-%m-%d %H:%i:%s'),elevasi_kepala_30_derajat,perawatan_mulut,suction_sekret,evaluasi_sedasi,spontaneous_awakening_trial,spontaneous_breathing_trial,profilaksis_dvt,profilaksis_ulkus_stres,COALESCE(tekanan_cuff_cmh2o,0),COALESCE(petugas,''),COALESCE(catatan,'') FROM checklist_vap WHERE id_pemakaian=? ORDER BY waktu_pemeriksaan DESC`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	d := []ChecklistVAP{}
	for rows.Next() {
		var x ChecklistVAP
		if e = rows.Scan(&x.ID, &x.IDPemakaian, &x.Waktu, &x.ElevasiKepala, &x.PerawatanMulut, &x.Suction, &x.EvaluasiSedasi, &x.SAT, &x.SBT, &x.PencegahanDVT, &x.PencegahanUlkus, &x.TekananCuff, &x.Petugas, &x.Catatan); e != nil {
			return nil, e
		}
		d = append(d, x)
	}
	return d, rows.Err()
}

func (r *Repositori) SimpanPemakaian(ctx context.Context, x Pemakaian) error {
	tx, e := r.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	res, e := tx.ExecContext(ctx, `INSERT INTO pemakaian_ventilator(no_rawat,kode_ventilator,tanggal_mulai,ruangan,indikasi,jenis_jalan_napas,ukuran_jalan_napas,kedalaman_jalan_napas,status,dokter_penanggung_jawab,petugas_pemasangan,catatan) VALUES(?,?,?,?,?,?,?,NULLIF(?,''),?,?,?,?)`, x.NoRawat, x.KodeVentilator, x.TanggalMulai, x.Ruangan, x.Indikasi, x.JenisJalanNapas, x.UkuranJalanNapas, x.KedalamanJalanNapas, "Aktif", x.DokterPenanggungJawab, x.PetugasPemasangan, x.Catatan)
	if e != nil {
		return e
	}
	if _, e = res.LastInsertId(); e != nil {
		return e
	}
	_, e = tx.ExecContext(ctx, `UPDATE master_ventilator SET status='Digunakan' WHERE kode_ventilator=?`, x.KodeVentilator)
	if e != nil {
		return e
	}
	return tx.Commit()
}
func (r *Repositori) SimpanSetting(ctx context.Context, x Setting) error {
	_, e := r.db.ExecContext(ctx, `INSERT INTO setting_ventilator(id_pemakaian,waktu_setting,mode_ventilator,fio2_persen,peep_cmh2o,tidal_volume_ml,frekuensi_set,pressure_control_cmh2o,pressure_support_cmh2o,petugas,catatan) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, x.IDPemakaian, x.Waktu, x.Mode, x.FiO2, x.PEEP, x.TidalVolume, x.Frekuensi, x.PressureControl, x.PressureSupport, x.Petugas, x.Catatan)
	return e
}
func (r *Repositori) SimpanMonitoring(ctx context.Context, x Monitoring) error {
	_, e := r.db.ExecContext(ctx, `INSERT INTO monitoring_ventilator(id_pemakaian,waktu_monitoring,kesadaran,tekanan_darah,nadi_per_menit,respirasi_per_menit,suhu_celsius,spo2_persen,tahap,petugas,catatan) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, x.IDPemakaian, x.Waktu, x.Kesadaran, x.TekananDarah, x.Nadi, x.Respirasi, x.Suhu, x.SpO2, x.Tahap, x.Petugas, x.Catatan)
	return e
}
func (r *Repositori) SimpanVAP(ctx context.Context, x ChecklistVAP) error {
	_, e := r.db.ExecContext(ctx, `INSERT INTO checklist_vap(id_pemakaian,waktu_pemeriksaan,elevasi_kepala_30_derajat,perawatan_mulut,suction_sekret,evaluasi_sedasi,spontaneous_awakening_trial,spontaneous_breathing_trial,profilaksis_dvt,profilaksis_ulkus_stres,tekanan_cuff_cmh2o,petugas,catatan) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, x.IDPemakaian, x.Waktu, x.ElevasiKepala, x.PerawatanMulut, x.Suction, x.EvaluasiSedasi, x.SAT, x.SBT, x.PencegahanDVT, x.PencegahanUlkus, x.TekananCuff, x.Petugas, x.Catatan)
	return e
}
