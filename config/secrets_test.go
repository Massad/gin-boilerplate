package config

import (
	"strings"
	"testing"
)

func TestValidateSigningSecrets(t *testing.T) {
	access := strings.Repeat("a1", 32)
	refresh := strings.Repeat("b2", 32)
	cases := []struct {
		name, access, refresh, want string
	}{
		{"independent keys without unused Redis secret", access, refresh, ""},
		{"missing access key", "", refresh, "ACCESS_SECRET"},
		{"missing refresh key", access, "", "REFRESH_SECRET"},
		{"whitespace access key", strings.Repeat(" ", 64), refresh, "ACCESS_SECRET"},
		{"whitespace refresh key", access, strings.Repeat("\t", 64), "REFRESH_SECRET"},
		{"short access key", strings.Repeat("a", 31), refresh, "ACCESS_SECRET"},
		{"short refresh key", access, strings.Repeat("b", 31), "REFRESH_SECRET"},
		{"old sample access key", "ashasdjhjhjadhasdaa123", refresh, "ACCESS_SECRET"},
		{"old sample refresh key", access, "hjsajdhkjhf41jhagggdga", "REFRESH_SECRET"},
		{"same keys", access, access, "must be different"},
		{"minimum length", strings.Repeat("a", 32), strings.Repeat("b", 32), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]string{"ACCESS_SECRET": tc.access, "REFRESH_SECRET": tc.refresh}
			err := ValidateSigningSecrets(func(name string) string { return values[name] })
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %v", tc.want, err)
			}
			if tc.access != "" && strings.Contains(err.Error(), tc.access) {
				t.Fatal("error reveals access key")
			}
			if tc.refresh != "" && strings.Contains(err.Error(), tc.refresh) {
				t.Fatal("error reveals refresh key")
			}
		})
	}
}
