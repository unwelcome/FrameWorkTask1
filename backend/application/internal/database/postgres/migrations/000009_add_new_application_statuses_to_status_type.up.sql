ALTER TYPE application_status ADD VALUE IF NOT EXISTS 'recalled';
ALTER TYPE application_status ADD VALUE IF NOT EXISTS 'pending_verification';
ALTER TYPE application_status ADD VALUE IF NOT EXISTS 'on_verification';
ALTER TYPE application_status ADD VALUE IF NOT EXISTS 'on_revision';