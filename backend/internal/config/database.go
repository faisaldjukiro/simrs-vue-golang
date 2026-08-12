package config

import (
	"fmt"
	"os"
	"strings"
)

type Database struct {
	Host     string
	Port     string
	Name     string
	Username string
	Password string
}

func ApplicationDatabase() (Database, error) {
	database := Database{
		Host:     valueOrDefault("DB_HOST", "127.0.0.1"),
		Port:     valueOrDefault("DB_PORT", "3306"),
		Name:     strings.TrimSpace(os.Getenv("DB_DATABASE")),
		Username: strings.TrimSpace(os.Getenv("DB_USERNAME")),
		Password: os.Getenv("DB_PASSWORD"),
	}

	if database.Name == "" {
		return Database{}, fmt.Errorf("DB_DATABASE wajib diisi")
	}
	if database.Username == "" {
		return Database{}, fmt.Errorf("DB_USERNAME wajib diisi")
	}
	return database, nil
}

func SIMRSDatabase() (Database, error) {
	database := Database{
		Host:     valueOrDefault("SIMRS_DB_HOST", "127.0.0.1"),
		Port:     valueOrDefault("SIMRS_DB_PORT", "3306"),
		Name:     strings.TrimSpace(os.Getenv("SIMRS_DB_DATABASE")),
		Username: strings.TrimSpace(os.Getenv("SIMRS_DB_USERNAME")),
		Password: os.Getenv("SIMRS_DB_PASSWORD"),
	}

	if database.Name == "" {
		return Database{}, fmt.Errorf("SIMRS_DB_DATABASE wajib diisi")
	}
	if database.Username == "" {
		return Database{}, fmt.Errorf("SIMRS_DB_USERNAME wajib diisi")
	}
	return database, nil
}

// ValidateMigrationTarget only allows the explicitly configured application
// host/database pair and always rejects the legacy SIMRS database.
func ValidateMigrationTarget(database Database) error {
	simrsDatabase := strings.TrimSpace(os.Getenv("SIMRS_DB_DATABASE"))
	if simrsDatabase != "" && strings.EqualFold(database.Name, simrsDatabase) {
		return fmt.Errorf("migration ditolak: target %q adalah database SIMRS", database.Name)
	}

	allowedHost := strings.TrimSpace(os.Getenv("MIGRATION_ALLOWED_HOST"))
	if allowedHost == "" {
		return fmt.Errorf("migration ditolak: MIGRATION_ALLOWED_HOST wajib diisi")
	}
	if !strings.EqualFold(database.Host, allowedHost) {
		return fmt.Errorf(
			"migration ditolak: DB_HOST %q tidak sama dengan MIGRATION_ALLOWED_HOST %q",
			database.Host,
			allowedHost,
		)
	}

	allowedDatabase := strings.TrimSpace(os.Getenv("MIGRATION_ALLOWED_DATABASE"))
	if allowedDatabase == "" {
		return fmt.Errorf("migration ditolak: MIGRATION_ALLOWED_DATABASE wajib diisi")
	}
	if !strings.EqualFold(database.Name, allowedDatabase) {
		return fmt.Errorf(
			"migration ditolak: DB_DATABASE %q tidak sama dengan MIGRATION_ALLOWED_DATABASE %q",
			database.Name,
			allowedDatabase,
		)
	}

	if strings.EqualFold(database.Username, "root") && !isLoopback(database.Host) {
		return fmt.Errorf("migration ditolak: root hanya diizinkan pada host lokal")
	}
	return nil
}

func isLoopback(host string) bool {
	switch strings.ToLower(strings.TrimSpace(host)) {
	case "127.0.0.1", "localhost", "::1":
		return true
	default:
		return false
	}
}

func valueOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
