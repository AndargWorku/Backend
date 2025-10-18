CREATE OR REPLACE FUNCTION find_or_create_category(category_name TEXT)
RETURNS SETOF categories AS $$
DECLARE
    category_id INT;
    new_slug TEXT;
BEGIN
    -- Sanitize the name to create a URL-friendly slug
    new_slug := lower(regexp_replace(category_name, '[^a-zA-Z0-9\s]+', '', 'g'));
    new_slug := regexp_replace(trim(new_slug), '\s+', '-', 'g');

    -- Attempt to insert the new category. If it already exists, do nothing.
    WITH ins AS (
        INSERT INTO categories (name, slug)
        VALUES (category_name, new_slug)
        ON CONFLICT (name) DO NOTHING
        RETURNING id
    )
    SELECT id INTO category_id FROM ins;

    -- If the insert did nothing, it means the category already existed. Find its ID.
    IF category_id IS NULL THEN
        SELECT id INTO category_id FROM categories WHERE categories.name = category_name;
    END IF;

    -- Return the full row for the final category, which Hasura can track.
    RETURN QUERY SELECT * FROM categories WHERE categories.id = category_id;
END;
$$ LANGUAGE plpgsql;
