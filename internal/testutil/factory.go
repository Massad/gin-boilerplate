// Package testutil carries the test-only helpers shared by the suites:
// seeded gofakeit factories, the testfixtures loader and the opt-in ephemeral
// PostgreSQL bootstrap. It must stay imported by *_test.go files only.
package testutil

import (
	"hash/crc32"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Massad/gin-boilerplate/internal/modules/article"
	"github.com/Massad/gin-boilerplate/internal/modules/tenant"
	user "github.com/Massad/gin-boilerplate/internal/modules/user"
)

// seed pins gofakeit to the test name so every generated failure replays exactly.
func seed(t *testing.T) {
	t.Helper()
	// gofakeit reseeds from crypto randomness on 0, which would silently break
	// reproducibility for a name hashing to 0.
	seedValue := int64(crc32.ChecksumIEEE([]byte(t.Name())))
	if seedValue == 0 {
		seedValue = 1
	}
	if err := gofakeit.Seed(seedValue); err != nil {
		t.Fatal(err)
	}
}

// NewTenantFixture returns a seeded tenant whose slug passes tenant.ValidSlug.
func NewTenantFixture(t *testing.T) tenant.Tenant {
	t.Helper()
	seed(t)
	var name, slug string
	for attempt := 0; attempt < 25; attempt++ {
		name = gofakeit.Company()
		slug = tenant.Slugify(name)
		if tenant.ValidSlug(slug) {
			return tenant.Tenant{
				ID:        uuid.MustParse(gofakeit.UUID()),
				Slug:      slug,
				Name:      name,
				Status:    1,
				CreatedAt: gofakeit.Date().Unix(),
				UpdatedAt: gofakeit.Date().Unix(),
			}
		}
	}
	t.Fatalf("no valid tenant slug after 25 seeded attempts")
	return tenant.Tenant{}
}

// NewUserFixture returns a seeded owner together with the registration form
// that generates it: the form fields pass the binding validators and the
// struct carries the bcrypt hash for direct row creation.
func NewUserFixture(t *testing.T) (user.User, user.RegisterForm) {
	t.Helper()
	seed(t)

	password := clampGenerated(gofakeit.Password(true, true, true, false, false, 16), 8, 50)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	name := clampGenerated(gofakeit.Name(), 3, 20)
	form := user.RegisterForm{
		Name:     name,
		Email:    gofakeit.Email(),
		Password: password,
		Company:  clampGenerated(gofakeit.Company(), 3, 60),
	}
	usr := user.User{
		ID:       uuid.MustParse(gofakeit.UUID()),
		Email:    form.Email,
		Password: string(passwordHash),
		Name:     form.Name,
		TenantID: uuid.MustParse(gofakeit.UUID()),
		Role:     "owner",
	}
	return usr, form
}

// NewAdminFixture returns a seeded superadmin pair: a unique email and a
// strong password.
func NewAdminFixture(t *testing.T) (email, password string) {
	t.Helper()
	seed(t)
	return gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 16)
}

// NewArticleFixture returns a seeded article form. The tenant and user ids are
// accepted for row-building callers; the form itself carries title and content
// only.
func NewArticleFixture(t *testing.T, tenantID, userID uuid.UUID) article.CreateArticleForm {
	t.Helper()
	seed(t)
	return article.CreateArticleForm{
		Title:   clampGenerated(gofakeit.JobTitle(), 3, 100),
		Content: clampGenerated(gofakeit.Sentence(9), 3, 1000),
	}
}

// clampGenerated bounds a seeded value to the validators' rune-count limits,
// trimming and padding with letters so the result always passes min/max.
func clampGenerated(value string, minLen, maxLen int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) > maxLen {
		runes = runes[:maxLen]
	}
	value = strings.TrimSpace(string(runes))
	for utf8.RuneCountInString(value) < minLen {
		value += gofakeit.Letter()
	}
	return value
}
