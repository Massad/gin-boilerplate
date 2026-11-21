-- name: GetAdminByEmail :one
SELECT id, email, password
FROM admin_user
WHERE email = LOWER(@email) LIMIT 1;

-- name: CreateAdmin :one
INSERT INTO admin_user (email, password)
VALUES (LOWER(@email), @password)
ON CONFLICT (email) DO UPDATE SET password = EXCLUDED.password, updated_at = now()
RETURNING id;

-- name: InsertAdminAudit :exec
INSERT INTO admin_audit (admin_id, tenant_id, action)
VALUES (@admin_id, sqlc.narg('tenant_id'), @action);
