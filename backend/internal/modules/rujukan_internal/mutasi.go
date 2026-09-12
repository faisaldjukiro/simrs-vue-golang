package rujukan_internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

type pembaca interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Mutasi struct {
	Asal Rujukan `json:"asal"`
	Baru Input   `json:"baru"`
}

const PesanArsipGagal = "Rujukan sudah tersimpan di Khanza, tetapi pengarsipan lokal gagal. Muat ulang dan kirim ulang untuk menyelesaikan pengarsipan."

func kunciTujuan(jenis string, input Input) string {
	if jenis == Ranap {
		return input.KodePoli
	}
	return input.KodeDokter
}

func targetKhanza(jenis string) (string, string) {
	if jenis == Ranap {
		return "rujukan_internal_ranap", "kd_poli"
	}
	return "rujukan_internal_poli", "kd_dokter"
}

func periksaKunjungan(ctx context.Context, db pembaca, noRawat, jenis string, kunci bool) error {
	query := `SELECT stts, status_lanjut, status_bayar,
		(SELECT COUNT(*) FROM billing WHERE billing.no_rawat = reg_periksa.no_rawat)
		FROM reg_periksa WHERE no_rawat = ?`
	if kunci {
		query += " FOR UPDATE"
	}
	var status, lanjut, bayar string
	var billing int
	err := db.QueryRowContext(ctx, query, noRawat).Scan(&status, &lanjut, &bayar, &billing)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTidakAda
	}
	if err != nil {
		return err
	}
	if status == "Batal" || lanjut != jenis {
		return ErrKunjungan
	}
	if bayar == "Sudah Bayar" || billing > 0 {
		return ErrBilling
	}
	return nil
}

func periksaReferensi(ctx context.Context, db pembaca, input Input) (string, string, error) {
	var dokter, poli string
	err := db.QueryRowContext(ctx, `SELECT nm_dokter FROM dokter WHERE kd_dokter = ? AND status = '1'`, input.KodeDokter).Scan(&dokter)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", fmt.Errorf("%w: dokter tujuan tidak aktif atau tidak ditemukan", ErrInput)
	}
	if err != nil {
		return "", "", err
	}
	err = db.QueryRowContext(ctx, `SELECT nm_poli FROM poliklinik WHERE kd_poli = ? AND status = '1'`, input.KodePoli).Scan(&poli)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", fmt.Errorf("%w: poli tujuan tidak aktif atau tidak ditemukan", ErrInput)
	}
	return dokter, poli, err
}

func errorSQL(err error) error {
	var mysqlError *mysql.MySQLError
	if errors.As(err, &mysqlError) && mysqlError.Number == 1062 {
		return ErrDuplikat
	}
	return err
}

func insertKhanza(ctx context.Context, tx *sql.Tx, jenis string, input Input) error {
	var err error
	if jenis == Ranap {
		_, err = tx.ExecContext(ctx, `INSERT INTO rujukan_internal_ranap (no_rawat, kd_dokter, kd_poli, tanggal, jam) VALUES (?, ?, ?, ?, ?)`,
			input.NoRawat, input.KodeDokter, input.KodePoli, input.Tanggal, input.Jam)
	} else {
		_, err = tx.ExecContext(ctx, `INSERT INTO rujukan_internal_poli (no_rawat, kd_dokter, kd_poli) VALUES (?, ?, ?)`, input.NoRawat, input.KodeDokter, input.KodePoli)
	}
	return errorSQL(err)
}

func (r *Repositori) Simpan(ctx context.Context, jenis string, input Input) error {
	input, err := Validasi(jenis, input)
	if err != nil {
		return err
	}
	tx, err := r.simrsDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = periksaKunjungan(ctx, tx, input.NoRawat, jenis, true); err != nil {
		return err
	}
	if _, _, err = periksaReferensi(ctx, tx, input); err != nil {
		return err
	}
	if err = insertKhanza(ctx, tx, jenis, input); err != nil {
		return err
	}
	return tx.Commit()
}

func validAsal(jenis string, asal Rujukan) error {
	if !validJenis(jenis) || !validNoRawat(asal.NoRawat) || kunciTujuan(jenis, asal.Input) == "" ||
		(asal.Sumber != "Khanza" && asal.Sumber != "SIRAPI") || (asal.Sumber == "SIRAPI" && asal.ID == 0) {
		return ErrInput
	}
	return nil
}

func bacaAsal(ctx context.Context, tx *sql.Tx, jenis string, asal Rujukan) (Input, error) {
	input := Input{NoRawat: asal.NoRawat}
	var query string
	var args []any
	if asal.Sumber == "SIRAPI" {
		query = `SELECT kd_dokter, kd_poli, COALESCE(DATE_FORMAT(tanggal,'%Y-%m-%d'),''), COALESCE(TIME_FORMAT(jam,'%H:%i:%s'),'')
			FROM sirapi_rujukan_internal WHERE id = ? AND jenis_rawat = ? AND no_rawat = ? AND dikirim_pada IS NULL FOR UPDATE`
		args = []any{asal.ID, jenis, asal.NoRawat}
	} else {
		tabel, kolom := targetKhanza(jenis)
		waktu := "'', ''"
		if jenis == Ranap {
			waktu = "COALESCE(DATE_FORMAT(tanggal,'%Y-%m-%d'),''), COALESCE(TIME_FORMAT(jam,'%H:%i:%s'),'')"
		}
		// Nama tabel/kolom hanya berasal dari pilihan tetap di atas, bukan input pengguna.
		query = "SELECT kd_dokter, COALESCE(kd_poli,''), " + waktu + " FROM " + tabel + " WHERE no_rawat = ? AND " + kolom + " = ? FOR UPDATE"
		args = []any{asal.NoRawat, kunciTujuan(jenis, asal.Input)}
	}
	err := tx.QueryRowContext(ctx, query, args...).Scan(&input.KodeDokter, &input.KodePoli, &input.Tanggal, &input.Jam)
	return input, err
}

// Snapshot asal mencegah edit/hapus menimpa perubahan petugas lain.
func (r *Repositori) Mutasi(ctx context.Context, jenis string, asal Rujukan, baru *Input) error {
	if err := validAsal(jenis, asal); err != nil {
		return err
	}
	if baru != nil {
		valid, err := Validasi(jenis, *baru)
		if err != nil {
			return err
		}
		if valid.NoRawat != asal.NoRawat {
			return ErrInput
		}
		baru = &valid
	}
	db := r.simrsDB
	if asal.Sumber == "SIRAPI" {
		db = r.aplikasiDB
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var referensi pembaca = tx
	if asal.Sumber == "SIRAPI" {
		referensi = r.simrsDB
	}
	if err = periksaKunjungan(ctx, referensi, asal.NoRawat, jenis, asal.Sumber == "Khanza"); err != nil {
		return err
	}
	aktual, err := bacaAsal(ctx, tx, jenis, asal)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrBerubah
	}
	if err != nil {
		return err
	}
	if aktual != asal.Input {
		return ErrBerubah
	}
	var namaDokter, namaPoli string
	if baru != nil {
		namaDokter, namaPoli, err = periksaReferensi(ctx, referensi, *baru)
		if err != nil {
			return err
		}
	}
	if asal.Sumber == "SIRAPI" {
		if baru == nil {
			_, err = tx.ExecContext(ctx, `DELETE FROM sirapi_rujukan_internal WHERE id = ? AND jenis_rawat = ? AND no_rawat = ? AND dikirim_pada IS NULL`, asal.ID, jenis, asal.NoRawat)
		} else {
			_, err = tx.ExecContext(ctx, `UPDATE sirapi_rujukan_internal SET kd_dokter = ?, nama_dokter = ?, kd_poli = ?, nama_poli = ?, kunci_tujuan = ?, tanggal = NULLIF(?,''), jam = NULLIF(?,'')
				WHERE id = ? AND jenis_rawat = ? AND no_rawat = ? AND dikirim_pada IS NULL`, baru.KodeDokter, namaDokter, baru.KodePoli, namaPoli, kunciTujuan(jenis, *baru), baru.Tanggal, baru.Jam, asal.ID, jenis, asal.NoRawat)
		}
	} else {
		tabel, kolom := targetKhanza(jenis)
		where := " WHERE no_rawat = ? AND " + kolom + " = ?"
		if baru == nil {
			_, err = tx.ExecContext(ctx, "DELETE FROM "+tabel+where, asal.NoRawat, kunciTujuan(jenis, asal.Input))
		} else if jenis == Ranap {
			_, err = tx.ExecContext(ctx, "UPDATE "+tabel+" SET kd_dokter = ?, kd_poli = ?, tanggal = ?, jam = ?"+where,
				baru.KodeDokter, baru.KodePoli, baru.Tanggal, baru.Jam, asal.NoRawat, kunciTujuan(jenis, asal.Input))
		} else {
			_, err = tx.ExecContext(ctx, "UPDATE "+tabel+" SET kd_dokter = ?, kd_poli = ?"+where,
				baru.KodeDokter, baru.KodePoli, asal.NoRawat, kunciTujuan(jenis, asal.Input))
		}
	}
	if err != nil {
		return errorSQL(err)
	}
	return tx.Commit()
}

// Pengiriman bersifat per baris. Data lokal diarsipkan, bukan dihapus.
// Dua koneksi tidak atomik: jika pengarsipan gagal, retry hanya menerima isi Khanza yang identik.
func (r *Repositori) Kirim(ctx context.Context, jenis string, asal Rujukan) (string, error) {
	if err := validAsal(jenis, asal); err != nil {
		return "", err
	}
	if asal.Sumber != "SIRAPI" {
		return "", ErrInput
	}
	remote, err := r.simrsDB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer remote.Rollback()
	if err = periksaKunjungan(ctx, remote, asal.NoRawat, jenis, true); err != nil {
		return "", err
	}
	lokal, err := r.aplikasiDB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer lokal.Rollback()
	input, err := bacaAsal(ctx, lokal, jenis, asal)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrBerubah
	}
	if err != nil {
		return "", err
	}
	if input != asal.Input {
		return "", ErrBerubah
	}
	input, err = Validasi(jenis, input)
	if err != nil {
		return "", err
	}
	if _, _, err = periksaReferensi(ctx, remote, input); err != nil {
		return "", err
	}
	aktual, err := bacaAsal(ctx, remote, jenis, Rujukan{Input: input, Sumber: "Khanza"})
	if errors.Is(err, sql.ErrNoRows) {
		if err = insertKhanza(ctx, remote, jenis, input); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	} else if aktual != input {
		return "", ErrDuplikat
	}
	if err = remote.Commit(); err != nil {
		return "", err
	}
	_, err = lokal.ExecContext(ctx, `UPDATE sirapi_rujukan_internal SET dikirim_pada = CURRENT_TIMESTAMP WHERE id = ? AND jenis_rawat = ? AND no_rawat = ?`, asal.ID, jenis, asal.NoRawat)
	if err == nil {
		err = lokal.Commit()
	}
	if err != nil {
		return PesanArsipGagal, nil
	}
	return "Rujukan tersimpan di Khanza. Salinan lokal telah diarsipkan.", nil
}
