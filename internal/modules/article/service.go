package article

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is returned when an update or a delete touches no row: the
// article does not exist or belongs to another tenant (RLS invisible).
var ErrNotFound = errors.New("record not found")

// ArticleRepository is the port the service consumes: tenant-scoped CRUD on
// the article resource.
type ArticleRepository interface {
	Create(ctx context.Context, tenantID, userID uuid.UUID, form CreateArticleForm) (uuid.UUID, error)
	One(ctx context.Context, tenantID, userID, id uuid.UUID) (ArticleResponse, error)
	All(ctx context.Context, tenantID, userID uuid.UUID) ([]Result, error)
	Update(ctx context.Context, tenantID, userID, id uuid.UUID, form CreateArticleForm) error
	Delete(ctx context.Context, tenantID, userID, id uuid.UUID) error
}

// Service orchestrates the article use cases on top of its repository port.
type Service struct {
	articles ArticleRepository
}

// NewService builds the article service on the given repository.
func NewService(articles ArticleRepository) *Service {
	return &Service{articles: articles}
}

// Create stores a new article for the tenant and returns its id.
func (s *Service) Create(ctx context.Context, tenantID, userID uuid.UUID, form CreateArticleForm) (uuid.UUID, error) {
	return s.articles.Create(ctx, tenantID, userID, form)
}

// One returns a single article of the tenant.
func (s *Service) One(ctx context.Context, tenantID, userID, id uuid.UUID) (ArticleResponse, error) {
	return s.articles.One(ctx, tenantID, userID, id)
}

// All returns the tenant articles under the list envelope.
func (s *Service) All(ctx context.Context, tenantID, userID uuid.UUID) ([]Result, error) {
	return s.articles.All(ctx, tenantID, userID)
}

// Update rewrites an existing article of the tenant.
func (s *Service) Update(ctx context.Context, tenantID, userID, id uuid.UUID, form CreateArticleForm) error {
	return s.articles.Update(ctx, tenantID, userID, id, form)
}

// Delete removes an existing article of the tenant.
func (s *Service) Delete(ctx context.Context, tenantID, userID, id uuid.UUID) error {
	return s.articles.Delete(ctx, tenantID, userID, id)
}

// newArticleResponse maps a queried row to the API response shape.
func newArticleResponse(id uuid.UUID, title, content string, createdAt, updatedAt time.Time, authorID uuid.UUID, authorName, authorEmail string) ArticleResponse {
	return ArticleResponse{
		ID:        id.String(),
		Title:     title,
		Content:   content,
		CreatedAt: createdAt.Unix(),
		UpdatedAt: updatedAt.Unix(),
		User: UserResponse{
			ID:    authorID.String(),
			Name:  authorName,
			Email: authorEmail,
		},
	}
}
