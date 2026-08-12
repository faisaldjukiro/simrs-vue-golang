// Command seed populates initial data in the application database.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"simrs-backend/internal/config"
	"simrs-backend/internal/platform/database"
	"simrs-backend/internal/platform/database/migrator"
	"simrs-backend/internal/platform/database/seeder"
)

func main() {
	if len(os.Args) != 1 {
		fail("usage: go run ./cmd/seed")
	}

	if err := config.LoadEnvFile(".env"); err != nil {
		fail(err.Error())
	}
	databaseConfig, err := config.ApplicationDatabase()
	if err != nil {
		fail(err.Error())
	}
	if err := config.ValidateMigrationTarget(databaseConfig); err != nil {
		fail(err.Error())
	}

	ctx := context.Background()
	db, err := database.OpenMySQL(ctx, databaseConfig)
	if err != nil {
		fail(err.Error())
	}
	defer db.Close()

	activeDatabase, err := migrator.DatabaseName(ctx, db)
	if err != nil {
		fail(err.Error())
	}
	if !strings.EqualFold(activeDatabase, databaseConfig.Name) {
		fail(fmt.Sprintf("seeder ditolak: database aktif %q tidak sesuai konfigurasi %q", activeDatabase, databaseConfig.Name))
	}

	fmt.Printf("target database: %s\n", activeDatabase)
	if err := seeder.New(db, "migrations/seeders").Run(ctx); err != nil {
		fail(err.Error())
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, "error:", message)
	os.Exit(1)
}
