-- This function queries the categories table, finds all the unique names,
-- and returns them as a table with a single "name" column.
CREATE OR REPLACE FUNCTION get_unique_categories()
RETURNS SETOF categories AS $$
BEGIN
    -- The DISTINCT ON(name) is a PostgreSQL feature that gets the first row
    -- for each unique name. It's generally more efficient than a simple DISTINCT.
    -- We order by name to ensure we get a consistent row each time.
    RETURN QUERY
    SELECT DISTINCT ON (name) *
    FROM categories
    ORDER BY name ASC;
END;
$$ LANGUAGE plpgsql STABLE;
