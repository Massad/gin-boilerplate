package config

import (
	"fmt"
	"strings"
)

// signingSecrets are the HMAC keys validated before the application connects to
// its databases or serves requests: the tenant access/refresh pair and the
// platform superadmin access key.
var signingSecrets = []string{"ACCESS_SECRET", "REFRESH_SECRET", "ADMIN_ACCESS_SECRET"}

// ValidateSigningSecrets checks the HMAC keys before the application connects to
// its databases or serves requests. Generate each key independently with a
// cryptographically secure random generator; length alone does not prove entropy.
func ValidateSigningSecrets(getenv func(string) string) error {
	for _, name := range signingSecrets {
		value := getenv(name)
		if strings.TrimSpace(value) == "" || len(value) < 32 {
			return fmt.Errorf("error: %s must contain at least 32 bytes; generate an independent random key with openssl rand -hex 32", name)
		}
	}
	for _, pair := range [][2]string{
		{"ACCESS_SECRET", "REFRESH_SECRET"},
		{"ACCESS_SECRET", "ADMIN_ACCESS_SECRET"},
		{"REFRESH_SECRET", "ADMIN_ACCESS_SECRET"},
	} {
		if getenv(pair[0]) == getenv(pair[1]) {
			return fmt.Errorf("error: %s and %s must be different", pair[0], pair[1])
		}
	}
	return nil
}
