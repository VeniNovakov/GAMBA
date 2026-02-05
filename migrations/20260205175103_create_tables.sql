-- Create "users" table
CREATE TABLE "public"."users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  "password_hash" text NOT NULL,
  "role" character varying(20) NULL DEFAULT 'player',
  "balance" numeric NULL DEFAULT 0,
  "is_active" boolean NULL DEFAULT true,
  "is_restricted" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_users_deleted_at" to table: "users"
CREATE INDEX "idx_users_deleted_at" ON "public"."users" ("deleted_at");
-- Create index "idx_users_name" to table: "users"
CREATE UNIQUE INDEX "idx_users_name" ON "public"."users" ("name");
