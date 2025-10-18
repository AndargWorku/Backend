CREATE TABLE "ingredients" (
    "id" SERIAL PRIMARY KEY,
    "recipe_id" UUID NOT NULL REFERENCES "recipes"("id") ON DELETE CASCADE,
    "name" TEXT NOT NULL,        
    "quantity" TEXT NOT NULL,      
    "unit" TEXT                   
);
