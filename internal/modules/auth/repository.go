package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	uuid "github.com/google/uuid"
)

// refreshGrace is the fixed retention window of a consumed refresh record
// (design decision 7): the consumption applies min(grace, remaining TTL), so
// the grace never outlives the token it protects. TTL eviction stays primary.
const refreshGrace = 5 * time.Minute

// userSessionIndex is the Redis set holding every active UUID of a user.
const userSessionIndex = "user_sessions:%v"

// consumeRefreshScript softly consumes the refresh record in one atomic
// operation (decision 7). KEYS[1] is the record UUID, ARGV[1] the claimed
// owner, ARGV[2] the grace window in milliseconds. States: active
// "userID:tenantID:role" (format unchanged), consumed
// "consumed:<state>:<value>" with state 1 (delivered), 0 (not delivered) and
// recovered (single re-emission). The owner segment is compared before any
// state transition; a mismatched owner never invalidates someone else's
// session. Returns 1 when the first consumption is claimed (delivered marked,
// TTL reduced to min(grace, remaining)), 2 when the recovery of a
// not-delivered record is claimed, and 0 when the record is refused:
// missing, mismatched owner, already delivered or recovered, unknown state
// (fail-closed).
var consumeRefreshScript = redis.NewScript(`
local stored = redis.call("GET", KEYS[1])
if not stored then
    return 0
end
local consumed_state = nil
if string.sub(stored, 1, 9) == "consumed:" then
    consumed_state = string.match(stored, "^consumed:([^:]+):")
    if consumed_state ~= "0" and consumed_state ~= "1" and consumed_state ~= "recovered" then
        return 0
    end
end
local owner = ""
if consumed_state == nil then
    owner = string.match(stored, "^([^:]+)")
else
    owner = string.match(stored, "^consumed:[^:]+:([^:]+)")
end
if not owner or owner ~= ARGV[1] then
    return 0
end
if consumed_state == "1" or consumed_state == "recovered" then
    return 0
end
local remainder = stored
if consumed_state ~= nil then
    remainder = string.sub(stored, 11 + string.len(consumed_state))
end
local next_state = "1"
if consumed_state == "0" then
    next_state = "recovered"
end
local ttl = redis.call("PTTL", KEYS[1])
if ttl < 1 or ttl > tonumber(ARGV[2]) then
    ttl = tonumber(ARGV[2])
end
redis.call("SET", KEYS[1], "consumed:" .. next_state .. ":" .. remainder, "PX", ttl)
if next_state == "recovered" then
    return 2
end
return 1
`)

// releaseRefreshScript downgrades a claimed-but-not-delivered consumption:
// the consuming request failed before emitting the new pair, so the replayed
// token becomes recoverable during the remaining grace. Active, recovered and
// unknown records are never touched. Returns 1 on release, 0 otherwise.
var releaseRefreshScript = redis.NewScript(`
local stored = redis.call("GET", KEYS[1])
if not stored then
    return 0
end
if string.sub(stored, 1, 9) ~= "consumed:" then
    return 0
end
local consumed_state = string.match(stored, "^consumed:([^:]+):")
if consumed_state ~= "1" then
    return 0
end
local owner = string.match(stored, "^consumed:[^:]+:([^:]+)")
if not owner or owner ~= ARGV[1] then
    return 0
end
local remainder = string.sub(stored, 11 + string.len(consumed_state))
local ttl = redis.call("PTTL", KEYS[1])
if ttl < 1 or ttl > tonumber(ARGV[2]) then
    ttl = tonumber(ARGV[2])
end
redis.call("SET", KEYS[1], "consumed:0:" .. remainder, "PX", ttl)
return 1
`)

// sweepCheckScript classifies one candidate record for the sweep:
// 0 = not a consumed record (active records are never touched), 1 = consumed
// record whose grace is still running (kept), 2 = consumed relic whose TTL
// anomaly makes it unevictable (purged).
var sweepCheckScript = redis.NewScript(`
local stored = redis.call("GET", KEYS[1])
if not stored then
    return 0
end
if string.sub(stored, 1, 9) ~= "consumed:" then
    return 0
end
local ttl = redis.call("PTTL", KEYS[1])
if ttl >= 0 and ttl <= tonumber(ARGV[1]) then
    return 1
end
redis.call("DEL", KEYS[1])
return 2
`)

// repository implements SessionRepository on a Redis client: every key, index
// and value format of the session records stays in this file.
type repository struct {
	client *redis.Client
}

// NewRepository builds the session repository. The client is resolved once at
// the composition root (internal/app reads the platform singleton during the
// transitory window) and injected here; the module never imports
// internal/platform/redisdb.
func NewRepository(client *redis.Client) SessionRepository {
	return repository{client: client}
}

// CreateTenantSession stores the two session halves under their UUID keys
// with the "userID:tenantID:role" value and indexes them under
// user_sessions:{userID} for global revocation.
func (r repository) CreateTenantSession(userID, tenantID uuid.UUID, role string, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()
	value := fmt.Sprintf("%v:%v:%v", userID, tenantID, role)

	errAccess := r.client.Set(td.AccessUUID, value, at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}
	errRefresh := r.client.Set(td.RefreshUUID, value, rt.Sub(now)).Err()
	if errRefresh != nil {
		return errRefresh
	}
	return r.client.SAdd(fmt.Sprintf(userSessionIndex, userID), td.AccessUUID, td.RefreshUUID).Err()
}

// FetchTenantSession verifies that the record exists and belongs to the
// claimed user before returning the verified identity.
func (r repository) FetchTenantSession(accessUUID string, claimedUserID uuid.UUID) (uuid.UUID, error) {
	stored, err := r.client.Get(accessUUID).Result()
	if err != nil {
		return uuid.Nil, err
	}
	parts := strings.Split(stored, ":")
	if len(parts) < 3 {
		return uuid.Nil, errors.New("invalid session record")
	}
	userID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, errors.New("invalid session owner")
	}
	if userID != claimedUserID {
		return uuid.Nil, errors.New("session owner mismatch")
	}
	return userID, nil
}

// DeleteTenantSession consumes one session key and removes it from the user's
// session index.
func (r repository) DeleteTenantSession(givenUUID string, userID uuid.UUID) (int64, error) {
	if userID != uuid.Nil {
		if err := r.client.SRem(fmt.Sprintf(userSessionIndex, userID), givenUUID).Err(); err != nil {
			return 0, err
		}
	}
	deleted, err := r.client.Del(givenUUID).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

// RevokeUserSessions deletes every active session of a user (role change,
// administrative logout).
func (r repository) RevokeUserSessions(userID uuid.UUID) error {
	index := fmt.Sprintf(userSessionIndex, userID)
	sessions, err := r.client.SMembers(index).Result()
	if err != nil {
		return err
	}
	if len(sessions) > 0 {
		if err := r.client.Del(sessions...).Err(); err != nil {
			return err
		}
	}
	return r.client.Del(index).Err()
}

// ConsumeRefresh claims the soft consumption (code 1: first consumption, 2:
// recovery of a not-delivered record) or refuses it (code 0). The script
// computes min(grace, remaining TTL) so a nearly-expired token is never
// revived past its lifetime.
func (r repository) ConsumeRefresh(refreshUUID string, userID uuid.UUID) (int64, error) {
	return consumeRefreshScript.Run(r.client, []string{refreshUUID}, userID.String(), refreshGrace.Milliseconds()).Int64()
}

// ReleaseRefresh downgrades a claimed-but-not-delivered consumption (code 1)
// and leaves every other state untouched (code 0).
func (r repository) ReleaseRefresh(refreshUUID string, userID uuid.UUID) (int64, error) {
	return releaseRefreshScript.Run(r.client, []string{refreshUUID}, userID.String(), refreshGrace.Milliseconds()).Int64()
}

// CreateAdminSession stores the admin session under the admin: prefix.
func (r repository) CreateAdminSession(adminID uuid.UUID, td *AdminTokenDetails) error {
	expires := time.Unix(td.AtExpires, 0)
	return r.client.Set(adminSessionKey(td.AccessUUID), adminID.String(), expires.Sub(time.Now())).Err()
}

// FetchAdminSession verifies the admin session exists and belongs to the
// claimed admin.
func (r repository) FetchAdminSession(accessUUID string, claimedAdminID uuid.UUID) (uuid.UUID, error) {
	stored, err := r.client.Get(adminSessionKey(accessUUID)).Result()
	if err != nil {
		return uuid.Nil, err
	}
	adminID, err := uuid.Parse(stored)
	if err != nil {
		return uuid.Nil, errors.New("invalid admin session")
	}
	if adminID != claimedAdminID {
		return uuid.Nil, errors.New("admin session owner mismatch")
	}
	return adminID, nil
}

// DeleteAdminSession consumes an admin session: immediate logout, no refresh.
func (r repository) DeleteAdminSession(givenUUID string) (int64, error) {
	return r.client.Del(adminSessionKey(givenUUID)).Result()
}

// SweepConsumed enumerates the session indexes (the only records the module
// produces), counts the present consumed records and purges the relics whose
// TTL anomaly makes them unevictable. It never touches an active record nor a
// consumed record whose grace is still running.
func (r repository) SweepConsumed() (consumed, purged int64, err error) {
	var cursor uint64
	for {
		indexes, next, scanErr := r.client.Scan(cursor, "user_sessions:*", 100).Result()
		if scanErr != nil {
			return consumed, purged, scanErr
		}
		cursor = next
		for _, index := range indexes {
			members, membersErr := r.client.SMembers(index).Result()
			if membersErr != nil {
				return consumed, purged, membersErr
			}
			for _, member := range members {
				code, runErr := sweepCheckScript.Run(r.client, []string{member}, refreshGrace.Milliseconds()).Int64()
				if runErr != nil {
					return consumed, purged, runErr
				}
				switch code {
				case 1:
					consumed++
				case 2:
					consumed++
					purged++
				}
			}
		}
		if cursor == 0 {
			return consumed, purged, nil
		}
	}
}

// adminSessionKey prefixes admin session keys so they cannot collide with
// tenant session keys.
func adminSessionKey(uuid string) string {
	return "admin:" + uuid
}
