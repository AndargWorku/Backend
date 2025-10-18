CREATE TABLE "recipes" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id" UUID NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    "title" TEXT NOT NULL,
    "description" TEXT NOT NULL,
    "prep_time_minutes" INTEGER,
    "cook_time_minutes" INTEGER,
    "servings" INTEGER,
    "featured_image_url" TEXT, 
    "is_premium" BOOLEAN NOT NULL DEFAULT false,
    "price" NUMERIC(10, 2),
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);
