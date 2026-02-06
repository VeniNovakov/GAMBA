-- Create "games" table
CREATE TABLE "public"."games" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  "description" text NULL,
  "category" character varying(30) NOT NULL,
  "status" character varying(20) NULL DEFAULT 'active',
  "min_bet" numeric NOT NULL,
  "max_bet" numeric NOT NULL,
  "house_edge" numeric NULL DEFAULT 0,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_games_deleted_at" to table: "games"
CREATE INDEX "idx_games_deleted_at" ON "public"."games" ("deleted_at");
-- Create "chats" table
CREATE TABLE "public"."chats" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user1_id" uuid NOT NULL,
  "user2_id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_chats_user1" FOREIGN KEY ("user1_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_chats_user2" FOREIGN KEY ("user2_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_chats_deleted_at" to table: "chats"
CREATE INDEX "idx_chats_deleted_at" ON "public"."chats" ("deleted_at");
-- Create index "idx_chats_user1_id" to table: "chats"
CREATE INDEX "idx_chats_user1_id" ON "public"."chats" ("user1_id");
-- Create index "idx_chats_user2_id" to table: "chats"
CREATE INDEX "idx_chats_user2_id" ON "public"."chats" ("user2_id");
-- Create "messages" table
CREATE TABLE "public"."messages" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "chat_id" uuid NOT NULL,
  "sender_id" uuid NOT NULL,
  "content" text NOT NULL,
  "status" character varying(20) NULL DEFAULT 'sent',
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_chats_messages" FOREIGN KEY ("chat_id") REFERENCES "public"."chats" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_messages_sender" FOREIGN KEY ("sender_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_messages_chat_id" to table: "messages"
CREATE INDEX "idx_messages_chat_id" ON "public"."messages" ("chat_id");
-- Create index "idx_messages_deleted_at" to table: "messages"
CREATE INDEX "idx_messages_deleted_at" ON "public"."messages" ("deleted_at");
-- Create index "idx_messages_sender_id" to table: "messages"
CREATE INDEX "idx_messages_sender_id" ON "public"."messages" ("sender_id");
