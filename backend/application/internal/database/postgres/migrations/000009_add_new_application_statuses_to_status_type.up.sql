ALTER TYPE application_status ADD VALUE IF NOT EXISTS 'recalled';
ALTER TYPE application_status ADD VALUE IF NOT EXISTS 'pending_verification';
ALTER TYPE application_status ADD VALUE IF NOT EXISTS 'on_verification';
ALTER TYPE application_status ADD VALUE IF NOT EXISTS 'on_revision';

UPDATE applications SET status = 'pending_verification' WHERE status = 'awaiting_approval';
UPDATE applications SET status = 'failed' WHERE status = 'cancelled';