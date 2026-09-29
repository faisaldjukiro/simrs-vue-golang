// Package arsiplokal menangani tabel duplikat dari versi lama, tanpa membuat tabel baru.
package arsiplokal

import (
	"errors"
	"github.com/go-sql-driver/mysql"
)

func TabelTidakAda(err error) bool {
	var e *mysql.MySQLError
	return errors.As(err, &e) && e.Number == 1146
}
