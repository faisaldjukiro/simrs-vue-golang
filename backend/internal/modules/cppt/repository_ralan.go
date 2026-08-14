package cppt

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// DaftarRalan membaca tab Pemeriksaan pada DlgRawatJalan.java.
func (r *Repositori) DaftarRalan(ctx context.Context, noRawat, nipLogin string, aksesPenuh bool) ([]Catatan, error) {
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT p.no_rawat,DATE_FORMAT(p.tgl_perawatan,'%Y-%m-%d'),TIME_FORMAT(p.jam_rawat,'%H:%i:%s'),
			COALESCE(p.suhu_tubuh,''),COALESCE(p.tensi,''),COALESCE(p.nadi,''),
			COALESCE(p.respirasi,''),COALESCE(p.tinggi,''),COALESCE(p.berat,''),
			COALESCE(p.spo2,''),COALESCE(p.gcs,''),COALESCE(p.kesadaran,''),
			COALESCE(p.keluhan,''),COALESCE(p.pemeriksaan,''),COALESCE(p.alergi,''),
			COALESCE(p.lingkar_perut,''),COALESCE(p.penilaian,''),COALESCE(p.rtl,''),
			COALESCE(p.instruksi,''),COALESCE(p.evaluasi,''),p.nip,
			COALESCE(pegawai.nama,''),COALESCE(pegawai.jbtn,'')
		FROM pemeriksaan_ralan p
		LEFT JOIN pegawai ON pegawai.nik=p.nip
		WHERE p.no_rawat=?
		ORDER BY p.tgl_perawatan DESC,p.jam_rawat DESC
	`, noRawat)
	if err != nil {
		return nil, fmt.Errorf("baca daftar pemeriksaan IGD: %w", err)
	}
	defer rows.Close()

	catatan := make([]Catatan, 0)
	for rows.Next() {
		item := Catatan{JenisRawat: "ralan"}
		if err := rows.Scan(
			&item.NoRawat, &item.TanggalPerawatan, &item.JamRawat,
			&item.SuhuTubuh, &item.Tensi, &item.Nadi, &item.Respirasi, &item.Tinggi,
			&item.Berat, &item.SpO2, &item.GCS, &item.Kesadaran, &item.Subjek,
			&item.Objek, &item.Alergi, &item.LingkarPerut, &item.Asesmen, &item.Plan,
			&item.Instruksi, &item.Evaluasi, &item.NIP, &item.NamaPetugas, &item.Jabatan,
		); err != nil {
			return nil, fmt.Errorf("scan daftar pemeriksaan IGD: %w", err)
		}
		item.BisaDiubah = aksesPenuh || item.NIP == nipLogin
		catatan = append(catatan, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi daftar pemeriksaan IGD: %w", err)
	}
	return catatan, nil
}

func (r *Repositori) SimpanRalan(ctx context.Context, catatan Catatan) error {
	_, err := r.simrsDB.ExecContext(ctx, `
		INSERT INTO pemeriksaan_ralan (
			no_rawat,tgl_perawatan,jam_rawat,suhu_tubuh,tensi,nadi,respirasi,
			tinggi,berat,spo2,gcs,kesadaran,keluhan,pemeriksaan,alergi,
			lingkar_perut,rtl,penilaian,instruksi,evaluasi,nip
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	`, catatan.NoRawat, catatan.TanggalPerawatan, catatan.JamRawat, catatan.SuhuTubuh,
		catatan.Tensi, catatan.Nadi, catatan.Respirasi, catatan.Tinggi, catatan.Berat,
		catatan.SpO2, catatan.GCS, catatan.Kesadaran, catatan.Subjek, catatan.Objek,
		catatan.Alergi, catatan.LingkarPerut, catatan.Plan, catatan.Asesmen,
		catatan.Instruksi, catatan.Evaluasi, catatan.NIP)
	if err != nil {
		return fmt.Errorf("simpan pemeriksaan IGD: %w", err)
	}
	return nil
}

func (r *Repositori) UbahRalan(ctx context.Context, kunci Kunci, catatan Catatan, nipLogin string, aksesPenuh bool) error {
	return r.dalamTransaksiDenganHakRalan(ctx, kunci, nipLogin, aksesPenuh, func(tx *sql.Tx, nipLama string) error {
		nipBaru := nipLama
		if aksesPenuh && catatan.NIP != "" {
			nipBaru = catatan.NIP
		}
		hasil, err := tx.ExecContext(ctx, `
			UPDATE pemeriksaan_ralan SET
				no_rawat=?,tgl_perawatan=?,jam_rawat=?,suhu_tubuh=?,tensi=?,nadi=?,
				respirasi=?,tinggi=?,berat=?,spo2=?,gcs=?,kesadaran=?,keluhan=?,
				pemeriksaan=?,alergi=?,lingkar_perut=?,rtl=?,penilaian=?,instruksi=?,
				evaluasi=?,nip=?
			WHERE no_rawat=? AND tgl_perawatan=? AND jam_rawat=?
		`, catatan.NoRawat, catatan.TanggalPerawatan, catatan.JamRawat, catatan.SuhuTubuh,
			catatan.Tensi, catatan.Nadi, catatan.Respirasi, catatan.Tinggi, catatan.Berat,
			catatan.SpO2, catatan.GCS, catatan.Kesadaran, catatan.Subjek, catatan.Objek,
			catatan.Alergi, catatan.LingkarPerut, catatan.Plan, catatan.Asesmen,
			catatan.Instruksi, catatan.Evaluasi, nipBaru, kunci.NoRawat,
			kunci.TanggalPerawatan, kunci.JamRawat)
		if err != nil {
			return fmt.Errorf("ubah pemeriksaan IGD: %w", err)
		}
		jumlah, err := hasil.RowsAffected()
		if err == nil && jumlah == 0 {
			return ErrCatatanTidakDitemukan
		}
		return err
	})
}

func (r *Repositori) HapusRalan(ctx context.Context, kunci Kunci, nipLogin string, aksesPenuh bool) error {
	return r.dalamTransaksiDenganHakRalan(ctx, kunci, nipLogin, aksesPenuh, func(tx *sql.Tx, _ string) error {
		hasil, err := tx.ExecContext(ctx, `DELETE FROM pemeriksaan_ralan
			WHERE no_rawat=? AND tgl_perawatan=? AND jam_rawat=?`,
			kunci.NoRawat, kunci.TanggalPerawatan, kunci.JamRawat)
		if err != nil {
			return fmt.Errorf("hapus pemeriksaan IGD: %w", err)
		}
		jumlah, err := hasil.RowsAffected()
		if err == nil && jumlah == 0 {
			return ErrCatatanTidakDitemukan
		}
		return err
	})
}

func (r *Repositori) dalamTransaksiDenganHakRalan(
	ctx context.Context,
	kunci Kunci,
	nipLogin string,
	aksesPenuh bool,
	aksi func(*sql.Tx, string) error,
) error {
	tx, err := r.simrsDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var pemilik string
	err = tx.QueryRowContext(ctx, `SELECT nip FROM pemeriksaan_ralan
		WHERE no_rawat=? AND tgl_perawatan=? AND jam_rawat=? FOR UPDATE`,
		kunci.NoRawat, kunci.TanggalPerawatan, kunci.JamRawat).Scan(&pemilik)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCatatanTidakDitemukan
	}
	if err != nil {
		return err
	}
	if !aksesPenuh && pemilik != nipLogin {
		return ErrTidakBerhak
	}
	if err := aksi(tx, pemilik); err != nil {
		return err
	}
	return tx.Commit()
}
