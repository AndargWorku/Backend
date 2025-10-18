CREATE TABLE "instructions" (
    "id" SERIAL PRIMARY KEY,
    "recipe_id" UUID NOT NULL REFERENCES "recipes"("id") ON DELETE CASCADE,
    "step_number" INTEGER NOT NULL,
    "description" TEXT NOT NULL
);
