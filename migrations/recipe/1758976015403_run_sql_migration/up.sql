-- The "Find or Create" function for ingredients.
CREATE OR REPLACE FUNCTION find_or_create_ingredient(ingredient_name TEXT)
RETURNS INT AS $$
DECLARE
    ingredient_id INT;
BEGIN
    WITH ins AS (
        INSERT INTO ingredients (name)
        VALUES (ingredient_name)
        ON CONFLICT (name) DO NOTHING
        RETURNING id
    )
    SELECT id INTO ingredient_id FROM ins;

    IF ingredient_id IS NULL THEN
        SELECT id INTO ingredient_id FROM ingredients WHERE name = ingredient_name;
    END IF;

    RETURN ingredient_id;
END;
$$ LANGUAGE plpgsql;
