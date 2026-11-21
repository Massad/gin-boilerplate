-- name: CreateArticle :one
INSERT INTO article (tenant_id, user_id, title, content)
VALUES (@tenant_id, @user_id, @title, @content)
RETURNING id;

-- name: GetArticle :one
SELECT a.id, a.title, a.content, a.created_at, a.updated_at,
       u.id AS author_id, u.name AS author_name, u.email AS author_email
FROM article a
JOIN "user" u ON u.id = a.user_id
WHERE a.tenant_id = @tenant_id AND a.user_id = @user_id AND a.id = @id
LIMIT 1;

-- name: ListArticles :many
SELECT a.id, a.title, a.content, a.created_at, a.updated_at,
       u.id AS author_id, u.name AS author_name, u.email AS author_email
FROM article a
JOIN "user" u ON u.id = a.user_id
WHERE a.tenant_id = @tenant_id AND a.user_id = @user_id
ORDER BY a.created_at DESC, a.id DESC;

-- name: CountArticles :one
SELECT count(*) FROM article
WHERE tenant_id = @tenant_id AND user_id = @user_id;

-- name: UpdateArticle :execrows
UPDATE article SET title = @title, content = @content
WHERE id = @id AND user_id = @user_id AND tenant_id = @tenant_id;

-- name: DeleteArticle :execrows
DELETE FROM article WHERE id = @id AND user_id = @user_id AND tenant_id = @tenant_id;
