ALTER TABLE recipes
ADD CONSTRAINT recipes_category_id_fkey
FOREIGN KEY (category_id) REFERENCES categories(id)
ON DELETE CASCADE;
