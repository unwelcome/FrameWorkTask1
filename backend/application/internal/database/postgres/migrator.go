package postgresDB

func migrationQueries() []string {
	return []string{
		`DO $$ BEGIN
			CREATE TYPE STATUS AS ENUM (
				'created',
				'assigned',
				'in_progress',
				'on_hold',
				'awaiting_approval',
				'completed',
				'cancelled',
				'failed',
				'archived'
			);
			EXCEPTION
				WHEN duplicate_object THEN null;
		END $$;`,

		`CREATE TABLE IF NOT EXISTS applications (
			uuid VARCHAR(36) UNIQUE NOT NULL,
			version INTEGER DEFAULT 1,
			company_uuid VARCHAR(36) NOT NULL,

			title VARCHAR(255) NOT NULL,
			description TEXT,
			status STATUS NOT NULL DEFAULT 'created',

			managed_by VARCHAR(36) DEFAULT NULL,
			executed_by VARCHAR(36) DEFAULT NULL,

			created_by VARCHAR(36) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			closed_at TIMESTAMP DEFAULT NULL,
			deleted_at TIMESTAMP DEFAULT NULL);`,

		`CREATE INDEX IF NOT EXISTS idx_applications_company_status ON applications(company_uuid, status) WHERE deleted_at IS NULL;`,
		`CREATE INDEX IF NOT EXISTS idx_applications_created_by ON applications(created_by) WHERE deleted_at IS NULL;`,
		`CREATE INDEX IF NOT EXISTS idx_applications_managed_by ON applications(managed_by) WHERE deleted_at IS NULL;`,
		`CREATE INDEX IF NOT EXISTS idx_applications_executed_by ON applications(executed_by) WHERE deleted_at IS NULL;`,

		`CREATE TABLE IF NOT EXISTS application_fix_logs (
			id SERIAL PRIMARY KEY,
			application_uuid VARCHAR(36) NOT NULL,
			text TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_by VARCHAR(36) NOT NULL
		);`,

		`ALTER TABLE application_fix_logs
			ADD CONSTRAINT fk_application_uuid FOREIGN KEY (application_uuid) REFERENCES applications(uuid) ON DELETE CASCADE;`,
	}
}
