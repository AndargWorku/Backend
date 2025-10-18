-- This table stores each user's subscription status.
CREATE TABLE "subscriptions" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "user_id" UUID NOT NULL UNIQUE, -- A user can only have one subscription record
    "payment_provider" TEXT NOT NULL, -- e.g., 'chapa' or 'stripe'
    "subscription_ref" TEXT UNIQUE NOT NULL, -- The unique ID from the payment provider
    "status" TEXT NOT NULL, -- e.g., 'active', 'canceled', 'expired'
    "current_period_end" TIMESTAMPTZ NOT NULL, -- The date when the subscription expires
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY ("id"),
    FOREIGN KEY ("user_id") REFERENCES "public"."users"("id") ON DELETE cascade
);
CREATE INDEX ON "public"."subscriptions" ("user_id");
