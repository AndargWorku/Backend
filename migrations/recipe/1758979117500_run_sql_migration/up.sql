CREATE TABLE "ingredients" (
    "id" SERIAL PRIMARY KEY,
    "recipe_id" UUID NOT NULL REFERENCES "recipes"("id") ON DELETE CASCADE,
    "name" TEXT NOT NULL,          -- The name of the ingredient, e.g., "All-Purpose Flour"
    "quantity" TEXT NOT NULL,      -- The amount, e.g., "2" or "1/2"
    "unit" TEXT                   -- The unit, e.g., "cups", "grams". Can be NULL.
);
