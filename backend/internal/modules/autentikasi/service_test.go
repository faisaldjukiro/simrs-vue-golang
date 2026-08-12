package autentikasi

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestSeededAdminPasswordIsCompatible(t *testing.T) {
	const hash = "$2y$12$/CBvNutc7PARdjAeInE5NeQ6DCUzxkiSQrfMGyY25LV5EOWFjfysG"
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("password")); err != nil {
		t.Fatalf("seeded admin password is not compatible with Go bcrypt: %v", err)
	}
}

func TestHashTokenDoesNotStoreRawToken(t *testing.T) {
	const token = "plain-access-token"
	hashed := hashToken(token)
	if hashed == token {
		t.Fatal("token hash must differ from raw token")
	}
	if len(hashed) != 64 {
		t.Fatalf("token hash length = %d, want 64", len(hashed))
	}
}
