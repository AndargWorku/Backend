CREATE TABLE "recipe_ingredients" (
    "id" SERIAL PRIMARY KEY,
    "recipe_id" UUID NOT NULL REFERENCES "recipes"("id") ON DELETE CASCADE,
    "ingredient_id" INTEGER NOT NULL REFERENCES "ingredients"("id") ON DELETE RESTRICT,
    "quantity" TEXT NOT NULL,
    "unit" TEXT
);
