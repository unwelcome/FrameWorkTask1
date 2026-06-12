ALTER TABLE users
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ALTER COLUMN email         DROP NOT NULL,
    ALTER COLUMN password_hash DROP NOT NULL,
    ALTER COLUMN first_name    DROP NOT NULL,
    ALTER COLUMN last_name     DROP NOT NULL;
