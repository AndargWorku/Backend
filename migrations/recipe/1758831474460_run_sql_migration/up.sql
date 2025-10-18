CREATE TABLE "categories" (
    "id" SERIAL PRIMARY KEY,
    "recipe_id" UUID NOT NULL REFERENCES "recipes"("id") ON DELETE CASCADE,
    "name" TEXT NOT NULL         
);
