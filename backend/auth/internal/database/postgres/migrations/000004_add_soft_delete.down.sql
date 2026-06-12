ALTER TABLE users
    DROP COLUMN deleted_at,
    ALTER COLUMN email         SET NOT NULL,
    ALTER COLUMN password_hash SET NOT NULL,
    ALTER COLUMN first_name    SET NOT NULL,
    ALTER COLUMN last_name     SET NOT NULL;
