package config

import (
	"fmt"
	"strings"
)

// ValidateSigningSecrets checks the HMAC keys before the application connects to
// its databases or serves requests. Generate each key independently with a
// cryptographically secure random generator; length alone does not prove entropy.
func ValidateSigningSecrets(getenv func(string) string) error {
	for _, name := range []string{"ACCESS_SECRET", "REFRESH_SECRET"} {
		value := getenv(name)
		if strings.TrimSpace(value) == "" || len(value) < 32 {
			return fmt.Errorf("error: %s must contain at least 32 bytes; generate an independent random key with openssl rand -hex 32", name)
		}
	}
	if getenv("ACCESS_SECRET") == getenv("REFRESH_SECRET") {
		return fmt.Errorf("error: ACCESS_SECRET and REFRESH_SECRET must be different")
	}
	return nil
}
