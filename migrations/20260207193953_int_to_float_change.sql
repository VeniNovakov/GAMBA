-- Modify "bets" table
ALTER TABLE "public"."bets" ALTER COLUMN "amount" TYPE numeric, ALTER COLUMN "odds" TYPE numeric, ALTER COLUMN "payout" TYPE numeric;
-- Modify "games" table
ALTER TABLE "public"."games" ALTER COLUMN "min_bet" TYPE numeric, ALTER COLUMN "max_bet" TYPE numeric, ALTER COLUMN "house_edge" TYPE numeric;
-- Modify "transactions" table
ALTER TABLE "public"."transactions" ALTER COLUMN "amount" TYPE numeric;
-- Modify "users" table
ALTER TABLE "public"."users" ALTER COLUMN "balance" TYPE numeric;
