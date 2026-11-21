package models

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	pgdb "github.com/Massad/gin-boilerplate/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"golang.org/x/text/unicode/norm"
)

// Tenant represents one organization sharing the platform database.
type Tenant struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Status    int16     `json:"status"`
	CreatedAt int64     `json:"-"`
	UpdatedAt int64     `json:"-"`
}

// AllTenantResponse wraps the tenant list for the backoffice API.
type AllTenantResponse struct {
	Tenants []Tenant `json:"tenants"`
}

// ErrInvalidSlug is returned when a company name does not map to a usable slug.
var ErrInvalidSlug = errors.New("the company name does not map to a valid slug")

// ErrSlugUnavailable is returned when every variation of the slug is taken.
var ErrSlugUnavailable = errors.New("a slug could not be generated for this company name")

// reservedSlugs lists route roots already used by the platform: a tenant path
// must never collide with them.
var reservedSlugs = map[string]bool{
	"v1":      true,
	"admin":   true,
	"public":  true,
	"swagger": true,
	"api":     true,
	"login":   true,
	"app":     true,
	"sw":      true,
	"static":  true,
	"www":     true,
	"user":    true,
	"article": true,
	"invoice": true,
	"token":   true,
}

// tenantSlugPattern bounds a slug between 3 and 40 characters, letters, digits
// and dashes only, without leading or trailing dash.
var tenantSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,38}[a-z0-9]$`)

// maxSlugLength is the longest slug accepted by tenantSlugPattern.
const maxSlugLength = 40

// maxSlugAttempts bounds the -2, -3, ... suffix search before giving up.
const maxSlugAttempts = 99

// Slugify turns a company name into its normalized slug: lowercase, without
// accents, with any other character run collapsed into a single dash.
func Slugify(name string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(strings.ToLower(name)) {
		switch {
		case unicode.Is(unicode.Mn, r): // combining mark: part of an accent
			continue
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	slug := dashRuns.ReplaceAllString(b.String(), "-")
	return strings.Trim(slug, "-")
}

// dashRuns collapses consecutive separator characters into a single dash.
var dashRuns = regexp.MustCompile("-{2,}")

// ValidSlug reports whether the slug is well formed and not reserved.
func ValidSlug(slug string) bool {
	return tenantSlugPattern.MatchString(slug) && !reservedSlugs[slug]
}

// TenantModel ...
type TenantModel struct{}

var tenantModel = new(TenantModel)

// Create creates a tenant with a slug derived from name, suffixed -2, -3, ...
// when the slug is already taken. It runs on the platform pool since tenant
// registration is not a tenant-scoped operation.
func (m TenantModel) Create(name string) (Tenant, error) {
	ctx := context.Background()
	var tenant Tenant
	err := pgdb.AdminStore().InTx(ctx, func(tx pgx.Tx, q *pgdb.Queries) error {
		var err error
		tenant, err = createTenantInTx(ctx, q, name)
		return err
	})
	if err != nil {
		return Tenant{}, err
	}
	return tenant, nil
}

// createTenantInTx inserts the tenant inside the ongoing platform transaction,
// retrying on slug collisions with the standard candidates.
func createTenantInTx(ctx context.Context, q *pgdb.Queries, name string) (Tenant, error) {
	base := Slugify(name)
	if !ValidSlug(base) {
		return Tenant{}, ErrInvalidSlug
	}
	for attempt := 1; attempt <= maxSlugAttempts; attempt++ {
		candidate := base
		if attempt > 1 {
			candidate = shortSlugCandidate(base, attempt)
		}
		row, err := q.CreateTenant(ctx, pgdb.CreateTenantParams{Slug: candidate, Name: name})
		if err == nil {
			return newTenant(row.ID, row.Slug, row.Name, row.Status, row.CreatedAt, row.UpdatedAt), nil
		}
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
			return Tenant{}, err
		}
	}
	return Tenant{}, ErrSlugUnavailable
}

// shortSlugCandidate builds suffixed candidates that keep the slug within the
// maximum length accepted by the pattern.
func shortSlugCandidate(base string, attempt int) string {
	suffix := "-" + strconv.Itoa(attempt)
	if len(base)+len(suffix) > maxSlugLength {
		base = strings.Trim(base[:maxSlugLength-len(suffix)], "-")
	}
	return base + suffix
}

// OneBySlug ...
func (m TenantModel) OneBySlug(slug string) (Tenant, error) {
	row, err := pgdb.TenantStore().GetTenantBySlug(context.Background(), slug)
	if err != nil {
		return Tenant{}, err
	}
	return newTenant(row.ID, row.Slug, row.Name, row.Status, row.CreatedAt, row.UpdatedAt), nil
}

// All lists the tenants for the platform backoffice.
func (m TenantModel) All() ([]Tenant, error) {
	rows, err := pgdb.AdminStore().ListTenants(context.Background())
	if err != nil {
		return nil, err
	}
	tenants := []Tenant{}
	for _, row := range rows {
		tenants = append(tenants, newTenant(row.ID, row.Slug, row.Name, row.Status, row.CreatedAt, row.UpdatedAt))
	}
	return tenants, nil
}

// SetStatus suspends (or reactivates) a tenant from the platform backoffice.
func (m TenantModel) SetStatus(id uuid.UUID, status int16) error {
	return pgdb.AdminStore().UpdateTenantStatus(context.Background(), pgdb.UpdateTenantStatusParams{ID: id, Status: status})
}

// newTenant maps a generated row to the API struct.
func newTenant(id uuid.UUID, slug, name string, status int16, createdAt, updatedAt time.Time) Tenant {
	return Tenant{
		ID:        id,
		Slug:      slug,
		Name:      name,
		Status:    status,
		CreatedAt: createdAt.Unix(),
		UpdatedAt: updatedAt.Unix(),
	}
}
