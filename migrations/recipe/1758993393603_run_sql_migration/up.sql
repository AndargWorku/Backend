-- ========= Function 4: Filter Recipes by a List of Ingredients =========
-- This is the most complex and powerful function. It fulfills your "filter recipes by ingredients"
-- requirement by finding recipes that contain ALL of the specified ingredients.

CREATE OR REPLACE FUNCTION filter_recipes_by_ingredients(ingredient_names TEXT[])
RETURNS SETOF recipes AS $$
BEGIN
    RETURN QUERY
    SELECT *
    FROM recipes
    WHERE id IN (
        SELECT recipe_id
        FROM ingredients
        WHERE name = ANY(ingredient_names)
        GROUP BY recipe_id
        HAVING COUNT(DISTINCT name) = array_length(ingredient_names, 1)
    );
END;
$$ LANGUAGE plpgsql STABLE;
