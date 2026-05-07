CREATE TABLE companies (
    uuid       UUID           PRIMARY KEY,
    title      VARCHAR(255)   NOT NULL,
    status     company_status NOT NULL DEFAULT 'close',
    created_by UUID           NOT NULL,
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);
