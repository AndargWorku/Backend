CREATE TABLE "comments" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "recipe_id" UUID NOT NULL REFERENCES "recipes"("id") ON DELETE CASCADE,
    "user_id" UUID NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    "content" TEXT NOT NULL,
    "parent_comment_id" UUID REFERENCES "comments"("id") ON DELETE CASCADE,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);
