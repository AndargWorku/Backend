CREATE TABLE "recipe_ratings" (
    "recipe_id" UUID NOT NULL REFERENCES "recipes"("id") ON DELETE CASCADE,
    "user_id" UUID NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    "rating" INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    PRIMARY KEY ("recipe_id", "user_id")
);
