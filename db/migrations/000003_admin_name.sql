-- +goose Up

-- The unified login salutation needs a stored display name. The DEFAULT keeps
-- the migration retro-compatible with admins created before the column.
-- Platform table: admin_user carries no RLS.
ALTER TABLE admin_user ADD COLUMN name varchar NOT NULL DEFAULT 'Superadmin';

-- +goose Down
ALTER TABLE admin_user DROP COLUMN name;
