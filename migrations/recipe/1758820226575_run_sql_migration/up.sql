CREATE TABLE "recipe_images" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "recipe_id" UUID NOT NULL REFERENCES "recipes"("id") ON DELETE CASCADE,
    "image_url" TEXT NOT NULL,
    "is_featured" BOOLEAN NOT NULL DEFAULT false
);
