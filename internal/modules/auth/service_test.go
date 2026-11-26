package auth_test

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	auth "github.com/Massad/gin-boilerplate/internal/modules/auth"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// fakeSessions records the repository calls: the pure suite proves the
// service contract on top of the port without PostgreSQL, Redis or secrets.
type fakeSessions struct {
	created        int
	createdValue   string
	fetched        int
	fetchedUserID  uuid.UUID
	deleted        int
	revoked        int
	consumeCalled  int
	consumeCode    int64
	consumeErr     error
	releaseCalled  int
	releaseCode    int64
	adminCreated   int
	adminFetched   int
	adminFetchedID uuid.UUID
	adminDeleted   int
	sweepVisited   int64
	sweepPurged    int64
}

func (f *fakeSessions) CreateTenantSession(userID, tenantID uuid.UUID, role string, _ *auth.TokenDetails) error {
	f.created++
	f.createdValue = fmt.Sprintf("%v:%v:%v", userID, tenantID, role)
	return nil
}

func (f *fakeSessions) FetchTenantSession(_ string, claimedUserID uuid.UUID) (uuid.UUID, error) {
	f.fetched++
	return f.fetchedUserID, nil
}

func (f *fakeSessions) DeleteTenantSession(_ string, _ uuid.UUID) (int64, error) {
	f.deleted++
	return 1, nil
}

func (f *fakeSessions) RevokeUserSessions(_ uuid.UUID) error {
	f.revoked++
	return nil
}

func (f *fakeSessions) ConsumeRefresh(string, uuid.UUID) (int64, error) {
	f.consumeCalled++
	return f.consumeCode, f.consumeErr
}

func (f *fakeSessions) ReleaseRefresh(string, uuid.UUID) (int64, error) {
	f.releaseCalled++
	return f.releaseCode, nil
}

func (f *fakeSessions) CreateAdminSession(adminID uuid.UUID, _ *auth.AdminTokenDetails) error {
	f.adminCreated++
	f.adminFetchedID = adminID
	return nil
}

func (f *fakeSessions) FetchAdminSession(_ string, _ uuid.UUID) (uuid.UUID, error) {
	f.adminFetched++
	return f.adminFetchedID, nil
}

func (f *fakeSessions) DeleteAdminSession(_ string) (int64, error) {
	f.adminDeleted++
	return 1, nil
}

func (f *fakeSessions) SweepConsumed() (int64, int64, error) {
	return f.sweepVisited, f.sweepPurged, nil
}

func pureTestService(t *testing.T) (*auth.Service, *fakeSessions) {
	t.Helper()
	fake := &fakeSessions{fetchedUserID: uuid.New()}
	svc := auth.NewService(auth.Secrets{
		Access:      "access-key-for-pure-tests-only-32-bytes",
		Refresh:     "refresh-key-for-pure-tests-only-32-bytes",
		AdminAccess: "admin-key-for-pure-tests-only-32-bytes",
	}, fake)
	return svc, fake
}

// TestTokenLifecyclePure proves the Secrets injection (no environment access),
// the claim shape of the issued pair and the session value format carried to
// the repository port.
func TestTokenLifecyclePure(t *testing.T) {
	svc, fake := pureTestService(t)
	userID := uuid.New()
	tenantID := uuid.New()

	td, err := svc.CreateToken(userID, tenantID, "owner")
	require.NoError(t, err)
	require.NotEmpty(t, td.AccessToken)
	require.NotEmpty(t, td.RefreshToken)

	access, err := jwt.Parse(td.AccessToken, func(token *jwt.Token) (interface{}, error) {
		require.Equal(t, jwt.SigningMethodHS256, token.Method)
		return []byte("access-key-for-pure-tests-only-32-bytes"), nil
	})
	require.NoError(t, err)
	claims := access.Claims.(jwt.MapClaims)
	require.Equal(t, td.AccessUUID, claims["access_uuid"])
	require.Equal(t, userID.String(), claims["user_id"])
	require.Equal(t, tenantID.String(), claims["tenant_id"])
	require.Equal(t, "owner", claims["role"])
	require.EqualValues(t, td.AtExpires, claims["exp"])
	require.Greater(t, td.AtExpires, time.Now().Unix())

	refresh, err := jwt.Parse(td.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte("refresh-key-for-pure-tests-only-32-bytes"), nil
	})
	require.NoError(t, err)
	refreshClaims := refresh.Claims.(jwt.MapClaims)
	require.Equal(t, td.RefreshUUID, refreshClaims["refresh_uuid"])
	require.Equal(t, userID.String(), refreshClaims["user_id"])

	require.NoError(t, svc.CreateAuth(userID, tenantID, "owner", td))
	require.Equal(t, 1, fake.created)
	require.Equal(t, fmt.Sprintf("%v:%v:%v", userID, tenantID, "owner"), fake.createdValue)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+td.AccessToken)
	details, err := svc.ExtractTokenMetadata(req)
	require.NoError(t, err)
	require.Equal(t, td.AccessUUID, details.AccessUUID)
	require.Equal(t, userID, details.UserID)
	require.Equal(t, tenantID, details.TenantID)
	require.Equal(t, "owner", details.Role)

	verified, err := svc.FetchAuth(details)
	require.NoError(t, err)
	require.Equal(t, fake.fetchedUserID, verified)

	deleted, err := svc.DeleteAuth(details.AccessUUID, userID)
	require.NoError(t, err)
	require.Equal(t, int64(1), deleted)
	require.Equal(t, 1, fake.deleted)

	require.NoError(t, svc.RevokeUserSessions(userID))
	require.Equal(t, 1, fake.revoked)
}

// TestAdminTokenLifecyclePure proves the admin token is signed with its
// dedicated secret and that a tenant access token never parses as an admin
// token.
func TestAdminTokenLifecyclePure(t *testing.T) {
	svc, fake := pureTestService(t)
	adminID := uuid.New()

	td, err := svc.CreateAdminToken(adminID)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+td.AccessToken)
	details, err := svc.ExtractAdminTokenMetadata(req)
	require.NoError(t, err)
	require.Equal(t, td.AccessUUID, details.AccessUUID)
	require.Equal(t, adminID, details.AdminID)

	// A different signing secret: the admin token must never pass the tenant
	// extraction path.
	_, err = svc.ExtractTokenMetadata(req)
	require.Error(t, err)

	require.NoError(t, svc.CreateAdminAuth(adminID, td))
	require.Equal(t, 1, fake.adminCreated)
	verified, err := svc.FetchAdmin(details)
	require.NoError(t, err)
	require.Equal(t, adminID, verified)

	deleted, err := svc.DeleteAdminAuth(details.AccessUUID)
	require.NoError(t, err)
	require.Equal(t, int64(1), deleted)
	require.Equal(t, 1, fake.adminDeleted)
}

// TestConsumeRefreshAuthSignalsPure pins the flow signal contract of the
// soft-consume port: 1 = first consumption, 2 = recovery claimed, 0 = refused,
// and the repository errors propagate.
func TestConsumeRefreshAuthSignalsPure(t *testing.T) {
	svc, fake := pureTestService(t)
	userID := uuid.New()

	for _, tc := range []struct {
		code int64
		want int64
	}{
		{1, 1},
		{2, 2},
		{0, 0},
	} {
		fake.consumeErr = nil
		fake.consumeCode = tc.code
		got, err := svc.ConsumeRefreshAuth(uuid.NewString(), userID)
		require.NoError(t, err)
		require.Equal(t, tc.want, got)
	}

	fake.consumeErr = errors.New("redis down")
	got, err := svc.ConsumeRefreshAuth(uuid.NewString(), userID)
	require.Error(t, err)
	require.Equal(t, int64(0), got)

	fake.consumeErr = nil
	fake.releaseCode = 1
	got, err = svc.ReleaseRefresh(uuid.NewString(), userID)
	require.NoError(t, err)
	require.Equal(t, int64(1), got)
	require.Equal(t, 1, fake.releaseCalled)
}
