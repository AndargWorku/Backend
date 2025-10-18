CREATE TABLE "user_bookmarks" (
    "recipe_id" UUID NOT NULL REFERENCES "recipes"("id") ON DELETE CASCADE,
    "user_id" UUID NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    PRIMARY KEY ("recipe_id", "user_id")
);
