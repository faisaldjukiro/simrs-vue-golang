// Package database contains database infrastructure shared by commands.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"

	"simrs-backend/internal/config"
)

func OpenMySQL(ctx context.Context, cfg config.Database) (*sql.DB, error) {
	db, err := openMySQL(cfg)
	if err != nil {
		return nil, err
	}

	pingContext, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingContext); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to application database: %w", err)
	}
	return db, nil
}

// OpenMySQLLazy creates a pool without pinging it. This lets the API remain
// available when the external SIMRS database is temporarily offline.
func OpenMySQLLazy(cfg config.Database) (*sql.DB, error) {
	return openMySQL(cfg)
}

func openMySQL(cfg config.Database) (*sql.DB, error) {
	driverConfig := mysql.Config{
		User:                 cfg.Username,
		Passwd:               cfg.Password,
		Net:                  "tcp",
		Addr:                 cfg.Host + ":" + cfg.Port,
		DBName:               cfg.Name,
		ParseTime:            true,
		Loc:                  time.Local,
		MultiStatements:      true,
		AllowNativePasswords: true,
		Params: map[string]string{
			"charset": "utf8mb4",
		},
	}

	db, err := sql.Open("mysql", driverConfig.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open MySQL: %w", err)
	}
	return db, nil
}
