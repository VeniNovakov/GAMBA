-- Create "events" table
CREATE TABLE "public"."events" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  "description" text NULL,
  "category" character varying(30) NOT NULL,
  "status" character varying(20) NULL DEFAULT 'upcoming',
  "starts_at" timestamptz NOT NULL,
  "ends_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_events_deleted_at" to table: "events"
CREATE INDEX "idx_events_deleted_at" ON "public"."events" ("deleted_at");
-- Create "refresh_tokens" table
CREATE TABLE "public"."refresh_tokens" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "token_hash" text NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "is_revoked" boolean NULL DEFAULT false,
  "revoked_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_refresh_tokens_deleted_at" to table: "refresh_tokens"
CREATE INDEX "idx_refresh_tokens_deleted_at" ON "public"."refresh_tokens" ("deleted_at");
-- Create index "idx_refresh_tokens_user_id" to table: "refresh_tokens"
CREATE INDEX "idx_refresh_tokens_user_id" ON "public"."refresh_tokens" ("user_id");
-- Create "event_outcomes" table
CREATE TABLE "public"."event_outcomes" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "event_id" uuid NOT NULL,
  "name" text NOT NULL,
  "odds" numeric NOT NULL,
  "is_winner" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_events_outcomes" FOREIGN KEY ("event_id") REFERENCES "public"."events" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_event_outcomes_deleted_at" to table: "event_outcomes"
CREATE INDEX "idx_event_outcomes_deleted_at" ON "public"."event_outcomes" ("deleted_at");
-- Create index "idx_event_outcomes_event_id" to table: "event_outcomes"
CREATE INDEX "idx_event_outcomes_event_id" ON "public"."event_outcomes" ("event_id");
