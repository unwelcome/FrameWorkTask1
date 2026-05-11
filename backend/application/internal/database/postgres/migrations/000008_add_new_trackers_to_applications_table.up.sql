ALTER TABLE applications ADD COLUMN revision_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE applications ADD COLUMN updated_at TIMESTAMPTZ;
ALTER TABLE applications ADD COLUMN updated_by UUID;
ALTER TABLE applications ADD COLUMN inspected_by UUID;

CREATE INDEX idx_applications_inspected_by ON applications(inspected_by) WHERE deleted_at IS NULL;