ALTER TABLE public.users
ADD COLUMN IF NOT EXISTS subscription_status TEXT NOT NULL DEFAULT 'free',
ADD COLUMN IF NOT EXISTS subscription_expires_at TIMESTAMPTZ;
