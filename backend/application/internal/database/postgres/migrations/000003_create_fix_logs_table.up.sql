CREATE TABLE application_fix_logs (
    uuid             UUID         PRIMARY KEY,
    application_uuid UUID         NOT NULL REFERENCES applications(uuid) ON DELETE CASCADE,
    text             TEXT         NOT NULL,
    created_by       UUID         NOT NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
