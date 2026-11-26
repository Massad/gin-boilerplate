package article

import (
	"context"

	pgdb "github.com/Massad/gin-boilerplate/internal/platform/db"
	"github.com/google/uuid"
)

// repository implements ArticleRepository on the tenant pool store: every
// operation runs inside the RLS tenant transaction.
type repository struct {
	store pgdb.Store
}

// NewRepository builds the tenant-backed article repository. The store is
// resolved once at the composition root and injected here.
func NewRepository(store pgdb.Store) ArticleRepository {
	return repository{store: store}
}

// Create inserts the article inside the tenant transaction: the context sets
// the RLS tenant and the row carries the tenant_id column.
func (r repository) Create(ctx context.Context, tenantID, userID uuid.UUID, form CreateArticleForm) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.store.WithTenant(ctx, tenantID, func(q *pgdb.Queries) error {
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

// One fetches one article of the tenant and maps the row to the response.
func (r repository) One(ctx context.Context, tenantID, userID, id uuid.UUID) (ArticleResponse, error) {
	var article ArticleResponse
	err := r.store.WithTenant(ctx, tenantID, func(q *pgdb.Queries) error {
		row, err := q.GetArticle(ctx, pgdb.GetArticleParams{TenantID: tenantID, UserID: userID, ID: id})
		if err != nil {
			return err
		}
		article = newArticleResponse(row.ID, row.Title, row.Content, row.CreatedAt, row.UpdatedAt, row.AuthorID, row.AuthorName, row.AuthorEmail)
		return nil
	})
	return article, err
}

// All lists the tenant articles and counts them for the list envelope.
func (r repository) All(ctx context.Context, tenantID, userID uuid.UUID) ([]Result, error) {
	var results []Result
	err := r.store.WithTenant(ctx, tenantID, func(q *pgdb.Queries) error {
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

// Update rewrites an article and reports a zero-row update as ErrNotFound.
func (r repository) Update(ctx context.Context, tenantID, userID, id uuid.UUID, form CreateArticleForm) (err error) {
	err = r.store.WithTenant(ctx, tenantID, func(q *pgdb.Queries) error {
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

// Delete removes an article and reports a zero-row delete as ErrNotFound.
func (r repository) Delete(ctx context.Context, tenantID, userID, id uuid.UUID) (err error) {
	err = r.store.WithTenant(ctx, tenantID, func(q *pgdb.Queries) error {
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
