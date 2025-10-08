package repositories

import (
	"context"
	"database/sql"
	"github.com/unwelcome/iqjtest/internal/entities"
)

type CatRepository interface {
	CreateCat(ctx context.Context, userID int, cat *entities.Cat) error
	GetCatByID(ctx context.Context, catID int) (*entities.Cat, error)
	GetAllCats(ctx context.Context) ([]*entities.CatWithPrimePhoto, error)
	UpdateCatName(ctx context.Context, catID int, newName string) error
	UpdateCatAge(ctx context.Context, catID int, newAge int) error
	UpdateCatDescription(ctx context.Context, catID int, newDescription string) error
	UpdateCat(ctx context.Context, catID int, catUpdateRequest *entities.CatUpdateRequest) error
	DeleteCat(ctx context.Context, catID int) error
}

type catRepositoryImpl struct {
	db *sql.DB
}

func NewCatRepository(db *sql.DB) CatRepository {
	return &catRepositoryImpl{db: db}
}

func (r *catRepositoryImpl) CreateCat(ctx context.Context, userID int, cat *entities.Cat) error {
	query := `INSERT INTO cats(name, age, description, created_by) VALUES ($1, $2, $3, $4) RETURNING id;`

	err := r.db.QueryRowContext(ctx, query, cat.Name, cat.Age, cat.Description, userID).Scan(&cat.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *catRepositoryImpl) GetCatByID(ctx context.Context, catID int) (*entities.Cat, error) {
	// Запрос на получение кота
	query := `SELECT name, age, description, created_at, created_by FROM cats WHERE id = $1;`

	cat := &entities.Cat{ID: catID}

	// Выполняем запрос
	err := r.db.QueryRowContext(ctx, query, catID).Scan(&cat.Name, &cat.Age, &cat.Description, &cat.CreatedAt, &cat.CreatedBy)
	if err != nil {
		return nil, err
	}

	return cat, nil
}

func (r *catRepositoryImpl) GetAllCats(ctx context.Context) ([]*entities.CatWithPrimePhoto, error) {
	// Запрос на получение всех котов с left join фото котов, сортируя по catID, затем по is_primary и в конце по photoID
	// Т.о. Получаем кота с первым is_primary фото либо кота с первым фото либо кота без фото
	query := `
		SELECT DISTINCT ON (c.id)
			c.id,
			c.name,
			c.age,
			cp.id AS photo_id, 
			cp.url
		FROM cats c
		LEFT JOIN cat_photos cp ON c.id = cp.cat_id
		ORDER BY c.id, cp.is_primary DESC NULLS LAST, cp.id ASC;
	`

	// Выполняем запрос
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		cats    []*entities.CatWithPrimePhoto
		photoID sql.NullInt64
		url     sql.NullString
	)

	// Мэппинг ответа в структуру
	for rows.Next() {
		cat := &entities.CatWithPrimePhoto{}

		err = rows.Scan(&cat.ID, &cat.Name, &cat.Age, &photoID, &url)
		if err != nil {
			return nil, err
		}

		// Если photoID не null
		if photoID.Valid {
			id := int(photoID.Int64)
			cat.PhotoID = &id
		}
		// Если url не null
		if url.Valid {
			urlStr := url.String
			cat.Url = &urlStr
		}

		cats = append(cats, cat)
	}

	return cats, nil
}

func (r *catRepositoryImpl) UpdateCatName(ctx context.Context, catID int, newName string) error {
	query := `UPDATE cats SET name = $1 WHERE id = $2;`

	_, err := r.db.ExecContext(ctx, query, newName, catID)
	if err != nil {
		return err
	}

	return nil
}

func (r *catRepositoryImpl) UpdateCatAge(ctx context.Context, catID int, newAge int) error {
	query := `UPDATE cats SET age = $1 WHERE id = $2;`

	_, err := r.db.ExecContext(ctx, query, newAge, catID)
	if err != nil {
		return err
	}

	return nil
}

func (r *catRepositoryImpl) UpdateCatDescription(ctx context.Context, catID int, newDescription string) error {
	query := `UPDATE cats SET description = $1 WHERE id = $2;`

	_, err := r.db.ExecContext(ctx, query, newDescription, catID)
	if err != nil {
		return err
	}

	return nil
}

func (r *catRepositoryImpl) UpdateCat(ctx context.Context, catID int, catUpdateRequest *entities.CatUpdateRequest) error {
	query := `UPDATE cats SET (name, age, description) = ($1, $2, $3) WHERE id = $4;`

	_, err := r.db.ExecContext(ctx, query, catUpdateRequest.Name, catUpdateRequest.Age, catUpdateRequest.Description, catID)
	if err != nil {
		return err
	}

	return nil
}

func (r *catRepositoryImpl) DeleteCat(ctx context.Context, catID int) error {
	query := `DELETE FROM cats WHERE id = $1;`

	_, err := r.db.ExecContext(ctx, query, catID)
	if err != nil {
		return err
	}

	return nil
}
