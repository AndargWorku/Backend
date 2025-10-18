CREATE OR REPLACE FUNCTION recipe_total_comments(recipe_row recipes)
RETURNS BIGINT AS $$
  SELECT COUNT(*) FROM "comments" WHERE recipe_id = recipe_row.id;
$$ LANGUAGE sql STABLE;
