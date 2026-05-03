package postgresDB

func migrationQueries() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS users (
			uuid VARCHAR(36) UNIQUE NOT NULL,
			email varchar(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name varchar(50) NOT NULL,
			last_name varchar(50) NOT NULL,
			patronymic varchar(50),
			created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
	}
}
