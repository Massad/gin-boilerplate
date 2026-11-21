-- name: GetUserForLogin :one
SELECT u.id, u.email, u.name, u.password, u.tenant_id, u.role,
       t.slug, t.name AS tenant_name, t.status AS tenant_status
FROM "user" u
JOIN tenant t ON t.id = u.tenant_id
WHERE u.email = LOWER(@email) LIMIT 1;

-- name: GetUserWithTenant :one
SELECT u.id, u.email, u.name, u.password, u.tenant_id, u.role,
       t.slug, t.name AS tenant_name, t.status AS tenant_status
FROM "user" u
JOIN tenant t ON t.id = u.tenant_id
WHERE u.id = @id LIMIT 1;

-- name: CreateUserOwner :one
INSERT INTO "user" (email, password, name, tenant_id, role)
VALUES (LOWER(@email), @password, @name, @tenant_id, 'owner')
ON CONFLICT (email) DO NOTHING
RETURNING id;

-- name: GetUserByID :one
SELECT id, email, name FROM "user" WHERE id = @id LIMIT 1;
