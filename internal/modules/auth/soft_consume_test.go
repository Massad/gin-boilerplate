//go:build all

package auth_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Massad/gin-boilerplate/internal/platform/redisdb"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// The dedicated soft-consume suite (design decision 7): it exercises the Lua
// semantics of the refresh record against the disposable Redis index pinned
// by the package TestMain. The record states are the ones the repository
// encodes: active "userID:tenantID:role" (format unchanged) and consumed
// "consumed:<state>:<value>".

// recordValue seeds one refresh record for the user in the given state and
// registers its cleanup.
func recordValue(userID, tenantID uuid.UUID) string {
	return fmt.Sprintf("%v:%v:%v", userID, tenantID, "owner")
}

func seedRecord(t *testing.T, value string, ttl time.Duration) string {
	t.Helper()
	id := uuid.NewString()
	require.NoError(t, redisdb.GetRedis().Set(id, value, ttl).Err())
	t.Cleanup(func() { redisdb.GetRedis().Del(id) })
	return id
}

func storedValue(t *testing.T, id string) string {
	t.Helper()
	stored, err := redisdb.GetRedis().Get(id).Result()
	require.NoError(t, err, "the record must not have been deleted")
	return stored
}

// TestConsumeRefreshMarksDelivered: the first consumption claims the flow
// (code 1), encodes the delivered state into the value (never deleting the
// record) and bounds the TTL by the grace.
func TestConsumeRefreshMarksDelivered(t *testing.T) {
	svc := middlewareTestSetup(t)
	userID := uuid.New()
	tenantID := uuid.New()
	value := recordValue(userID, tenantID)
	id := seedRecord(t, value, 7*24*time.Hour)

	code, err := svc.ConsumeRefreshAuth(id, userID)
	require.NoError(t, err)
	require.Equal(t, int64(1), code, "the first consumption must be claimed")
	require.Equal(t, "consumed:1:"+value, storedValue(t, id), "the delivered state must be encoded in the value")

	ttl, err := redisdb.GetRedis().PTTL(id).Result()
	require.NoError(t, err)
	require.Greater(t, ttl.Milliseconds(), int64(0), "the consumed record keeps a positive TTL")
	require.LessOrEqual(t, ttl.Milliseconds(), int64(300000), "the TTL never exceeds the 5 minutes grace")

	// A replay of a delivered record is refused: same shape as the invalid
	// path (code 0), record untouched.
	code, err = svc.ConsumeRefreshAuth(id, userID)
	require.NoError(t, err)
	require.Equal(t, int64(0), code, "a delivered record must be refused")
	require.Equal(t, "consumed:1:"+value, storedValue(t, id))
}

// TestConsumeRefreshGraceNeverOutlivesToken: min(grace, remaining TTL) — a
// nearly expired token is consumed within its own remaining lifetime, never
// revived past it.
func TestConsumeRefreshGraceNeverOutlivesToken(t *testing.T) {
	svc := middlewareTestSetup(t)
	userID := uuid.New()
	tenantID := uuid.New()
	value := recordValue(userID, tenantID)
	id := seedRecord(t, value, time.Second)

	code, err := svc.ConsumeRefreshAuth(id, userID)
	require.NoError(t, err)
	require.Equal(t, int64(1), code)

	ttl, err := redisdb.GetRedis().PTTL(id).Result()
	require.NoError(t, err)
	require.GreaterOrEqual(t, ttl.Milliseconds(), int64(900), "the remaining TTL is preserved")
	require.LessOrEqual(t, ttl.Milliseconds(), int64(1000), "the remaining TTL is never extended to the grace")
}

// TestReleaseRefreshOnlyDowngradesDelivered: only a claimed consumption can
// release its record; the release enables the recovery of the replayed token.
func TestReleaseRefreshOnlyDowngradesDelivered(t *testing.T) {
	svc := middlewareTestSetup(t)
	userID := uuid.New()
	tenantID := uuid.New()
	value := recordValue(userID, tenantID)
	id := seedRecord(t, value, 7*24*time.Hour)

	// An active record is never released.
	code, err := svc.ReleaseRefresh(id, userID)
	require.NoError(t, err)
	require.Equal(t, int64(0), code)
	require.Equal(t, value, storedValue(t, id), "an active record is untouched")

	code, err = svc.ConsumeRefreshAuth(id, userID)
	require.NoError(t, err)
	require.Equal(t, int64(1), code)

	code, err = svc.ReleaseRefresh(id, userID)
	require.NoError(t, err)
	require.Equal(t, int64(1), code, "the claimed consumption may release")
	require.Equal(t, "consumed:0:"+value, storedValue(t, id))

	code, err = svc.ReleaseRefresh(id, userID)
	require.NoError(t, err)
	require.Equal(t, int64(0), code, "a not-delivered record is not released twice")
}

// TestReplayDuringGraceRecoversOnce: the replay of a released record claims
// the single recovery; every later replay or release is refused.
func TestReplayDuringGraceRecoversOnce(t *testing.T) {
	svc := middlewareTestSetup(t)
	userID := uuid.New()
	tenantID := uuid.New()
	value := recordValue(userID, tenantID)
	id := seedRecord(t, value, 7*24*time.Hour)

	code, err := svc.ConsumeRefreshAuth(id, userID)
	require.NoError(t, err)
	require.Equal(t, int64(1), code)
	code, err = svc.ReleaseRefresh(id, userID)
	require.NoError(t, err)
	require.Equal(t, int64(1), code)

	code, err = svc.ConsumeRefreshAuth(id, userID)
	require.NoError(t, err)
	require.Equal(t, int64(2), code, "the replay during the grace claims the recovery")
	require.Equal(t, "consumed:recovered:"+value, storedValue(t, id))

	ttl, err := redisdb.GetRedis().PTTL(id).Result()
	require.NoError(t, err)
	require.Greater(t, ttl.Milliseconds(), int64(0))
	require.LessOrEqual(t, ttl.Milliseconds(), int64(300000))

	code, err = svc.ConsumeRefreshAuth(id, userID)
	require.NoError(t, err)
	require.Equal(t, int64(0), code, "the recovery is unique")
	require.Equal(t, "consumed:recovered:"+value, storedValue(t, id))

	code, err = svc.ReleaseRefresh(id, userID)
	require.NoError(t, err)
	require.Equal(t, int64(0), code)
}

// TestConsumeRefreshOwnerMismatchLeavesRecord: a mismatched owner never
// consumes, never releases and never claims someone else's recovery.
func TestConsumeRefreshOwnerMismatchLeavesRecord(t *testing.T) {
	svc := middlewareTestSetup(t)
	userID := uuid.New()
	tenantID := uuid.New()
	stranger := uuid.New()
	value := recordValue(userID, tenantID)

	id := seedRecord(t, value, 7*24*time.Hour)
	code, err := svc.ConsumeRefreshAuth(id, stranger)
	require.NoError(t, err)
	require.Equal(t, int64(0), code, "a foreign user never consumes the record")
	require.Equal(t, value, storedValue(t, id), "the mismatched consumption leaves the record untouched")
	code, err = svc.ReleaseRefresh(id, stranger)
	require.NoError(t, err)
	require.Equal(t, int64(0), code)
	code, err = svc.ConsumeRefreshAuth(id, userID)
	require.NoError(t, err)
	require.Equal(t, int64(1), code, "the owner still consumes its record")

	code, err = svc.ReleaseRefresh(id, userID)
	require.NoError(t, err)
	require.Equal(t, int64(1), code)
	code, err = svc.ConsumeRefreshAuth(id, stranger)
	require.NoError(t, err)
	require.Equal(t, int64(0), code, "a foreign user never claims the recovery of a released record")
	require.Equal(t, "consumed:0:"+value, storedValue(t, id))
	code, err = svc.ConsumeRefreshAuth(id, userID)
	require.NoError(t, err)
	require.Equal(t, int64(2), code, "the owner still recovers its record")
}

// TestConsumeRefreshFailsClosedOnUnknownState: any value outside the known
// states is refused (fail-closed) and left untouched.
func TestConsumeRefreshFailsClosedOnUnknownState(t *testing.T) {
	svc := middlewareTestSetup(t)
	userID := uuid.New()
	tenantID := uuid.New()

	unknownFlag := seedRecord(t, "consumed:9:"+recordValue(userID, tenantID), 7*24*time.Hour)
	code, err := svc.ConsumeRefreshAuth(unknownFlag, userID)
	require.NoError(t, err)
	require.Equal(t, int64(0), code, "an unknown consumed state fails closed")
	require.Equal(t, "consumed:9:"+recordValue(userID, tenantID), storedValue(t, unknownFlag))

	malformed := seedRecord(t, "consumed:", 7*24*time.Hour)
	code, err = svc.ConsumeRefreshAuth(malformed, userID)
	require.NoError(t, err)
	require.Equal(t, int64(0), code, "a consumed state without owner fails closed")
	require.Equal(t, "consumed:", storedValue(t, malformed))

	garbage := seedRecord(t, "garbage-without-colon-segments", 7*24*time.Hour)
	code, err = svc.ConsumeRefreshAuth(garbage, userID)
	require.NoError(t, err)
	require.Equal(t, int64(0), code, "an unparsable value fails closed")
	require.Equal(t, "garbage-without-colon-segments", storedValue(t, garbage))
}

// TestSweepConsumedPurgesRelics: the sweep counts the consumed records,
// purges the ones whose TTL anomaly makes them unevictable, and never touches
// an active record nor a consumed record whose grace is still running.
func TestSweepConsumedPurgesRelics(t *testing.T) {
	svc := middlewareTestSetup(t)
	userID := uuid.New()
	tenantID := uuid.New()
	value := recordValue(userID, tenantID)

	active := seedRecord(t, value, 7*24*time.Hour)
	inGrace := seedRecord(t, "consumed:1:"+value, 4*time.Minute)
	relicNoTTL := seedRecord(t, "consumed:1:"+value, 0)
	relicLongTTL := seedRecord(t, "consumed:0:"+value, time.Hour)
	index := fmt.Sprintf("user_sessions:%v", userID)
	require.NoError(t, redisdb.GetRedis().SAdd(
		index, active, inGrace, relicNoTTL, relicLongTTL,
	).Err())
	t.Cleanup(func() { redisdb.GetRedis().Del(index) })

	consumed, purged, err := svc.SweepConsumed()
	require.NoError(t, err)
	require.GreaterOrEqual(t, consumed, int64(3), "the consumed records must be counted")
	require.GreaterOrEqual(t, purged, int64(2), "the TTL relics must be purged")

	remains, err := redisdb.GetRedis().Exists(active).Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), remains, "the sweep never touches an active record")
	remains, err = redisdb.GetRedis().Exists(inGrace).Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), remains, "the sweep never touches a consumed record whose grace runs")
	remains, err = redisdb.GetRedis().Exists(relicNoTTL).Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), remains, "the sweep purges the relic without TTL")
	remains, err = redisdb.GetRedis().Exists(relicLongTTL).Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), remains, "the sweep purges the relic with a TTL above the grace")
}

// TestSweepTickerLifecycle: the sweep ticker runs in-process and stops
// cleanly (start/stop tied to the service, testable).
func TestSweepTickerLifecycle(t *testing.T) {
	svc := middlewareTestSetup(t)
	userID := uuid.New()
	tenantID := uuid.New()
	relic := seedRecord(t, "consumed:1:"+recordValue(userID, tenantID), 0)
	index := fmt.Sprintf("user_sessions:%v", userID)
	require.NoError(t, redisdb.GetRedis().SAdd(index, relic).Err())
	t.Cleanup(func() { redisdb.GetRedis().Del(index) })

	stop := svc.StartSweep(20 * time.Millisecond)
	t.Cleanup(stop)

	require.Eventually(t, func() bool {
		remains, err := redisdb.GetRedis().Exists(relic).Result()
		require.NoError(t, err)
		return remains == 0
	}, 2*time.Second, 20*time.Millisecond, "the sweep ticker must purge the relic")

	stop()
	stop()
}
