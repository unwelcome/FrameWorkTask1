\connect app_db;

CREATE TABLE "users" (
    "id" SERIAL PRIMARY KEY,
    "login" varchar(255) NOT NULL UNIQUE,
    "password_hash" varchar(255) NOT NULL,
    "created_at" timestamp NOT NULL DEFAULT NOW()
);

CREATE TABLE "cats" (
    "id" SERIAL PRIMARY KEY,
    "name" varchar(255) NOT NULL,
    "age" integer,
    "description" text,
    "created_by" integer,
    "created_at" timestamp NOT NULL DEFAULT NOW()
);

CREATE TABLE "cat_photos" (
    "id" SERIAL PRIMARY KEY,
    "cat_id" integer NOT NULL,
    "url" text NOT NULL,
    "filename" text UNIQUE,
    "filesize" integer,
    "mime_type" varchar(255),
    "created_at" timestamp NOT NULL DEFAULT NOW(),
    "is_primary" bool DEFAULT false
);

CREATE INDEX idx_users_login ON users(login);
CREATE INDEX idx_cat_photos_cat_id ON cat_photos(cat_id);
CREATE INDEX idx_cat_photos_primary ON cat_photos(cat_id, is_primary);

ALTER TABLE "cats" ADD CONSTRAINT "cats_to_users" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE CASCADE;
ALTER TABLE "cat_photos" ADD CONSTRAINT "cat_photos_to_cats" FOREIGN KEY ("cat_id") REFERENCES "cats" ("id") ON DELETE CASCADE;
