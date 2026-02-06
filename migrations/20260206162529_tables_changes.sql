-- Modify "bets" table
ALTER TABLE "public"."bets" ALTER COLUMN "odds" TYPE bigint;
-- Drop index "idx_chats_user1_id" from table: "chats"
DROP INDEX "public"."idx_chats_user1_id";
-- Drop index "idx_chats_user2_id" from table: "chats"
DROP INDEX "public"."idx_chats_user2_id";
-- Modify "chats" table
ALTER TABLE "public"."chats" DROP COLUMN "deleted_at";
-- Create index "idx_chat_users" to table: "chats"
CREATE UNIQUE INDEX "idx_chat_users" ON "public"."chats" ("user1_id", "user2_id");
-- Modify "event_outcomes" table
ALTER TABLE "public"."event_outcomes" ALTER COLUMN "odds" TYPE bigint;
-- Rename a column from "updated_at" to "read_at"
ALTER TABLE "public"."messages" RENAME COLUMN "updated_at" TO "read_at";
-- Modify "messages" table
ALTER TABLE "public"."messages" DROP COLUMN "status", DROP COLUMN "deleted_at";
-- Modify "refresh_tokens" table
ALTER TABLE "public"."refresh_tokens" DROP COLUMN "deleted_at";
-- Modify "ticket_messages" table
ALTER TABLE "public"."ticket_messages" DROP COLUMN "deleted_at";
-- Drop index "idx_tournament_participants_tournament_id" from table: "tournament_participants"
DROP INDEX "public"."idx_tournament_participants_tournament_id";
-- Drop index "idx_tournament_participants_user_id" from table: "tournament_participants"
DROP INDEX "public"."idx_tournament_participants_user_id";
-- Modify "tournament_participants" table
ALTER TABLE "public"."tournament_participants" ALTER COLUMN "score" TYPE bigint, ALTER COLUMN "prize_won" TYPE bigint, DROP COLUMN "deleted_at";
-- Create index "idx_tournament_user" to table: "tournament_participants"
CREATE UNIQUE INDEX "idx_tournament_user" ON "public"."tournament_participants" ("tournament_id", "user_id");
-- Drop index "idx_users_name" from table: "users"
DROP INDEX "public"."idx_users_name";
-- Rename a column from "name" to "username"
ALTER TABLE "public"."users" RENAME COLUMN "name" TO "username";
-- Create index "idx_users_username" to table: "users"
CREATE UNIQUE INDEX "idx_users_username" ON "public"."users" ("username");
