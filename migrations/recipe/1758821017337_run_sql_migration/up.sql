CREATE INDEX ON "recipes" ("user_id");
CREATE INDEX ON "recipe_images" ("recipe_id");
CREATE INDEX ON "recipe_ingredients" ("recipe_id");
CREATE INDEX ON "recipe_ingredients" ("ingredient_id");
CREATE INDEX ON "instructions" ("recipe_id");
CREATE INDEX ON "comments" ("recipe_id");
CREATE INDEX ON "comments" ("user_id");
