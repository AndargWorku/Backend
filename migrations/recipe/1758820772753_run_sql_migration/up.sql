CREATE OR REPLACE FUNCTION set_current_timestamp_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_recipes_updated_at
BEFORE UPDATE ON "recipes"
FOR EACH ROW
EXECUTE FUNCTION set_current_timestamp_updated_at();
