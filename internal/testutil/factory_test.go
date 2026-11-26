package testutil

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/Massad/gin-boilerplate/internal/modules/tenant"
	validate "github.com/Massad/gin-boilerplate/internal/platform/validate"
)

// TestNewTenantFixtureIsSlugValid checks the generated tenant against the
// tenant.ValidSlug contract.
func TestNewTenantFixtureIsSlugValid(t *testing.T) {
	fixture := NewTenantFixture(t)

	require.True(t, tenant.ValidSlug(fixture.Slug), "slug %q must pass ValidSlug", fixture.Slug)
	require.NotZero(t, fixture.ID)
	require.Equal(t, int16(1), fixture.Status)
	require.GreaterOrEqual(t, fixture.CreatedAt, int64(0))
}

// TestNewUserFixtureMatchesValidators runs the production binding validators
// on the generated registration form and checks the struct against the bcrypt
// storage contract.
func TestNewUserFixtureMatchesValidators(t *testing.T) {
	user, form := NewUserFixture(t)

	require.NoError(t, validateForm(t, form))

	require.Equal(t, form.Email, user.Email)
	require.Equal(t, form.Name, user.Name)
	require.Equal(t, "owner", user.Role)
	require.NotZero(t, user.TenantID)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password)), "the struct hash must verify the form password")
}

// TestNewUserFixtureDeterministic replays the same test name and expects the
// exact same seeded output (modulo the random bcrypt salt).
func TestNewUserFixtureDeterministic(t *testing.T) {
	userA, formA := NewUserFixture(t)
	userB, formB := NewUserFixture(t)

	require.Equal(t, formA, formB)
	require.Equal(t, userA.ID, userB.ID)
	require.Equal(t, userA.Email, userB.Email)
	require.Equal(t, userA.Name, userB.Name)
	require.Equal(t, userA.TenantID, userB.TenantID)
	require.Equal(t, userA.Role, userB.Role)
}

// TestNewAdminFixture checks the seeded superadmin credentials.
func TestNewAdminFixture(t *testing.T) {
	email, password := NewAdminFixture(t)

	require.NotEmpty(t, email)
	require.GreaterOrEqual(t, runeLen(password), 8)
	require.LessOrEqual(t, runeLen(password), 50)
}

// TestNewArticleFixtureChecksBounds runs the production article form
// validators on the generated form.
func TestNewArticleFixtureChecksBounds(t *testing.T) {
	tenantID, userID := uuid.New(), uuid.New()
	form := NewArticleFixture(t, tenantID, userID)

	require.NoError(t, validateForm(t, form))
}

// TestNewArticleFixtureDeterministic replays the same test name and expects
// the exact same seeded form.
func TestNewArticleFixtureDeterministic(t *testing.T) {
	formA := NewArticleFixture(t, uuid.New(), uuid.New())
	formB := NewArticleFixture(t, uuid.New(), uuid.New())

	require.Equal(t, formA, formB)
}

// validateForm runs the same validate.DefaultValidator the routes use, so the
// factories cannot silently drift from the real validators.
func validateForm(t *testing.T, form interface{}) error {
	t.Helper()
	binding.Validator = new(validate.DefaultValidator)
	return binding.Validator.ValidateStruct(form)
}

// runeLen counts characters the way the binding validators do.
func runeLen(value string) int { return len([]rune(value)) }
