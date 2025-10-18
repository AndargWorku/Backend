-- This function takes a username (TEXT) as input.
-- It also returns SETOF recipes.
CREATE OR REPLACE FUNCTION get_recipes_by_username(creator_username TEXT)
RETURNS SETOF recipes AS $$
BEGIN
    RETURN QUERY
    -- We select all columns from the recipes table (aliased as 'r').
    SELECT r.*
    FROM recipes AS r
    -- We must JOIN the recipes table with the users table to find the user by their name.
    JOIN users AS u ON r.user_id = u.id
    -- The WHERE clause now filters on the username.
    WHERE u.username = creator_username
    ORDER BY r.created_at DESC;
END;
$$ LANGUAGE plpgsql STABLE;
