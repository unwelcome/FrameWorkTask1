CREATE TABLE employees (
    company_uuid UUID          NOT NULL REFERENCES companies(uuid) ON DELETE CASCADE,
    user_uuid    UUID          NOT NULL,
    role         employee_role NOT NULL DEFAULT 'unemployed',
    joined_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_uuid, user_uuid)
);
