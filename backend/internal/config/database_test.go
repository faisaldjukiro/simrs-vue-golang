package config

import "testing"

func TestValidateMigrationTargetAllowsRootOnExplicitLocalTarget(t *testing.T) {
	t.Setenv("SIMRS_DB_DATABASE", "simrs")
	t.Setenv("MIGRATION_ALLOWED_HOST", "127.0.0.1")
	t.Setenv("MIGRATION_ALLOWED_DATABASE", "simrs-golang")

	err := ValidateMigrationTarget(Database{Host: "127.0.0.1", Name: "simrs-golang", Username: "root"})
	if err != nil {
		t.Fatalf("expected local root account to be allowed: %v", err)
	}
}

func TestValidateMigrationTargetRejectsSIMRSDatabase(t *testing.T) {
	t.Setenv("SIMRS_DB_DATABASE", "simrs")
	t.Setenv("MIGRATION_ALLOWED_HOST", "127.0.0.1")
	t.Setenv("MIGRATION_ALLOWED_DATABASE", "simrs")

	err := ValidateMigrationTarget(Database{Host: "127.0.0.1", Name: "simrs", Username: "root"})
	if err == nil {
		t.Fatal("expected SIMRS database to be rejected")
	}
}

func TestValidateMigrationTargetRejectsUnexpectedDatabase(t *testing.T) {
	t.Setenv("SIMRS_DB_DATABASE", "simrs")
	t.Setenv("MIGRATION_ALLOWED_HOST", "127.0.0.1")
	t.Setenv("MIGRATION_ALLOWED_DATABASE", "simrs-golang")

	err := ValidateMigrationTarget(Database{Host: "127.0.0.1", Name: "database_lain", Username: "root"})
	if err == nil {
		t.Fatal("expected unexpected database to be rejected")
	}
}

func TestValidateMigrationTargetRejectsUnexpectedHost(t *testing.T) {
	t.Setenv("SIMRS_DB_DATABASE", "simrs")
	t.Setenv("MIGRATION_ALLOWED_HOST", "127.0.0.1")
	t.Setenv("MIGRATION_ALLOWED_DATABASE", "simrs-golang")

	err := ValidateMigrationTarget(Database{Host: "192.168.20.1", Name: "simrs-golang", Username: "root"})
	if err == nil {
		t.Fatal("expected unexpected host to be rejected")
	}
}
