// Package auth owns the session primitives: issuing and verifying the tenant
// and superadmin JWTs, the Redis session records with the soft-consume of the
// refresh token (design decision 7), the TokenAuth/AdminAuth/RequirePermission
// middlewares and the logout handlers. The context keys set by TokenAuth are
// "userID", "tenantID" and "role"; AdminAuth sets "adminID".
package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	uuid "github.com/google/uuid"
)

// SessionRepository is the port the service consumes: the Redis session
// records with the soft-consume mechanics of the refresh token. Formats,
// keys, index and scripts stay inside the repository implementation.
type SessionRepository interface {
	CreateTenantSession(userID, tenantID uuid.UUID, role string, td *TokenDetails) error
	FetchTenantSession(accessUUID string, claimedUserID uuid.UUID) (uuid.UUID, error)
	DeleteTenantSession(givenUUID string, userID uuid.UUID) (int64, error)
	RevokeUserSessions(userID uuid.UUID) error
	// ConsumeRefresh softly consumes the refresh record and returns the
	// flow signal: 1 = first consumption claimed, 2 = recovery of a
	// not-delivered record claimed, 0 = refused.
	ConsumeRefresh(refreshUUID string, userID uuid.UUID) (int64, error)
	// ReleaseRefresh downgrades a claimed-but-not-delivered consumption so
	// the replayed token becomes recoverable during the remaining grace.
	ReleaseRefresh(refreshUUID string, userID uuid.UUID) (int64, error)
	CreateAdminSession(adminID uuid.UUID, td *AdminTokenDetails) error
	FetchAdminSession(accessUUID string, adminID uuid.UUID) (uuid.UUID, error)
	DeleteAdminSession(givenUUID string) (int64, error)
	// SweepConsumed counts the present consumed records and purges the
	// relics the TTL failed to evict.
	SweepConsumed() (consumed, purged int64, err error)
}

// DefaultSweepInterval is the production cadence of the in-process sweep
// (decision 7: a ~5 minutes ticker; TTL eviction stays primary).
const DefaultSweepInterval = 5 * time.Minute

// Service implements the session primitives on top of its repository port.
// It is consumed by the module middlewares and handlers and, through them, by
// the root controllers during the transitory migration window.
type Service struct {
	secrets  Secrets
	sessions SessionRepository
}

// NewService builds the service on the injected signing secrets and session
// repository.
func NewService(secrets Secrets, sessions SessionRepository) *Service {
	return &Service{secrets: secrets, sessions: sessions}
}

// CreateToken issues an access/refresh pair whose claims carry the tenant and
// the role: the tenant is always resolved from the signed token, never from
// client-provided data.
func (s *Service) CreateToken(userID, tenantID uuid.UUID, role string) (*TokenDetails, error) {

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
	td.AccessToken, err = at.SignedString([]byte(s.secrets.Access))
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
	td.RefreshToken, err = rt.SignedString([]byte(s.secrets.Refresh))
	if err != nil {
		return nil, err
	}
	return td, nil
}

// CreateAuth stores both session halves as "userID:tenantID:role" and indexes
// them under user_sessions:{userID} for global revocation.
func (s *Service) CreateAuth(userID, tenantID uuid.UUID, role string, td *TokenDetails) error {
	return s.sessions.CreateTenantSession(userID, tenantID, role, td)
}

// ExtractToken ...
func (s *Service) ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	//normally Authorization the_token_xxx
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

// VerifyToken ...
func (s *Service) VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := s.ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secrets.Access), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// ExtractTokenMetadata ...
func (s *Service) ExtractTokenMetadata(r *http.Request) (*AccessDetails, error) {
	token, err := s.VerifyToken(r)
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
func (s *Service) FetchAuth(authD *AccessDetails) (uuid.UUID, error) {
	return s.sessions.FetchTenantSession(authD.AccessUUID, authD.UserID)
}

// DeleteAuth consumes one session key and removes it from the user's session
// index.
func (s *Service) DeleteAuth(givenUUID string, userID uuid.UUID) (int64, error) {
	return s.sessions.DeleteTenantSession(givenUUID, userID)
}

// RevokeUserSessions deletes every active session of a user (role change,
// administrative logout).
func (s *Service) RevokeUserSessions(userID uuid.UUID) error {
	return s.sessions.RevokeUserSessions(userID)
}

// ConsumeRefreshAuth atomically verifies the stored owner and softly consumes
// a refresh token (design decision 7). The record is never deleted by the
// consumption: a first consumption marks it "consumed:1" (delivered) within a
// reduced TTL, a replay of a not-delivered record ("consumed:0", released by
// the failed delivery) claims the single recovery, and any other record
// (delivered, recovered, missing, mismatched owner or unknown state) is
// refused. Only the owner segment of the value is compared, so the stored
// format stays robust across role changes between a login and a refresh.
// Returns 1 when the first consumption is claimed, 2 when the recovery is
// claimed and 0 when the record is refused.
func (s *Service) ConsumeRefreshAuth(refreshUUID string, userID uuid.UUID) (int64, error) {
	code, err := s.sessions.ConsumeRefresh(refreshUUID, userID)
	if err != nil {
		return 0, err
	}
	// The session incident stays observable at the marking and at the
	// recovery (support-facing observability, decision 7).
	switch code {
	case 1:
		log.Printf("auth: refresh token %s consumed (delivery pending), user %s", refreshUUID, userID)
	case 2:
		log.Printf("auth: refresh token %s recovered (re-emission), user %s", refreshUUID, userID)
	}
	return code, nil
}

// ReleaseRefresh downgrades a claimed-but-not-delivered consumption: the
// consuming request failed before emitting the new pair, so the replayed
// token stays recoverable during the remaining grace. Returns 1 when the
// release happened and 0 when the record did not allow it.
func (s *Service) ReleaseRefresh(refreshUUID string, userID uuid.UUID) (int64, error) {
	return s.sessions.ReleaseRefresh(refreshUUID, userID)
}

// StartSweep runs the in-process sweep ticker until the returned stop
// function is called. Every tick counts the present consumed records for
// observability and defensively purges the relics the TTL failed to evict
// (e.g. an incident removed the expiry). It is never the purge dependency:
// TTL eviction stays primary, and the sweep never touches an active record or
// one whose grace is still running.
func (s *Service) StartSweep(interval time.Duration) (stop func()) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				consumed, purged, err := s.sessions.SweepConsumed()
				if err != nil {
					log.Printf("auth: session sweep failed: %v", err)
					continue
				}
				if consumed > 0 || purged > 0 {
					log.Printf("auth: session sweep: %d consumed record(s) present, %d relic(s) purged", consumed, purged)
				}
			}
		}
	}()
	return func() {
		cancel()
		<-done
	}
}

// SweepConsumed counts the present consumed session records and purges the
// relics whose TTL anomaly makes them unevictable.
func (s *Service) SweepConsumed() (consumed, purged int64, err error) {
	return s.sessions.SweepConsumed()
}

// CreateAdminToken issues the single superadmin access token: signed with the
// dedicated ADMIN_ACCESS_SECRET, ~8h TTL, no refresh.
func (s *Service) CreateAdminToken(adminID uuid.UUID) (*AdminTokenDetails, error) {
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
	td.AccessToken, err = at.SignedString([]byte(s.secrets.AdminAccess))
	if err != nil {
		return nil, err
	}
	return td, nil
}

// CreateAdminAuth stores the admin session under the admin: Redis prefix.
func (s *Service) CreateAdminAuth(adminID uuid.UUID, td *AdminTokenDetails) error {
	return s.sessions.CreateAdminSession(adminID, td)
}

// ExtractAdminTokenMetadata validates an admin token against the ADMIN_ACCESS
// SECRET: a tenant token carries a different signature and no admin_id claim.
func (s *Service) ExtractAdminTokenMetadata(r *http.Request) (*AdminAccessDetails, error) {
	tokenString := s.ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secrets.AdminAccess), nil
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
func (s *Service) FetchAdmin(authD *AdminAccessDetails) (uuid.UUID, error) {
	return s.sessions.FetchAdminSession(authD.AccessUUID, authD.AdminID)
}

// DeleteAdminAuth consumes an admin session: immediate logout, no refresh.
func (s *Service) DeleteAdminAuth(givenUUID string) (int64, error) {
	return s.sessions.DeleteAdminSession(givenUUID)
}
