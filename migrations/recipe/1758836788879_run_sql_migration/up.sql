CREATE OR REPLACE FUNCTION find_or_create_ingredient(ingredient_name TEXT)
RETURNS INT AS $$
DECLARE
    ingredient_id INT;
BEGIN
    -- Attempt to insert the new ingredient.
    -- If an ingredient with the same name exists, DO NOTHING.
    WITH ins AS (
        INSERT INTO ingredients (name)
        VALUES (ingredient_name)
        ON CONFLICT (name) DO NOTHING
        RETURNING id
    )
    -- Get the ID from the row that was just inserted (if any).
    SELECT id INTO ingredient_id FROM ins;

    -- If ingredient_id is NULL, it means it already existed. Select its ID.
    IF ingredient_id IS NULL THEN
        SELECT id INTO ingredient_id FROM ingredients WHERE name = ingredient_name;
    END IF;

    -- Return the final ID.
    RETURN ingredient_id;
END;
$$ LANGUAGE plpgsql;
