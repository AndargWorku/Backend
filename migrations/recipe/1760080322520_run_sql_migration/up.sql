-- Create indexes for faster lookups.
CREATE INDEX IF NOT EXISTS idx_payments_user_id ON public.payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_tx_ref ON public.payments(tx_ref);
