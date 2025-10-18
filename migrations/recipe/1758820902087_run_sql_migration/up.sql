CREATE OR REPLACE FUNCTION recipe_average_rating(recipe_row recipes)
RETURNS NUMERIC AS $$
  SELECT AVG(rating) FROM "recipe_ratings" WHERE recipe_id = recipe_row.id;
$$ LANGUAGE sql STABLE;
