package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateSigningSecrets(t *testing.T) {
	access := strings.Repeat("a1", 32)
	refresh := strings.Repeat("b2", 32)
	admin := strings.Repeat("c3", 32)
	cases := []struct {
		name, access, refresh, admin, want string
	}{
		{"independent keys", access, refresh, admin, ""},
		{"missing access key", "", refresh, admin, "ACCESS_SECRET"},
		{"missing refresh key", access, "", admin, "REFRESH_SECRET"},
		{"missing admin key", access, refresh, "", "ADMIN_ACCESS_SECRET"},
		{"whitespace access key", strings.Repeat(" ", 64), refresh, admin, "ACCESS_SECRET"},
		{"whitespace refresh key", access, strings.Repeat("\t", 64), admin, "REFRESH_SECRET"},
		{"whitespace admin key", access, refresh, strings.Repeat("\n", 64), "ADMIN_ACCESS_SECRET"},
		{"short access key", strings.Repeat("a", 31), refresh, admin, "ACCESS_SECRET"},
		{"short refresh key", access, strings.Repeat("b", 31), admin, "REFRESH_SECRET"},
		{"short admin key", access, refresh, strings.Repeat("c", 31), "ADMIN_ACCESS_SECRET"},
		{"old sample access key", "ashasdjhjhjadhasdaa123", refresh, admin, "ACCESS_SECRET"},
		{"old sample refresh key", access, "hjsajdhkjhf41jhagggdga", admin, "REFRESH_SECRET"},
		{"old sample admin key", access, refresh, "admin-sentence-too-short", "ADMIN_ACCESS_SECRET"},
		{"access and refresh equal", access, access, admin, "must be different"},
		{"access and admin equal", access, refresh, access, "must be different"},
		{"refresh and admin equal", access, refresh, refresh, "must be different"},
		{"minimum length", strings.Repeat("a", 32), strings.Repeat("b", 32), strings.Repeat("c", 32), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]string{"ACCESS_SECRET": tc.access, "REFRESH_SECRET": tc.refresh, "ADMIN_ACCESS_SECRET": tc.admin}
			err := ValidateSigningSecrets(func(name string) string { return values[name] })
			if tc.want == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.want)
			for name, key := range map[string]string{"ACCESS_SECRET": tc.access, "REFRESH_SECRET": tc.refresh, "ADMIN_ACCESS_SECRET": tc.admin} {
				if key != "" {
					require.NotContains(t, err.Error(), key, "%s key must not be revealed", name)
				}
			}
		})
	}
}
