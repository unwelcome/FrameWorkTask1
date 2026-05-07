CREATE TABLE applications (
    uuid         UUID               PRIMARY KEY,
    version      INTEGER            NOT NULL DEFAULT 1,
    company_uuid UUID               NOT NULL,
    title        VARCHAR(255)       NOT NULL,
    description  TEXT,
    status       application_status NOT NULL DEFAULT 'created',
    managed_by   UUID,
    executed_by  UUID,
    created_by   UUID               NOT NULL,
    created_at   TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    closed_at    TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_applications_company_status ON applications(company_uuid, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_applications_created_by     ON applications(created_by)           WHERE deleted_at IS NULL;
CREATE INDEX idx_applications_managed_by     ON applications(managed_by)           WHERE deleted_at IS NULL;
CREATE INDEX idx_applications_executed_by    ON applications(executed_by)          WHERE deleted_at IS NULL;
