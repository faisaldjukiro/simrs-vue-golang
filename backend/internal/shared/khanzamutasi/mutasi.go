// Package khanzamutasi membatasi mutasi ke satu baris dengan snapshot asli.
package khanzamutasi

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var ErrKonflik = errors.New("data berubah, hilang, atau kunci tidak unik; muat ulang riwayat")

// Kolom/tabel wajib konstanta server, tidak boleh berasal dari request.
func Kondisi(kolom []string, nilai []any) (string, []any) {
	syarat := make([]string, len(kolom))
	for i, k := range kolom {
		syarat[i] = "BINARY COALESCE(" + k + ",'') = BINARY ?"
	}
	return strings.Join(syarat, " AND "), append([]any(nil), nilai...)
}

func Jalankan(ctx context.Context, db *sql.DB, tabel string, kolom []string, lama, baru []any, hapus bool) error {
	if len(kolom) == 0 || len(lama) != len(kolom) || (!hapus && len(baru) != len(kolom)) {
		return ErrKonflik
	}
	where, args := Kondisi(kolom, lama)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, "SELECT 1 FROM "+tabel+" WHERE "+where+" LIMIT 2 FOR UPDATE", args...)
	if err != nil {
		return err
	}
	jumlah := 0
	for rows.Next() {
		jumlah++
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if jumlah != 1 {
		return ErrKonflik
	}
	if hapus {
		_, err = tx.ExecContext(ctx, "DELETE FROM "+tabel+" WHERE "+where+" LIMIT 1", args...)
	} else {
		set := make([]string, len(kolom))
		for i, k := range kolom {
			set[i] = k + "=?"
		}
		_, err = tx.ExecContext(ctx, "UPDATE "+tabel+" SET "+strings.Join(set, ",")+" WHERE "+where+" LIMIT 1", append(append([]any(nil), baru...), args...)...)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}
