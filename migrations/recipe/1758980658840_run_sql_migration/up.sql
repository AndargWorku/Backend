-- This function takes an array of ingredient names (e.g., ['Flour', 'Sugar']).
-- It returns a SETOF recipe IDs that match the criteria.
CREATE OR REPLACE FUNCTION filter_recipes_by_ingredients(ingredient_names TEXT[])
RETURNS SETOF UUID AS $$
BEGIN
    RETURN QUERY
    -- Select the recipe_id from the ingredients table
    SELECT recipe_id
    FROM ingredients
    -- Filter to only include rows where the ingredient name is in our input array
    WHERE name = ANY(ingredient_names)
    -- Group the results by recipe_id
    GROUP BY recipe_id
    -- This is the crucial part:
    -- HAVING clause ensures that the count of matching ingredients for a recipe
    -- is equal to the total number of ingredients we are searching for.
    -- This means the recipe must contain ALL the specified ingredients.
    HAVING COUNT(DISTINCT name) = array_length(ingredient_names, 1);
END;
$$ LANGUAGE plpgsql STABLE;
