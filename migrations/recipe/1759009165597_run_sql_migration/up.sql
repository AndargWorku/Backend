-- ========= Function 1: The "Smart Librarian" for Categories =========
-- Takes a category name (TEXT), checks if it exists, creates it if it's new,
-- and always returns its final, correct ID (INTEGER).

CREATE OR REPLACE FUNCTION find_or_create_category(category_name TEXT)
RETURNS INT AS $$
DECLARE
    -- A variable to hold the final ID
    category_id INT;
    -- A variable to hold the URL-friendly version of the name
    new_slug TEXT;
BEGIN
    -- 1. Sanitize the user's input to create a clean slug
    -- (e.g., "Quick & Easy!" -> "quick-and-easy")
    new_slug := lower(regexp_replace(category_name, '[^a-zA-Z0-9\s]+', '', 'g'));
    new_slug := regexp_replace(trim(new_slug), '\s+', '-', 'g');

    -- 2. Attempt to insert the new category into the master 'categories' table.
    --    The "WITH" clause is a clean way to handle the insert.
    WITH ins AS (
        INSERT INTO categories (name, slug)
        VALUES (category_name, new_slug)
        -- This is the magic part: if a category with the same 'name' already
        -- exists (due to the UNIQUE constraint), DO NOTHING and don't raise an error.
        ON CONFLICT (name) DO NOTHING
        -- If the insert was successful, return the 'id' of the new row.
        RETURNING id
    )
    -- 3. Try to get the ID from the row that was just inserted (if any).
    SELECT id INTO category_id FROM ins;

    -- 4. If 'category_id' is still NULL, it means the INSERT did nothing because
    --    the category already existed. So, we now select its existing ID.
    IF category_id IS NULL THEN
        SELECT id INTO category_id FROM categories WHERE categories.name = category_name;
    END IF;

    -- 5. Return the final ID, whether it was newly created or already existed.
    RETURN category_id;
END;
$$ LANGUAGE plpgsql;


-- ========= Function 2: The "Smart Librarian" for Ingredients =========
-- Takes an ingredient name (TEXT), checks if it exists, creates it if it's new,
-- and always returns its final, correct ID (INTEGER).

CREATE OR REPLACE FUNCTION find_or_create_ingredient(ingredient_name TEXT)
RETURNS INT AS $$
DECLARE
    ingredient_id INT;
BEGIN
    -- This function follows the exact same logic as the category function,
    -- but is simpler because it doesn't need to create a slug.
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

    RETURN ingredient_id;
END;
$$ LANGUAGE plpgsql;
