// Command migrate manages the application database schema.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"simrs-backend/internal/config"
	"simrs-backend/internal/platform/database"
	"simrs-backend/internal/platform/database/migrator"
)

func main() {
	if len(os.Args) != 2 {
		fail("usage: go run ./cmd/migrate [up|status|down|fresh]")
	}

	command := os.Args[1]
	if command != "up" && command != "status" && command != "down" && command != "fresh" {
		fail("command tidak dikenal; gunakan up, status, down, atau fresh")
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
	if activeDatabase != databaseConfig.Name {
		fail(fmt.Sprintf("migration ditolak: database aktif %q tidak sesuai konfigurasi %q", activeDatabase, databaseConfig.Name))
	}

	fmt.Printf("target database: %s\n", activeDatabase)
	runner := migrator.New(db, "migrations")
	if command == "fresh" {
		confirmFresh(activeDatabase)
	}

	switch command {
	case "up":
		err = runner.Up(ctx)
	case "status":
		err = runner.Status(ctx)
	case "down":
		err = runner.Down(ctx)
	case "fresh":
		err = runner.Fresh(ctx)
	}
	if err != nil {
		fail(err.Error())
	}
}

func confirmFresh(databaseName string) {
	fmt.Printf("PERINGATAN: seluruh tabel aplikasi pada %q akan dibuat ulang.\n", databaseName)
	fmt.Printf("Ketik nama database %q untuk melanjutkan: ", databaseName)

	answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fail("gagal membaca konfirmasi fresh")
	}
	if strings.TrimSpace(answer) != databaseName {
		fail("fresh dibatalkan: nama database tidak cocok")
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, "error:", message)
	os.Exit(1)
}
