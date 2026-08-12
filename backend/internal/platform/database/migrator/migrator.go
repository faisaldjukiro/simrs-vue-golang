// Package migrator applies versioned SQL migrations to the application DB.
package migrator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	migrationTable = "schema_migrations"
	lockName       = "simrs_backend_application_migrations"
)

var migrationFilePattern = regexp.MustCompile(`^(\d+)_(.+)\.(up|down)\.sql$`)

type Migration struct {
	Version  int64
	Name     string
	UpPath   string
	DownPath string
}

type Migrator struct {
	db        *sql.DB
	directory string
}

func New(db *sql.DB, directory string) *Migrator {
	return &Migrator{db: db, directory: directory}
}

func (m *Migrator) Up(ctx context.Context) error {
	return m.withLock(ctx, func() error {
		if err := m.ensureTable(ctx); err != nil {
			return err
		}

		migrations, err := m.discover()
		if err != nil {
			return err
		}
		applied, err := m.appliedVersions(ctx)
		if err != nil {
			return err
		}

		count := 0
		for _, migration := range migrations {
			if applied[migration.Version] {
				continue
			}
			if migration.UpPath == "" {
				return fmt.Errorf("migration %d tidak memiliki file up", migration.Version)
			}
			if err := m.applyUp(ctx, migration); err != nil {
				return err
			}
			fmt.Printf("applied %06d_%s\n", migration.Version, migration.Name)
			count++
		}

		if count == 0 {
			fmt.Println("no pending migrations")
		}
		return nil
	})
}

func (m *Migrator) Down(ctx context.Context) error {
	return m.withLock(ctx, func() error {
		if err := m.ensureTable(ctx); err != nil {
			return err
		}

		var version int64
		var name string
		err := m.db.QueryRowContext(ctx,
			"SELECT version, name FROM "+migrationTable+" ORDER BY version DESC LIMIT 1",
		).Scan(&version, &name)
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("no applied migrations")
			return nil
		}
		if err != nil {
			return fmt.Errorf("read last migration: %w", err)
		}

		migrations, err := m.discover()
		if err != nil {
			return err
		}
		for _, migration := range migrations {
			if migration.Version != version {
				continue
			}
			if migration.DownPath == "" {
				return fmt.Errorf("migration %d tidak memiliki file down", version)
			}
			if err := m.applyDown(ctx, migration); err != nil {
				return err
			}
			fmt.Printf("rolled back %06d_%s\n", version, name)
			return nil
		}
		return fmt.Errorf("file untuk migration %d tidak ditemukan", version)
	})
}

// Fresh removes only objects declared by known down migrations, then applies
// every migration again. It never drops the database itself.
func (m *Migrator) Fresh(ctx context.Context) error {
	return m.withLock(ctx, func() error {
		if err := m.ensureTable(ctx); err != nil {
			return err
		}

		migrations, err := m.discover()
		if err != nil {
			return err
		}

		for index := len(migrations) - 1; index >= 0; index-- {
			migration := migrations[index]
			if migration.DownPath == "" {
				return fmt.Errorf("migration %d tidak memiliki file down", migration.Version)
			}
			if err := m.applyDown(ctx, migration); err != nil {
				return err
			}
			fmt.Printf("cleared %06d_%s\n", migration.Version, migration.Name)
		}

		for _, migration := range migrations {
			if migration.UpPath == "" {
				return fmt.Errorf("migration %d tidak memiliki file up", migration.Version)
			}
			if err := m.applyUp(ctx, migration); err != nil {
				return err
			}
			fmt.Printf("applied %06d_%s\n", migration.Version, migration.Name)
		}
		return nil
	})
}

func (m *Migrator) Status(ctx context.Context) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	migrations, err := m.discover()
	if err != nil {
		return err
	}
	applied, err := m.appliedVersions(ctx)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		status := "pending"
		if applied[migration.Version] {
			status = "applied"
		}
		fmt.Printf("%-8s %06d_%s\n", status, migration.Version, migration.Name)
	}
	return nil
}

func (m *Migrator) ensureTable(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT NOT NULL,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (version)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	if err != nil {
		return fmt.Errorf("ensure migration table: %w", err)
	}
	return nil
}

func (m *Migrator) discover() ([]Migration, error) {
	entries, err := os.ReadDir(m.directory)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	byVersion := make(map[int64]*Migration)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := migrationFilePattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}
		version, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse migration version %q: %w", matches[1], err)
		}

		migration, exists := byVersion[version]
		if !exists {
			migration = &Migration{Version: version, Name: matches[2]}
			byVersion[version] = migration
		}
		if migration.Name != matches[2] {
			return nil, fmt.Errorf("migration version %d memiliki nama berbeda", version)
		}

		path := filepath.Join(m.directory, entry.Name())
		if matches[3] == "up" {
			migration.UpPath = path
		} else {
			migration.DownPath = path
		}
	}

	migrations := make([]Migration, 0, len(byVersion))
	for _, migration := range byVersion {
		migrations = append(migrations, *migration)
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
	return migrations, nil
}

func (m *Migrator) appliedVersions(ctx context.Context) (map[int64]bool, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT version FROM "+migrationTable)
	if err != nil {
		return nil, fmt.Errorf("read applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int64]bool)
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan migration version: %w", err)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate migration versions: %w", err)
	}
	return applied, nil
}

func (m *Migrator) applyUp(ctx context.Context, migration Migration) error {
	script, err := os.ReadFile(migration.UpPath)
	if err != nil {
		return fmt.Errorf("read up migration %d: %w", migration.Version, err)
	}
	if _, err := m.db.ExecContext(ctx, string(script)); err != nil {
		return fmt.Errorf("apply migration %d: %w", migration.Version, err)
	}
	if _, err := m.db.ExecContext(ctx,
		"INSERT INTO "+migrationTable+" (version, name) VALUES (?, ?)",
		migration.Version,
		migration.Name,
	); err != nil {
		return fmt.Errorf("record migration %d: %w", migration.Version, err)
	}
	return nil
}

func (m *Migrator) applyDown(ctx context.Context, migration Migration) error {
	script, err := os.ReadFile(migration.DownPath)
	if err != nil {
		return fmt.Errorf("read down migration %d: %w", migration.Version, err)
	}
	if _, err := m.db.ExecContext(ctx, string(script)); err != nil {
		return fmt.Errorf("rollback migration %d: %w", migration.Version, err)
	}
	if _, err := m.db.ExecContext(ctx,
		"DELETE FROM "+migrationTable+" WHERE version = ?",
		migration.Version,
	); err != nil {
		return fmt.Errorf("remove migration record %d: %w", migration.Version, err)
	}
	return nil
}

func (m *Migrator) withLock(ctx context.Context, operation func() error) error {
	var acquired sql.NullInt64
	if err := m.db.QueryRowContext(ctx, "SELECT GET_LOCK(?, 10)", lockName).Scan(&acquired); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	if !acquired.Valid || acquired.Int64 != 1 {
		return fmt.Errorf("migration lain sedang berjalan")
	}
	defer m.db.ExecContext(context.Background(), "SELECT RELEASE_LOCK(?)", lockName)

	return operation()
}

// DatabaseName returns the active MySQL database without exposing credentials.
func DatabaseName(ctx context.Context, db *sql.DB) (string, error) {
	var name sql.NullString
	if err := db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&name); err != nil {
		return "", fmt.Errorf("read active database: %w", err)
	}
	if !name.Valid || strings.TrimSpace(name.String) == "" {
		return "", fmt.Errorf("no active database selected")
	}
	return name.String, nil
}
