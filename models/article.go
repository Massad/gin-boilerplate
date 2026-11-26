package models

import (
	"context"
	"errors"
	"time"

	"github.com/Massad/gin-boilerplate/forms"
	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/google/uuid"
)

// AllArticleResponse represents the top-level list structure.
type AllArticleResponse struct {
	Results []Result `json:"results"`
}

// OneArticleResponse wraps a single article.
type OneArticleResponse struct {
	Data ArticleResponse `json:"data"`
}

// Result represents each result item
type Result struct {
	Data []ArticleResponse `json:"data"`
	Meta Meta              `json:"meta"`
}

// ArticleResponse represents an individual article
type ArticleResponse struct {
	ID        string       `json:"id"`
	Title     string       `json:"title"`
	Content   string       `json:"content"`
	UpdatedAt int64        `json:"updated_at"`
	CreatedAt int64        `json:"created_at"`
	User      UserResponse `json:"user"`
}

// UserResponse represents the article's author
type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Meta represents metadata for pagination or total count
type Meta struct {
	Total int `json:"total"`
}

// ArticleModel ...
type ArticleModel struct{}

// ErrNotFound is returned when an update or a delete touches no row: the
// article does not exist or belongs to another tenant (RLS invisible).
var ErrNotFound = errors.New("record not found")

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

// Create inserts the article inside the tenant transaction: the context sets
// the RLS tenant and the row carries the tenant_id column.
func (m ArticleModel) Create(ctx context.Context, tenantID, userID uuid.UUID, form forms.CreateArticleForm) (uuid.UUID, error) {
	var id uuid.UUID
	err := pgdb.TenantStore().WithTenant(ctx, tenantID, func(q *pgdb.Queries) error {
		rowID, err := q.CreateArticle(ctx, pgdb.CreateArticleParams{
			TenantID: tenantID,
			UserID:   userID,
			Title:    form.Title,
			Content:  form.Content,
		})
		if err != nil {
			return err
		}
		id = rowID
		return nil
	})
	return id, err
}

// One ...
func (m ArticleModel) One(ctx context.Context, tenantID, userID, id uuid.UUID) (ArticleResponse, error) {
	var article ArticleResponse
	err := pgdb.TenantStore().WithTenant(ctx, tenantID, func(q *pgdb.Queries) error {
		row, err := q.GetArticle(ctx, pgdb.GetArticleParams{TenantID: tenantID, UserID: userID, ID: id})
		if err != nil {
			return err
		}
		article = newArticleResponse(row.ID, row.Title, row.Content, row.CreatedAt, row.UpdatedAt, row.AuthorID, row.AuthorName, row.AuthorEmail)
		return nil
	})
	return article, err
}

// All ...
func (m ArticleModel) All(ctx context.Context, tenantID, userID uuid.UUID) ([]Result, error) {
	var results []Result
	err := pgdb.TenantStore().WithTenant(ctx, tenantID, func(q *pgdb.Queries) error {
		rows, err := q.ListArticles(ctx, pgdb.ListArticlesParams{TenantID: tenantID, UserID: userID})
		if err != nil {
			return err
		}
		total, err := q.CountArticles(ctx, pgdb.CountArticlesParams{TenantID: tenantID, UserID: userID})
		if err != nil {
			return err
		}
		articles := []ArticleResponse{}
		for _, row := range rows {
			articles = append(articles, newArticleResponse(row.ID, row.Title, row.Content, row.CreatedAt, row.UpdatedAt, row.AuthorID, row.AuthorName, row.AuthorEmail))
		}
		results = []Result{{Data: articles, Meta: Meta{Total: int(total)}}}
		return nil
	})
	return results, err
}

// Update ...
func (m ArticleModel) Update(ctx context.Context, tenantID, userID, id uuid.UUID, form forms.CreateArticleForm) (err error) {
	err = pgdb.TenantStore().WithTenant(ctx, tenantID, func(q *pgdb.Queries) error {
		affected, err := q.UpdateArticle(ctx, pgdb.UpdateArticleParams{
			Title:    form.Title,
			Content:  form.Content,
			ID:       id,
			UserID:   userID,
			TenantID: tenantID,
		})
		if err != nil {
			return err
		}
		if affected == 0 {
			return ErrNotFound
		}
		return nil
	})
	return err
}

// Delete ...
func (m ArticleModel) Delete(ctx context.Context, tenantID, userID, id uuid.UUID) (err error) {
	err = pgdb.TenantStore().WithTenant(ctx, tenantID, func(q *pgdb.Queries) error {
		affected, err := q.DeleteArticle(ctx, pgdb.DeleteArticleParams{ID: id, UserID: userID, TenantID: tenantID})
		if err != nil {
			return err
		}
		if affected == 0 {
			return ErrNotFound
		}
		return nil
	})
	return err
}
