ALTER TABLE applications DROP COLUMN IF EXISTS revision_count;
ALTER TABLE applications DROP COLUMN IF EXISTS updated_at;
ALTER TABLE applications DROP COLUMN IF EXISTS updated_by;
ALTER TABLE applications DROP COLUMN IF EXISTS inspected_by;
