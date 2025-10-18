-- This function is the only one you need for complex filtering.
-- It takes an array of ingredient names (e.g., ['Flour', 'Sugar']).
-- It returns a set of 'recipes' that contain ALL of those ingredients.

CREATE OR REPLACE FUNCTION filter_recipes_by_ingredients(ingredient_names TEXT[])
RETURNS SETOF recipes AS $$
BEGIN
    RETURN QUERY
    -- First, select the full recipe rows from the 'recipes' table.
    SELECT *
    FROM recipes
    -- We only want recipes where the ID is in our list of matches.
    WHERE id IN (
        -- This subquery is the "brain". It finds the recipe_ids that match.
        SELECT
            ri.recipe_id
        FROM
            recipe_ingredients ri
        -- We join to the master 'ingredients' table to check the name.
        JOIN
            ingredients i ON ri.ingredient_id = i.id
        -- We only care about ingredients from our input list.
        WHERE
            i.name = ANY(ingredient_names)
        -- We group by recipe to count the matches for each one.
        GROUP BY
            ri.recipe_id
        -- This is the crucial filter: the count of distinct matching ingredients
        -- for a recipe MUST be equal to the number of ingredients we searched for.
        -- This ensures it has ALL of them.
        HAVING
            COUNT(DISTINCT i.id) = array_length(ingredient_names, 1)
    );
END;
$$ LANGUAGE plpgsql STABLE;
