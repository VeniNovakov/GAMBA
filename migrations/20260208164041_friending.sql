-- Create "friends" table
CREATE TABLE "public"."friends" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "friend_id" uuid NOT NULL,
  "status" character varying(20) NULL DEFAULT 'pending',
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_friends_friend" FOREIGN KEY ("friend_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_friends_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_friends_friend_id" to table: "friends"
CREATE INDEX "idx_friends_friend_id" ON "public"."friends" ("friend_id");
-- Create index "idx_friends_user_id" to table: "friends"
CREATE INDEX "idx_friends_user_id" ON "public"."friends" ("user_id");
