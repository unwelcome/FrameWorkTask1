CREATE TABLE departments (
    uuid         UUID PRIMARY KEY,
    company_uuid UUID         NOT NULL REFERENCES companies(uuid) ON DELETE CASCADE,
    title        VARCHAR(255) NOT NULL,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    created_by   UUID         NOT NULL
);