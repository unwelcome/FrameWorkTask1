package postgresDB

import "database/sql"

type UserRepository interface {
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

//func (r *userRepository) CreateUser()
