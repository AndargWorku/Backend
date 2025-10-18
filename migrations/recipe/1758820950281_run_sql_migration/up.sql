CREATE OR REPLACE FUNCTION recipe_total_likes(recipe_row recipes)
RETURNS BIGINT AS $$
  SELECT COUNT(*) FROM "recipe_likes" WHERE recipe_id = recipe_row.id;
$$ LANGUAGE sql STABLE;
