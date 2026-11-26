package redisdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRedisDBIndex proves the shared resolution of the REDIS_DB index: an
// absent or empty value falls back to DefaultRedisDB, integer forms within
// the default Redis range are accepted, anything else is rejected.
func TestRedisDBIndex(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		want    int
		wantErr bool
	}{
		{name: "absent value means dev", value: "", want: 1},
		{name: "blank value means dev", value: "   ", want: 1},
		{name: "tests index", value: "2", want: 2},
		{name: "tests index trimmed", value: " 2 ", want: 2},
		{name: "zero is a valid index", value: "0", want: 0},
		{name: "last valid index", value: "15", want: 15},
		{name: "plus sign is accepted", value: "+2", want: 2},
		{name: "leading zero is accepted", value: "02", want: 2},
		{name: "negative zero resolves to zero", value: "-0", want: 0},
		{name: "negative index is rejected", value: "-1", wantErr: true},
		{name: "index over the range is rejected", value: "16", wantErr: true},
		{name: "non-numeric value is rejected", value: "abc", wantErr: true},
		{name: "decimal value is rejected", value: "1.5", wantErr: true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			getenv := func(name string) string {
				if name == "REDIS_DB" {
					return testCase.value
				}
				return ""
			}
			index, err := RedisDBIndex(getenv)
			if testCase.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, testCase.want, index)
		})
	}
}

// TestInitRedisRejectsOutOfRange proves the explicit failure path without any
// Redis service: an out-of-range REDIS_DB stops the caller instead of
// silently drifting to the dev index.
func TestInitRedisRejectsOutOfRange(t *testing.T) {
	t.Setenv("REDIS_DB", "16")
	require.Error(t, InitRedis(), "an out-of-range REDIS_DB must be rejected explicitly")
}

// TestInitRedisPutsTheParsedIndexOnTheClient proves the nominal path without
// any service: NewClient does not connect at creation, so the parsed index
// can be observed on the client options.
func TestInitRedisPutsTheParsedIndexOnTheClient(t *testing.T) {
	t.Setenv("REDIS_DB", "3")
	require.NoError(t, InitRedis())
	require.Equal(t, 3, GetRedis().Options().DB, "the parsed index must reach the client")
}
