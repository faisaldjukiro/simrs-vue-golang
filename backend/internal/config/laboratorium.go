package config

import (
	"os"
	"strings"
)

// Samakan dengan AKTIFKANBILLINGPARSIAL Khanza; jangan membuka billing secara default.
func LaboratoriumBillingParsial() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("LAB_AKTIFKAN_BILLING_PARSIAL")), "yes")
}
