-- Create the table to track individual recipe purchases with Chapa
CREATE TABLE "user_purchased_recipes" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "user_id" UUID NOT NULL,
    "recipe_id" UUID NOT NULL,
    "chapa_transaction_ref" TEXT NOT NULL UNIQUE, -- The unique Tx_Ref from Chapa
    "amount_paid" NUMERIC(10, 2) NOT NULL,
    "currency" VARCHAR(3) NOT NULL,
    "purchased_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    PRIMARY KEY ("id"),
    FOREIGN KEY ("user_id") REFERENCES "public"."users"("id") ON DELETE cascade,
    FOREIGN KEY ("recipe_id") REFERENCES "public"."recipes"("id") ON DELETE cascade,
    
    -- A user can only buy a specific recipe once
    UNIQUE ("user_id", "recipe_id")
);
