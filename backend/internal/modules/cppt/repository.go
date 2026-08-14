package cppt

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrCatatanTidakDitemukan = errors.New("catatan CPPT tidak ditemukan")
	ErrTidakBerhak           = errors.New("catatan hanya dapat diubah atau dihapus oleh petugas yang membuatnya")
)

type Petugas struct {
	NIP     string `json:"nip"`
	Nama    string `json:"nama"`
	Jabatan string `json:"jabatan"`
}

type Catatan struct {
	JenisRawat       string `json:"jenis_rawat"`
	NoRawat          string `json:"no_rawat"`
	TanggalPerawatan string `json:"tgl_perawatan"`
	JamRawat         string `json:"jam_rawat"`
	SuhuTubuh        string `json:"suhu_tubuh"`
	Tensi            string `json:"tensi"`
	Nadi             string `json:"nadi"`
	Respirasi        string `json:"respirasi"`
	Tinggi           string `json:"tinggi"`
	Berat            string `json:"berat"`
	SpO2             string `json:"spo2"`
	GCS              string `json:"gcs"`
	Kesadaran        string `json:"kesadaran"`
	Subjek           string `json:"subjek"`
	Objek            string `json:"objek"`
	Alergi           string `json:"alergi"`
	LingkarPerut     string `json:"lingkar_perut,omitempty"`
	Asesmen          string `json:"asesmen"`
	Plan             string `json:"plan"`
	Instruksi        string `json:"instruksi"`
	Evaluasi         string `json:"evaluasi"`
	NIP              string `json:"nip"`
	NamaPetugas      string `json:"nama_petugas"`
	Jabatan          string `json:"jabatan"`
	BisaDiubah       bool   `json:"bisa_diubah"`
}

type Kunci struct {
	JenisRawat       string `json:"jenis_rawat"`
	NoRawat          string `json:"no_rawat"`
	TanggalPerawatan string `json:"tgl_perawatan"`
	JamRawat         string `json:"jam_rawat"`
}

type Repositori struct {
	aplikasiDB *sql.DB
	simrsDB    *sql.DB
}

func NewRepositori(aplikasiDB, simrsDB *sql.DB) *Repositori {
	return &Repositori{aplikasiDB: aplikasiDB, simrsDB: simrsDB}
}

func (r *Repositori) PetugasLogin(ctx context.Context, username string) (Petugas, error) {
	var petugas Petugas
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT nik, COALESCE(nama, ''), COALESCE(jbtn, '')
		FROM pegawai
		WHERE nik = ?
		LIMIT 1
	`, strings.TrimSpace(username)).Scan(&petugas.NIP, &petugas.Nama, &petugas.Jabatan)
	return petugas, err
}

func (r *Repositori) CariPetugas(ctx context.Context, kataKunci string) ([]Petugas, error) {
	kataKunci = strings.TrimSpace(kataKunci)
	if len(kataKunci) < 2 {
		return []Petugas{}, nil
	}

	seperti := "%" + kataKunci + "%"
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT nik, COALESCE(nama, ''), COALESCE(jbtn, '')
		FROM pegawai
		WHERE nik LIKE ? OR nama LIKE ? OR jbtn LIKE ?
		ORDER BY nama
		LIMIT 20
	`, seperti, seperti, seperti)
	if err != nil {
		return nil, fmt.Errorf("cari petugas SIMRS: %w", err)
	}
	defer rows.Close()

	petugas := make([]Petugas, 0)
	for rows.Next() {
		var item Petugas
		if err := rows.Scan(&item.NIP, &item.Nama, &item.Jabatan); err != nil {
			return nil, fmt.Errorf("scan petugas SIMRS: %w", err)
		}
		petugas = append(petugas, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi petugas SIMRS: %w", err)
	}
	return petugas, nil
}

func (r *Repositori) AksesPenuh(ctx context.Context, userID uint64) (bool, error) {
	var jumlah int
	err := r.aplikasiDB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM user_permissions
		INNER JOIN permissions ON permissions.id = user_permissions.permission_id
		WHERE user_permissions.user_id = ? AND permissions.code = '*'
	`, userID).Scan(&jumlah)
	return jumlah > 0, err
}

func (r *Repositori) Daftar(ctx context.Context, noRawat, nipLogin string, aksesPenuh bool) ([]Catatan, error) {
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT
			p.no_rawat,
			DATE_FORMAT(p.tgl_perawatan, '%Y-%m-%d'),
			TIME_FORMAT(p.jam_rawat, '%H:%i:%s'),
			COALESCE(p.suhu_tubuh, ''), COALESCE(p.tensi, ''), COALESCE(p.nadi, ''),
			COALESCE(p.respirasi, ''), COALESCE(p.tinggi, ''), COALESCE(p.berat, ''),
			COALESCE(p.spo2, ''), COALESCE(p.gcs, ''), COALESCE(p.kesadaran, ''),
			COALESCE(p.keluhan, ''), COALESCE(p.pemeriksaan, ''), COALESCE(p.alergi, ''),
			COALESCE(p.penilaian, ''), COALESCE(p.rtl, ''), COALESCE(p.instruksi, ''),
			COALESCE(p.evaluasi, ''), p.nip, COALESCE(pegawai.nama, ''), COALESCE(pegawai.jbtn, '')
		FROM pemeriksaan_ranap p
		LEFT JOIN pegawai ON pegawai.nik = p.nip
		WHERE p.no_rawat = ?
		ORDER BY p.tgl_perawatan DESC, p.jam_rawat DESC
	`, noRawat)
	if err != nil {
		return nil, fmt.Errorf("baca daftar CPPT: %w", err)
	}
	defer rows.Close()

	catatan := make([]Catatan, 0)
	for rows.Next() {
		item := Catatan{JenisRawat: "ranap"}
		if err := rows.Scan(
			&item.NoRawat, &item.TanggalPerawatan, &item.JamRawat,
			&item.SuhuTubuh, &item.Tensi, &item.Nadi, &item.Respirasi, &item.Tinggi,
			&item.Berat, &item.SpO2, &item.GCS, &item.Kesadaran, &item.Subjek,
			&item.Objek, &item.Alergi, &item.Asesmen, &item.Plan, &item.Instruksi,
			&item.Evaluasi, &item.NIP, &item.NamaPetugas, &item.Jabatan,
		); err != nil {
			return nil, fmt.Errorf("scan daftar CPPT: %w", err)
		}
		item.BisaDiubah = aksesPenuh || item.NIP == nipLogin
		catatan = append(catatan, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi daftar CPPT: %w", err)
	}
	return catatan, nil
}

func (r *Repositori) Simpan(ctx context.Context, catatan Catatan) error {
	_, err := r.simrsDB.ExecContext(ctx, `
		INSERT INTO pemeriksaan_ranap (
			no_rawat, tgl_perawatan, jam_rawat, suhu_tubuh, tensi, nadi, respirasi,
			tinggi, berat, spo2, gcs, kesadaran, keluhan, pemeriksaan, alergi,
			penilaian, rtl, instruksi, evaluasi, nip
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, catatan.NoRawat, catatan.TanggalPerawatan, catatan.JamRawat, catatan.SuhuTubuh,
		catatan.Tensi, catatan.Nadi, catatan.Respirasi, catatan.Tinggi, catatan.Berat,
		catatan.SpO2, catatan.GCS, catatan.Kesadaran, catatan.Subjek, catatan.Objek,
		catatan.Alergi, catatan.Asesmen, catatan.Plan, catatan.Instruksi, catatan.Evaluasi,
		catatan.NIP)
	if err != nil {
		return fmt.Errorf("simpan CPPT: %w", err)
	}
	return nil
}

func (r *Repositori) Ubah(ctx context.Context, kunci Kunci, catatan Catatan, nipLogin string, aksesPenuh bool) error {
	return r.dalamTransaksiDenganHak(ctx, kunci, nipLogin, aksesPenuh, func(tx *sql.Tx, nipLama string) error {
		nipBaru := nipLama
		if aksesPenuh && catatan.NIP != "" {
			nipBaru = catatan.NIP
		}
		hasil, err := tx.ExecContext(ctx, `
			UPDATE pemeriksaan_ranap SET
				no_rawat = ?, tgl_perawatan = ?, jam_rawat = ?, suhu_tubuh = ?, tensi = ?,
				nadi = ?, respirasi = ?, tinggi = ?, berat = ?, spo2 = ?, gcs = ?, kesadaran = ?,
				keluhan = ?, pemeriksaan = ?, alergi = ?, penilaian = ?, rtl = ?, instruksi = ?,
				evaluasi = ?, nip = ?
			WHERE no_rawat = ? AND tgl_perawatan = ? AND jam_rawat = ?
		`, catatan.NoRawat, catatan.TanggalPerawatan, catatan.JamRawat, catatan.SuhuTubuh,
			catatan.Tensi, catatan.Nadi, catatan.Respirasi, catatan.Tinggi, catatan.Berat,
			catatan.SpO2, catatan.GCS, catatan.Kesadaran, catatan.Subjek, catatan.Objek,
			catatan.Alergi, catatan.Asesmen, catatan.Plan, catatan.Instruksi, catatan.Evaluasi,
			nipBaru, kunci.NoRawat, kunci.TanggalPerawatan, kunci.JamRawat)
		if err != nil {
			return fmt.Errorf("ubah CPPT: %w", err)
		}
		jumlah, err := hasil.RowsAffected()
		if err != nil {
			return err
		}
		if jumlah == 0 {
			return ErrCatatanTidakDitemukan
		}
		return nil
	})
}

func (r *Repositori) Hapus(ctx context.Context, kunci Kunci, nipLogin string, aksesPenuh bool) error {
	return r.dalamTransaksiDenganHak(ctx, kunci, nipLogin, aksesPenuh, func(tx *sql.Tx, _ string) error {
		hasil, err := tx.ExecContext(ctx, `
			DELETE FROM pemeriksaan_ranap
			WHERE no_rawat = ? AND tgl_perawatan = ? AND jam_rawat = ?
		`, kunci.NoRawat, kunci.TanggalPerawatan, kunci.JamRawat)
		if err != nil {
			return fmt.Errorf("hapus CPPT: %w", err)
		}
		jumlah, err := hasil.RowsAffected()
		if err != nil {
			return err
		}
		if jumlah == 0 {
			return ErrCatatanTidakDitemukan
		}
		return nil
	})
}

func (r *Repositori) dalamTransaksiDenganHak(
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
	err = tx.QueryRowContext(ctx, `
		SELECT nip FROM pemeriksaan_ranap
		WHERE no_rawat = ? AND tgl_perawatan = ? AND jam_rawat = ?
		FOR UPDATE
	`, kunci.NoRawat, kunci.TanggalPerawatan, kunci.JamRawat).Scan(&pemilik)
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
