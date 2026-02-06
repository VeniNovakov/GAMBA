-- Modify "games" table
ALTER TABLE "public"."games" ALTER COLUMN "min_bet" TYPE bigint, ALTER COLUMN "max_bet" TYPE bigint, ALTER COLUMN "house_edge" TYPE bigint;
-- Modify "users" table
ALTER TABLE "public"."users" ALTER COLUMN "balance" TYPE bigint;
-- Create "bets" table
CREATE TABLE "public"."bets" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "type" character varying(20) NOT NULL,
  "game_id" uuid NULL,
  "event_id" uuid NULL,
  "outcome_id" uuid NULL,
  "amount" bigint NOT NULL,
  "odds" numeric NOT NULL,
  "status" character varying(20) NULL DEFAULT 'pending',
  "payout" bigint NULL DEFAULT 0,
  "settled_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_bets_event" FOREIGN KEY ("event_id") REFERENCES "public"."events" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "fk_bets_game" FOREIGN KEY ("game_id") REFERENCES "public"."games" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "fk_bets_outcome" FOREIGN KEY ("outcome_id") REFERENCES "public"."event_outcomes" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "fk_bets_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_bets_deleted_at" to table: "bets"
CREATE INDEX "idx_bets_deleted_at" ON "public"."bets" ("deleted_at");
-- Create index "idx_bets_event_id" to table: "bets"
CREATE INDEX "idx_bets_event_id" ON "public"."bets" ("event_id");
-- Create index "idx_bets_game_id" to table: "bets"
CREATE INDEX "idx_bets_game_id" ON "public"."bets" ("game_id");
-- Create index "idx_bets_outcome_id" to table: "bets"
CREATE INDEX "idx_bets_outcome_id" ON "public"."bets" ("outcome_id");
-- Create index "idx_bets_user_id" to table: "bets"
CREATE INDEX "idx_bets_user_id" ON "public"."bets" ("user_id");
-- Create "tickets" table
CREATE TABLE "public"."tickets" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "assigned_to" uuid NULL,
  "subject" text NOT NULL,
  "description" text NOT NULL,
  "status" character varying(20) NULL DEFAULT 'open',
  "priority" character varying(20) NULL DEFAULT 'medium',
  "resolved_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_tickets_assignee" FOREIGN KEY ("assigned_to") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "fk_tickets_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_tickets_assigned_to" to table: "tickets"
CREATE INDEX "idx_tickets_assigned_to" ON "public"."tickets" ("assigned_to");
-- Create index "idx_tickets_deleted_at" to table: "tickets"
CREATE INDEX "idx_tickets_deleted_at" ON "public"."tickets" ("deleted_at");
-- Create index "idx_tickets_user_id" to table: "tickets"
CREATE INDEX "idx_tickets_user_id" ON "public"."tickets" ("user_id");
-- Create "ticket_messages" table
CREATE TABLE "public"."ticket_messages" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "ticket_id" uuid NOT NULL,
  "sender_id" uuid NOT NULL,
  "content" text NOT NULL,
  "created_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_ticket_messages_sender" FOREIGN KEY ("sender_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_tickets_messages" FOREIGN KEY ("ticket_id") REFERENCES "public"."tickets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_ticket_messages_deleted_at" to table: "ticket_messages"
CREATE INDEX "idx_ticket_messages_deleted_at" ON "public"."ticket_messages" ("deleted_at");
-- Create index "idx_ticket_messages_sender_id" to table: "ticket_messages"
CREATE INDEX "idx_ticket_messages_sender_id" ON "public"."ticket_messages" ("sender_id");
-- Create index "idx_ticket_messages_ticket_id" to table: "ticket_messages"
CREATE INDEX "idx_ticket_messages_ticket_id" ON "public"."ticket_messages" ("ticket_id");
-- Create "tournaments" table
CREATE TABLE "public"."tournaments" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  "description" text NULL,
  "status" character varying(20) NULL DEFAULT 'draft',
  "game_id" uuid NULL,
  "entry_fee" bigint NULL DEFAULT 0,
  "prize_pool" bigint NULL DEFAULT 0,
  "max_participants" bigint NOT NULL,
  "starts_at" timestamptz NOT NULL,
  "ends_at" timestamptz NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_tournaments_game" FOREIGN KEY ("game_id") REFERENCES "public"."games" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_tournaments_deleted_at" to table: "tournaments"
CREATE INDEX "idx_tournaments_deleted_at" ON "public"."tournaments" ("deleted_at");
-- Create index "idx_tournaments_game_id" to table: "tournaments"
CREATE INDEX "idx_tournaments_game_id" ON "public"."tournaments" ("game_id");
-- Create "tournament_participants" table
CREATE TABLE "public"."tournament_participants" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "tournament_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "score" numeric NULL DEFAULT 0,
  "rank" bigint NULL DEFAULT 0,
  "prize_won" numeric NULL DEFAULT 0,
  "joined_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_tournament_participants_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_tournaments_participants" FOREIGN KEY ("tournament_id") REFERENCES "public"."tournaments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_tournament_participants_deleted_at" to table: "tournament_participants"
CREATE INDEX "idx_tournament_participants_deleted_at" ON "public"."tournament_participants" ("deleted_at");
-- Create index "idx_tournament_participants_tournament_id" to table: "tournament_participants"
CREATE INDEX "idx_tournament_participants_tournament_id" ON "public"."tournament_participants" ("tournament_id");
-- Create index "idx_tournament_participants_user_id" to table: "tournament_participants"
CREATE INDEX "idx_tournament_participants_user_id" ON "public"."tournament_participants" ("user_id");
-- Create "transactions" table
CREATE TABLE "public"."transactions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "type" character varying(30) NOT NULL,
  "status" character varying(20) NULL DEFAULT 'pending',
  "amount" bigint NOT NULL,
  "reference_id" uuid NULL,
  "reference_type" character varying(30) NULL,
  "description" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_transactions_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_transactions_deleted_at" to table: "transactions"
CREATE INDEX "idx_transactions_deleted_at" ON "public"."transactions" ("deleted_at");
-- Create index "idx_transactions_reference_id" to table: "transactions"
CREATE INDEX "idx_transactions_reference_id" ON "public"."transactions" ("reference_id");
-- Create index "idx_transactions_user_id" to table: "transactions"
CREATE INDEX "idx_transactions_user_id" ON "public"."transactions" ("user_id");
