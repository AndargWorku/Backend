CREATE OR REPLACE FUNCTION find_or_create_category(category_name TEXT)
RETURNS INT AS $$
DECLARE
    category_id INT;
    new_slug TEXT;
BEGIN
    new_slug := lower(regexp_replace(category_name, '[^a-zA-Z0-9\s]+', '', 'g'));
    new_slug := regexp_replace(trim(new_slug), '\s+', '-', 'g');

    WITH ins AS (
        INSERT INTO categories (name, slug)
        VALUES (category_name, new_slug)
        ON CONFLICT (name) DO NOTHING
        RETURNING id
    )
    SELECT id INTO category_id FROM ins;

    IF category_id IS NULL THEN
        SELECT id INTO category_id FROM categories WHERE name = category_name;
    END IF;

    RETURN category_id;
END;
$$ LANGUAGE plpgsql;
