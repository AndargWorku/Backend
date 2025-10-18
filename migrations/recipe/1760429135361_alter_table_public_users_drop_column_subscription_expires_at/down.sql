alter table "public"."users" alter column "subscription_expires_at" drop not null;
alter table "public"."users" add column "subscription_expires_at" timestamptz;
