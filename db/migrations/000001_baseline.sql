-- +goose Up

CREATE TABLE "user" (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email      varchar     NOT NULL UNIQUE,
    password   varchar     NOT NULL,
    name       varchar     NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE article (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES "user"(id) ON UPDATE CASCADE ON DELETE CASCADE,
    title      varchar     NOT NULL,
    content    text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX article_user_idx ON article(user_id);

-- +goose StatementBegin
CREATE FUNCTION set_updated_at() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END $$;
-- +goose StatementEnd

CREATE TRIGGER trg_article_updated_at BEFORE UPDATE ON article FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_user_updated_at    BEFORE UPDATE ON "user" FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TABLE article;
DROP TABLE "user";
DROP FUNCTION set_updated_at();
