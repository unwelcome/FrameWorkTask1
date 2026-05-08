CREATE TABLE application_versions (
    uuid             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    application_uuid UUID        NOT NULL REFERENCES applications(uuid) ON DELETE CASCADE,
    version          INTEGER     NOT NULL,
    body             JSONB       NOT NULL,
    saved_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (application_uuid, version)
);

CREATE INDEX idx_application_versions_app_uuid ON application_versions(application_uuid);
