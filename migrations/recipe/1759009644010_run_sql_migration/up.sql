-- ========= Corrected Function for Categories =========
-- Notice the return type is now "RETURNS SETOF categories"
CREATE OR REPLACE FUNCTION find_or_create_category(category_name TEXT)
RETURNS SETOF categories AS $$
DECLARE
    category_id INT;
    new_slug TEXT;
BEGIN
    new_slug := lower(regexp_replace(category_name, '[^a-zA-Z0-9\s]+', '', 'g'));
    new_slug := regexp_replace(trim(new_slug), '\s+', '-', 'g');

    -- The logic inside is the same, but we will use the final ID
    -- to return the full category row at the end.
    WITH ins AS (
        INSERT INTO categories (name, slug)
        VALUES (category_name, new_slug)
        ON CONFLICT (name) DO NOTHING
        RETURNING id
    )
    SELECT id INTO category_id FROM ins;

    IF category_id IS NULL THEN
        SELECT id INTO category_id FROM categories WHERE categories.name = category_name;
    END IF;

    -- This is the new part: return the full row from the 'categories' table
    -- that matches the final ID.
    RETURN QUERY SELECT * FROM categories WHERE categories.id = category_id;
END;
$$ LANGUAGE plpgsql;


-- ========= Corrected Function for Ingredients =========
-- Notice the return type is now "RETURNS SETOF ingredients"
CREATE OR REPLACE FUNCTION find_or_create_ingredient(ingredient_name TEXT)
RETURNS SETOF ingredients AS $$
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
        SELECT id INTO ingredient_id FROM ingredients WHERE ingredients.name = ingredient_name;
    END IF;

    -- Return the full row from the 'ingredients' table.
    RETURN QUERY SELECT * FROM ingredients WHERE ingredients.id = ingredient_id;
END;
$$ LANGUAGE plpgsql;
