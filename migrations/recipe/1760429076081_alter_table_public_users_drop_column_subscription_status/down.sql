alter table "public"."users" alter column "subscription_status" set default ''free'::text';
alter table "public"."users" alter column "subscription_status" drop not null;
alter table "public"."users" add column "subscription_status" text;
