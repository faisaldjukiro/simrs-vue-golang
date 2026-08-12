// Package seeder applies repeatable SQL seed files to the application DB.
package seeder

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const lockName = "simrs_backend_application_seeders"

type Seeder struct {
	db        *sql.DB
	directory string
}

func New(db *sql.DB, directory string) *Seeder {
	return &Seeder{db: db, directory: directory}
}

func (s *Seeder) Run(ctx context.Context) error {
	return s.withLock(ctx, func() error {
		files, err := s.discover()
		if err != nil {
			return err
		}
		if len(files) == 0 {
			fmt.Println("no seed files found")
			return nil
		}

		for _, path := range files {
			script, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read seed file %s: %w", filepath.Base(path), err)
			}
			if _, err := s.db.ExecContext(ctx, string(script)); err != nil {
				return fmt.Errorf("apply seed file %s: %w", filepath.Base(path), err)
			}
			fmt.Printf("seeded %s\n", filepath.Base(path))
		}
		return nil
	})
}

func (s *Seeder) discover() ([]string, error) {
	entries, err := os.ReadDir(s.directory)
	if err != nil {
		return nil, fmt.Errorf("read seeders directory: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".sql") {
			continue
		}
		files = append(files, filepath.Join(s.directory, entry.Name()))
	}
	sort.Strings(files)
	return files, nil
}

func (s *Seeder) withLock(ctx context.Context, operation func() error) error {
	var acquired sql.NullInt64
	if err := s.db.QueryRowContext(ctx, "SELECT GET_LOCK(?, 10)", lockName).Scan(&acquired); err != nil {
		return fmt.Errorf("acquire seeder lock: %w", err)
	}
	if !acquired.Valid || acquired.Int64 != 1 {
		return fmt.Errorf("seeder lain sedang berjalan")
	}
	defer s.db.ExecContext(context.Background(), "SELECT RELEASE_LOCK(?)", lockName)

	return operation()
}
