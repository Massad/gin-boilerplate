-- name: CreateTenant :one
INSERT INTO tenant (slug, name)
VALUES (@slug, @name)
RETURNING id, slug, name, status, created_at, updated_at;

-- name: GetTenantBySlug :one
SELECT * FROM tenant WHERE slug = @slug LIMIT 1;

-- name: ListTenants :many
SELECT * FROM tenant
ORDER BY created_at DESC, id;

-- name: UpdateTenantStatus :exec
UPDATE tenant SET status = @status WHERE id = @id;

-- name: GetTenantByID :one
SELECT * FROM tenant WHERE id = @id LIMIT 1;
