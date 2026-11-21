-- +goose Up

-- Cluster roles, created ahead of any grants. They must be plain logins (not
-- superusers) so that the row level security below is actually exercised: the
-- production passwords are set with ALTER ROLE ... PASSWORD by the operators.
-- +goose StatementBegin
DO $$ BEGIN
    CREATE ROLE app_user LOGIN PASSWORD 'app_user_dev_password';
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
    CREATE ROLE app_admin LOGIN PASSWORD 'app_admin_dev_password' BYPASSRLS;
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
-- +goose StatementEnd

-- Tenants are the isolation root: every multi-tenant row carries tenant_id.
CREATE TABLE tenant (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug       varchar     NOT NULL UNIQUE,
    name       varchar     NOT NULL,
    status     smallint    NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_tenant_updated_at BEFORE UPDATE ON tenant FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE "user"
    ADD COLUMN tenant_id uuid NOT NULL REFERENCES tenant(id) ON DELETE CASCADE,
    ADD COLUMN role varchar(16) NOT NULL DEFAULT 'member' CHECK (role IN ('owner', 'member'));

-- (id, tenant_id) allows the article composite foreign key below: a user_id
-- must never point at a user of another tenant.
ALTER TABLE "user" ADD CONSTRAINT user_id_tenant UNIQUE (id, tenant_id);
CREATE UNIQUE INDEX user_one_owner_idx ON "user"(tenant_id) WHERE role = 'owner';
CREATE INDEX user_tenant_idx ON "user"(tenant_id);

ALTER TABLE article
    ADD COLUMN tenant_id uuid NOT NULL REFERENCES tenant(id) ON DELETE CASCADE;
ALTER TABLE article DROP CONSTRAINT IF EXISTS article_user_id_fkey;
CREATE INDEX article_tenant_user_idx ON article(tenant_id, user_id);
ALTER TABLE article ADD CONSTRAINT article_user_fk
    FOREIGN KEY (tenant_id, user_id) REFERENCES "user"(tenant_id, id)
    ON UPDATE CASCADE ON DELETE CASCADE;

-- Platform backoffice, kept outside of the tenant isolation.
CREATE TABLE admin_user (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email      varchar     NOT NULL UNIQUE,
    password   varchar     NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_admin_user_updated_at BEFORE UPDATE ON admin_user FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Insert-only trail. tenant_id carries no foreign key on purpose: the audit
-- must survive the deletion of a tenant.
CREATE TABLE admin_audit (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id   uuid        NOT NULL REFERENCES admin_user(id) ON DELETE CASCADE,
    tenant_id  uuid,
    action     varchar     NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Row level security, fail closed. Only tenant data is protected: tenant
-- (public directory) and the admin_* tables are platform tables reached through
-- the app_admin pool.
ALTER TABLE article ENABLE ROW LEVEL SECURITY;
ALTER TABLE article FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON article
    USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

ALTER TABLE "user" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "user" FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON "user"
    USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

GRANT USAGE ON SCHEMA public TO app_user, app_admin;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user, app_admin;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_user, app_admin;

-- +goose Down
REVOKE ALL ON SCHEMA public FROM app_user, app_admin;
REVOKE SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public FROM app_user, app_admin;
REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public FROM app_user, app_admin;

DROP POLICY tenant_isolation ON article;
ALTER TABLE article NO FORCE ROW LEVEL SECURITY;
ALTER TABLE article DISABLE ROW LEVEL SECURITY;
DROP POLICY tenant_isolation ON "user";
ALTER TABLE "user" NO FORCE ROW LEVEL SECURITY;
ALTER TABLE "user" DISABLE ROW LEVEL SECURITY;

DROP TABLE admin_audit;
DROP TABLE admin_user;

DROP INDEX article_tenant_user_idx;
ALTER TABLE article DROP CONSTRAINT article_user_fk;
ALTER TABLE article DROP COLUMN tenant_id;
ALTER TABLE article ADD CONSTRAINT article_user_id FOREIGN KEY (user_id) REFERENCES "user"(id) ON UPDATE CASCADE ON DELETE CASCADE;

DROP INDEX user_one_owner_idx;
DROP INDEX user_tenant_idx;
ALTER TABLE "user" DROP CONSTRAINT user_id_tenant;
ALTER TABLE "user" DROP COLUMN role;
ALTER TABLE "user" DROP COLUMN tenant_id;

DROP TRIGGER trg_tenant_updated_at ON tenant;
DROP TABLE tenant;
