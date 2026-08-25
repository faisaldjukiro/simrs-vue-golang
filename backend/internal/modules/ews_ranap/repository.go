package ews_ranap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrCatatanTidakDitemukan = errors.New("catatan EWS Ranap tidak ditemukan")
	ErrTidakBerhak           = errors.New("catatan hanya dapat diubah atau dihapus oleh petugas yang membuatnya")
)

type Petugas struct {
	NIP     string `json:"nip"`
	Nama    string `json:"nama"`
	Jabatan string `json:"jabatan"`
}

type Catatan struct {
	NoRawat        string `json:"no_rawat"`
	Tanggal        string `json:"tanggal"`
	Jam            string `json:"jam"`
	NIP            string `json:"nip"`
	NamaPetugas    string `json:"nama_petugas"`
	Jabatan        string `json:"jabatan"`
	Pernafasan     string `json:"pernafasan"`
	SkorPernafasan string `json:"score_pernafasan"`
	Saturasi       string `json:"saturasi"`
	SkorSaturasi   string `json:"score_saturasi"`
	Alat           string `json:"alat"`
	SkorAlat       string `json:"score_alat"`
	Suhu           string `json:"suhu"`
	SkorSuhu       string `json:"score_suhu"`
	Denyut         string `json:"denyut"`
	SkorDenyut     string `json:"score_denyut"`
	Tekanan        string `json:"tekanan"`
	Diastol        string `json:"diastol"`
	SkorTekanan    string `json:"score_tekanan"`
	Kesadaran      string `json:"kesadaran"`
	SkorKesadaran  string `json:"score_kesadaran"`
	TotalSkor      string `json:"total_score"`
	Klasifikasi    string `json:"klasifikasi"`
	Respon         string `json:"respon"`
	Tindakan       string `json:"tindakan"`
	Frekuensi      string `json:"frekuensi"`
	SkalaNyeri     string `json:"skala_nyeri"`
	BB             string `json:"bb"`
	TB             string `json:"tb"`
	LK             string `json:"lk"`
	LP             string `json:"lp"`
	Masuk1         string `json:"masuk1"`
	Masuk2         string `json:"masuk2"`
	JumlahMasuk    string `json:"jumlahmasuk"`
	Keluar1        string `json:"keluar1"`
	Keluar2        string `json:"keluar2"`
	Keluar3        string `json:"keluar3"`
	Keluar4        string `json:"keluar4"`
	Keluar5        string `json:"keluar5"`
	JumlahKeluar   string `json:"jumlahkeluar"`
	BC             string `json:"bc"`
	BisaDiubah     bool   `json:"bisa_diubah"`
}

type Kunci struct {
	NoRawat string `json:"no_rawat"`
	Tanggal string `json:"tanggal"`
	Jam     string `json:"jam"`
}

type Repositori struct {
	aplikasiDB *sql.DB
	simrsDB    *sql.DB
}

func NewRepositori(aplikasiDB, simrsDB *sql.DB) *Repositori {
	return &Repositori{aplikasiDB: aplikasiDB, simrsDB: simrsDB}
}

func (r *Repositori) BillingTerkunci(ctx context.Context, noRawat string) (bool, error) {
	var jumlah int
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM billing WHERE no_rawat = ?) +
			(SELECT COUNT(*) FROM reg_periksa WHERE no_rawat = ? AND stts = 'Batal')
	`, strings.TrimSpace(noRawat), strings.TrimSpace(noRawat)).Scan(&jumlah)
	if err != nil {
		return false, fmt.Errorf("periksa status billing EWS Ranap: %w", err)
	}
	return jumlah > 0, nil
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
		return nil, fmt.Errorf("cari petugas EWS Ranap: %w", err)
	}
	defer rows.Close()

	daftar := make([]Petugas, 0)
	for rows.Next() {
		var item Petugas
		if err := rows.Scan(&item.NIP, &item.Nama, &item.Jabatan); err != nil {
			return nil, fmt.Errorf("scan petugas EWS Ranap: %w", err)
		}
		daftar = append(daftar, item)
	}
	return daftar, rows.Err()
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
			e.no_rawat, DATE_FORMAT(e.tanggal, '%Y-%m-%d'), TIME_FORMAT(e.jam, '%H:%i:%s'),
			e.nik, COALESCE(p.nama, ''), COALESCE(p.jbtn, ''),
			COALESCE(e.pernafasan, ''), COALESCE(e.score_pernafasan, ''),
			COALESCE(e.saturasi, ''), COALESCE(e.score_saturasi, ''),
			COALESCE(e.alat, ''), COALESCE(e.score_alat, ''),
			COALESCE(e.suhu, ''), COALESCE(e.score_suhu, ''),
			COALESCE(e.denyut, ''), COALESCE(e.score_denyut, ''),
			COALESCE(e.tekanan, ''), COALESCE(e.diastol, ''), COALESCE(e.score_tekanan, ''),
			COALESCE(e.kesadaran, ''), COALESCE(e.score_kesadaran, ''),
			COALESCE(e.total_score, ''), COALESCE(e.klasifikasi, ''),
			COALESCE(e.respon, ''), COALESCE(e.tindakan, ''), COALESCE(e.frekuensi, ''),
			COALESCE(e.skala_nyeri, ''), COALESCE(e.bb, ''), COALESCE(e.tb, ''),
			COALESCE(e.lk, ''), COALESCE(e.lp, ''), COALESCE(e.masuk1, ''),
			COALESCE(e.masuk2, ''), COALESCE(e.jumlahmasuk, ''), COALESCE(e.keluar1, ''),
			COALESCE(e.keluar2, ''), COALESCE(e.keluar3, ''), COALESCE(e.keluar4, ''),
			COALESCE(e.keluar5, ''), COALESCE(e.jumlahkeluar, ''), COALESCE(e.bc, '')
		FROM ews_ranap e
		LEFT JOIN pegawai p ON p.nik = e.nik
		WHERE e.no_rawat = ?
		ORDER BY e.tanggal DESC, e.jam DESC
	`, noRawat)
	if err != nil {
		return nil, fmt.Errorf("baca daftar EWS Ranap: %w", err)
	}
	defer rows.Close()

	daftar := make([]Catatan, 0)
	for rows.Next() {
		var item Catatan
		err := rows.Scan(
			&item.NoRawat, &item.Tanggal, &item.Jam, &item.NIP, &item.NamaPetugas, &item.Jabatan,
			&item.Pernafasan, &item.SkorPernafasan, &item.Saturasi, &item.SkorSaturasi,
			&item.Alat, &item.SkorAlat, &item.Suhu, &item.SkorSuhu, &item.Denyut, &item.SkorDenyut,
			&item.Tekanan, &item.Diastol, &item.SkorTekanan, &item.Kesadaran, &item.SkorKesadaran,
			&item.TotalSkor, &item.Klasifikasi, &item.Respon, &item.Tindakan, &item.Frekuensi,
			&item.SkalaNyeri, &item.BB, &item.TB, &item.LK, &item.LP, &item.Masuk1, &item.Masuk2,
			&item.JumlahMasuk, &item.Keluar1, &item.Keluar2, &item.Keluar3, &item.Keluar4,
			&item.Keluar5, &item.JumlahKeluar, &item.BC,
		)
		if err != nil {
			return nil, fmt.Errorf("scan daftar EWS Ranap: %w", err)
		}
		item.BisaDiubah = aksesPenuh || item.NIP == nipLogin
		daftar = append(daftar, item)
	}
	return daftar, rows.Err()
}

func (r *Repositori) Simpan(ctx context.Context, catatan Catatan) error {
	_, err := r.simrsDB.ExecContext(ctx, `
		INSERT INTO ews_ranap (
			no_rawat, tanggal, jam, nik, pernafasan, score_pernafasan, saturasi, score_saturasi,
			alat, score_alat, suhu, score_suhu, denyut, score_denyut, tekanan, score_tekanan,
			kesadaran, score_kesadaran, total_score, klasifikasi, respon, tindakan, frekuensi,
			skala_nyeri, bb, tb, lk, lp, masuk1, masuk2, jumlahmasuk, keluar1, keluar2,
			keluar3, keluar4, keluar5, jumlahkeluar, bc, diastol
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, catatan.NoRawat, catatan.Tanggal, catatan.Jam, catatan.NIP, catatan.Pernafasan,
		catatan.SkorPernafasan, catatan.Saturasi, catatan.SkorSaturasi, catatan.Alat,
		catatan.SkorAlat, catatan.Suhu, catatan.SkorSuhu, catatan.Denyut, catatan.SkorDenyut,
		catatan.Tekanan, catatan.SkorTekanan, catatan.Kesadaran, catatan.SkorKesadaran,
		catatan.TotalSkor, catatan.Klasifikasi, catatan.Respon, catatan.Tindakan,
		catatan.Frekuensi, catatan.SkalaNyeri, catatan.BB, catatan.TB, catatan.LK, catatan.LP,
		catatan.Masuk1, catatan.Masuk2, catatan.JumlahMasuk, catatan.Keluar1, catatan.Keluar2,
		catatan.Keluar3, catatan.Keluar4, catatan.Keluar5, catatan.JumlahKeluar, catatan.BC,
		catatan.Diastol)
	if err != nil {
		return fmt.Errorf("simpan EWS Ranap: %w", err)
	}
	return nil
}

func (r *Repositori) Ubah(ctx context.Context, kunci Kunci, catatan Catatan, nipLogin string, aksesPenuh bool) error {
	return r.dalamTransaksiDenganHak(ctx, kunci, nipLogin, aksesPenuh, func(tx *sql.Tx, nipLama string) error {
		nipBaru := nipLama
		if aksesPenuh && strings.TrimSpace(catatan.NIP) != "" {
			nipBaru = catatan.NIP
		}
		_, err := tx.ExecContext(ctx, `
			UPDATE ews_ranap SET
				no_rawat = ?, tanggal = ?, jam = ?, nik = ?, pernafasan = ?, score_pernafasan = ?,
				saturasi = ?, score_saturasi = ?, alat = ?, score_alat = ?, suhu = ?,
				score_suhu = ?, denyut = ?, score_denyut = ?, tekanan = ?, score_tekanan = ?,
				kesadaran = ?, score_kesadaran = ?, total_score = ?, klasifikasi = ?, respon = ?,
				tindakan = ?, frekuensi = ?, skala_nyeri = ?, bb = ?, tb = ?, lk = ?, lp = ?,
				masuk1 = ?, masuk2 = ?, jumlahmasuk = ?, keluar1 = ?, keluar2 = ?, keluar3 = ?,
				keluar4 = ?, keluar5 = ?, jumlahkeluar = ?, bc = ?, diastol = ?
			WHERE no_rawat = ? AND tanggal = ? AND jam = ?
		`, catatan.NoRawat, catatan.Tanggal, catatan.Jam, nipBaru, catatan.Pernafasan,
			catatan.SkorPernafasan, catatan.Saturasi, catatan.SkorSaturasi, catatan.Alat,
			catatan.SkorAlat, catatan.Suhu, catatan.SkorSuhu, catatan.Denyut, catatan.SkorDenyut,
			catatan.Tekanan, catatan.SkorTekanan, catatan.Kesadaran, catatan.SkorKesadaran,
			catatan.TotalSkor, catatan.Klasifikasi, catatan.Respon, catatan.Tindakan,
			catatan.Frekuensi, catatan.SkalaNyeri, catatan.BB, catatan.TB, catatan.LK, catatan.LP,
			catatan.Masuk1, catatan.Masuk2, catatan.JumlahMasuk, catatan.Keluar1, catatan.Keluar2,
			catatan.Keluar3, catatan.Keluar4, catatan.Keluar5, catatan.JumlahKeluar, catatan.BC,
			catatan.Diastol, kunci.NoRawat, kunci.Tanggal, kunci.Jam)
		if err != nil {
			return fmt.Errorf("ubah EWS Ranap: %w", err)
		}
		return nil
	})
}

func (r *Repositori) Hapus(ctx context.Context, kunci Kunci, nipLogin string, aksesPenuh bool) error {
	return r.dalamTransaksiDenganHak(ctx, kunci, nipLogin, aksesPenuh, func(tx *sql.Tx, _ string) error {
		hasil, err := tx.ExecContext(ctx, `
			DELETE FROM ews_ranap
			WHERE no_rawat = ? AND tanggal = ? AND jam = ?
		`, kunci.NoRawat, kunci.Tanggal, kunci.Jam)
		if err != nil {
			return fmt.Errorf("hapus EWS Ranap: %w", err)
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
		SELECT nik FROM ews_ranap
		WHERE no_rawat = ? AND tanggal = ? AND jam = ?
		FOR UPDATE
	`, kunci.NoRawat, kunci.Tanggal, kunci.Jam).Scan(&pemilik)
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
