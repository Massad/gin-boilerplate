package models

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Massad/gin-boilerplate/db"
	"github.com/go-redis/redis/v7"
	jwt "github.com/golang-jwt/jwt/v4"
	uuid "github.com/google/uuid"
)

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// TokenDetails ...
type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUUID   string
	RefreshUUID  string
	AtExpires    int64
	RtExpires    int64
}

// AdminTokenDetails carries the single, long-lived superadmin access token.
type AdminTokenDetails struct {
	AccessToken string
	AccessUUID  string
	AtExpires   int64
}

// AccessDetails carries the tenant-scoped session identity extracted from a
// access token: an admin token is rejected here because it lacks the claims.
type AccessDetails struct {
	AccessUUID string
	UserID     uuid.UUID
	TenantID   uuid.UUID
	Role       string
}

// AdminAccessDetails carries the platform session identity of an admin token.
type AdminAccessDetails struct {
	AccessUUID string
	AdminID    uuid.UUID
}

// Token is the session token pair of a tenant session. RefreshToken is
// omitted when empty: the unified login reuses this model for its superadmin
// branch, which carries an access token only (no refresh half).
type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// AdminToken is the token payload of a superadmin session (no refresh).
type AdminToken struct {
	AccessToken string `json:"access_token"`
}

// AuthModel ...
type AuthModel struct{}

// CreateToken issues an access/refresh pair whose claims carry the tenant and
// the role: the tenant is always resolved from the signed token, never from
// client-provided data.
func (m AuthModel) CreateToken(userID, tenantID uuid.UUID, role string) (*TokenDetails, error) {

	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
	td.AccessUUID = uuid.New().String()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUUID = uuid.New().String()

	var err error
	//Creating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUUID
	atClaims["user_id"] = userID.String()
	atClaims["tenant_id"] = tenantID.String()
	atClaims["role"] = role
	atClaims["exp"] = td.AtExpires

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}
	//Creating Refresh Token
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUUID
	rtClaims["user_id"] = userID.String()
	rtClaims["tenant_id"] = tenantID.String()
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return nil, err
	}
	return td, nil
}

// CreateAuth stores both session halves as "userID:tenantID:role" and indexes
// them under user_sessions:{userID} for global revocation.
func (m AuthModel) CreateAuth(userID, tenantID uuid.UUID, role string, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()
	value := fmt.Sprintf("%v:%v:%v", userID, tenantID, role)

	errAccess := db.GetRedis().Set(td.AccessUUID, value, at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}
	errRefresh := db.GetRedis().Set(td.RefreshUUID, value, rt.Sub(now)).Err()
	if errRefresh != nil {
		return errRefresh
	}
	return db.GetRedis().SAdd(fmt.Sprintf(userSessionIndex, userID), td.AccessUUID, td.RefreshUUID).Err()
}

// ExtractToken ...
func (m AuthModel) ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	//normally Authorization the_token_xxx
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

// VerifyToken ...
func (m AuthModel) VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := m.ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// ExtractTokenMetadata ...
func (m AuthModel) ExtractTokenMetadata(r *http.Request) (*AccessDetails, error) {
	token, err := m.VerifyToken(r)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	accessUUID, ok := claims["access_uuid"].(string)
	if !ok {
		return nil, errors.New("invalid access_uuid claim")
	}

	userID, err := uuidClaim(claims, "user_id")
	if err != nil {
		return nil, err
	}
	tenantID, err := uuidClaim(claims, "tenant_id")
	if err != nil {
		return nil, err
	}
	role, ok := claims["role"].(string)
	if !ok || role == "" {
		return nil, errors.New("invalid role claim")
	}

	return &AccessDetails{
		AccessUUID: accessUUID,
		UserID:     userID,
		TenantID:   tenantID,
		Role:       role,
	}, nil
}

// uuidClaim reads a UUID claim stored as its string representation.
func uuidClaim(claims jwt.MapClaims, name string) (uuid.UUID, error) {
	value, ok := claims[name].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid %s claim", name)
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid %s claim", name)
	}
	return parsed, nil
}

// FetchAuth verifies that the session exists in Redis and belongs to the
// claimed user, before handing the tenant identity back to the middleware.
func (m AuthModel) FetchAuth(authD *AccessDetails) (uuid.UUID, error) {
	stored, err := db.GetRedis().Get(authD.AccessUUID).Result()
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
	if userID != authD.UserID {
		return uuid.Nil, errors.New("session owner mismatch")
	}
	return userID, nil
}

// DeleteAuth consumes one session key and removes it from the user's session
// index.
func (m AuthModel) DeleteAuth(givenUUID string, userID uuid.UUID) (int64, error) {
	if userID != uuid.Nil {
		if err := db.GetRedis().SRem(fmt.Sprintf(userSessionIndex, userID), givenUUID).Err(); err != nil {
			return 0, err
		}
	}
	deleted, err := db.GetRedis().Del(givenUUID).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

// RevokeUserSessions deletes every active session of a user (role change,
// administrative logout).
func (m AuthModel) RevokeUserSessions(userID uuid.UUID) error {
	index := fmt.Sprintf(userSessionIndex, userID)
	sessions, err := db.GetRedis().SMembers(index).Result()
	if err != nil {
		return err
	}
	if len(sessions) > 0 {
		if err := db.GetRedis().Del(sessions...).Err(); err != nil {
			return err
		}
	}
	return db.GetRedis().Del(index).Err()
}

// userSessionIndex is the Redis set holding every active UUID of a user.
const userSessionIndex = "user_sessions:%v"

var consumeRefreshAuth = redis.NewScript(`
local stored = redis.call("GET", KEYS[1])
if stored and string.match(stored, "^([^:]+)") == ARGV[1] then
    return redis.call("DEL", KEYS[1])
end
return 0
`)

// ConsumeRefreshAuth atomically verifies the stored owner and consumes a refresh
// token once. Only the leading userID segment is compared, so the stored value
// format ("userID:tenantID:role") stays robust across role changes between a
// login and a refresh. A mismatched owner must not invalidate someone else's
// session.
func (m AuthModel) ConsumeRefreshAuth(refreshUUID string, userID uuid.UUID) (int64, error) {
	return consumeRefreshAuth.Run(db.GetRedis(), []string{refreshUUID}, userID.String()).Int64()
}

// CreateAdminToken issues the single superadmin access token: signed with the
// dedicated ADMIN_ACCESS_SECRET, ~8h TTL, no refresh.
func (m AuthModel) CreateAdminToken(adminID uuid.UUID) (*AdminTokenDetails, error) {
	td := &AdminTokenDetails{}
	td.AtExpires = time.Now().Add(time.Hour * 8).Unix()
	td.AccessUUID = uuid.New().String()

	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["access_uuid"] = td.AccessUUID
	claims["admin_id"] = adminID.String()
	claims["role"] = "superadmin"
	claims["exp"] = td.AtExpires

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	var err error
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ADMIN_ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}
	return td, nil
}

// CreateAdminAuth stores the admin session under the admin: Redis prefix.
func (m AuthModel) CreateAdminAuth(adminID uuid.UUID, td *AdminTokenDetails) error {
	expires := time.Unix(td.AtExpires, 0)
	return db.GetRedis().Set(adminSessionKey(td.AccessUUID), adminID.String(), expires.Sub(time.Now())).Err()
}

// ExtractAdminTokenMetadata validates an admin token against the ADMIN_ACCESS
// SECRET: a tenant token carries a different signature and no admin_id claim.
func (m AuthModel) ExtractAdminTokenMetadata(r *http.Request) (*AdminAccessDetails, error) {
	tokenString := m.ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ADMIN_ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	accessUUID, ok := claims["access_uuid"].(string)
	if !ok {
		return nil, errors.New("invalid access_uuid claim")
	}
	adminID, err := uuidClaim(claims, "admin_id")
	if err != nil {
		return nil, err
	}
	return &AdminAccessDetails{AccessUUID: accessUUID, AdminID: adminID}, nil
}

// FetchAdmin verifies the admin session exists in Redis.
func (m AuthModel) FetchAdmin(authD *AdminAccessDetails) (uuid.UUID, error) {
	stored, err := db.GetRedis().Get(adminSessionKey(authD.AccessUUID)).Result()
	if err != nil {
		return uuid.Nil, err
	}
	adminID, err := uuid.Parse(stored)
	if err != nil {
		return uuid.Nil, errors.New("invalid admin session")
	}
	if adminID != authD.AdminID {
		return uuid.Nil, errors.New("admin session owner mismatch")
	}
	return adminID, nil
}

// DeleteAdminAuth consumes an admin session: immediate logout, no refresh.
func (m AuthModel) DeleteAdminAuth(givenUUID string) (int64, error) {
	return db.GetRedis().Del(adminSessionKey(givenUUID)).Result()
}

// adminSessionKey prefixes admin session keys so they cannot collide with
// tenant session keys.
func adminSessionKey(uuid string) string {
	return "admin:" + uuid
}
