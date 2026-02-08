-- Modify "tournament_participants" table
ALTER TABLE "public"."tournament_participants" ALTER COLUMN "score" TYPE numeric, ALTER COLUMN "prize_won" TYPE numeric;
-- Modify "tournaments" table
ALTER TABLE "public"."tournaments" ALTER COLUMN "entry_fee" TYPE numeric, ALTER COLUMN "prize_pool" TYPE numeric;
