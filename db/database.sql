-- Database: golang_gin_db

-- DROP DATABASE golang_gin_db;

CREATE DATABASE golang_gin_db
  WITH OWNER = postgres
       ENCODING = 'UTF8'
       TABLESPACE = pg_default
       LC_COLLATE = 'en_US.UTF-8'
       LC_CTYPE = 'en_US.UTF-8'
       CONNECTION LIMIT = -1;

-- Table: article

-- DROP TABLE article;

CREATE TABLE article
(
  id serial NOT NULL,
  user_id integer,
  title character varying,
  content text,
  updated_at integer,
  created_at integer,
  CONSTRAINT article_id PRIMARY KEY (id),
  CONSTRAINT article_user_id FOREIGN KEY (user_id)
      REFERENCES "user" (id) MATCH SIMPLE
      ON UPDATE CASCADE ON DELETE CASCADE
)
WITH (
  OIDS=FALSE
);
ALTER TABLE article
  OWNER TO postgres;

  -- Table: "user"

  -- DROP TABLE "user";

  CREATE TABLE "user"
  (
    id serial NOT NULL,
    email character varying,
    password character varying,
    name character varying,
    updated_at integer,
    created_at integer,
    CONSTRAINT user_id PRIMARY KEY (id)
  )
  WITH (
    OIDS=FALSE
  );
  ALTER TABLE "user"
    OWNER TO postgres;
