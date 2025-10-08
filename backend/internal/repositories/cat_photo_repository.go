package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/unwelcome/iqjtest/internal/entities"
	"github.com/unwelcome/iqjtest/pkg/utils"
)

type CatPhotoRepository interface {
	AddCatPhoto(ctx context.Context, catID int, req *entities.CatPhotoUploadRequest) (*entities.CatPhotoUploadSuccess, error)
	GetAllCatPhotos(ctx context.Context, catID int) ([]*entities.CatPhotoUrl, error)
	GetCatPhotoByID(ctx context.Context, photoID int) (*entities.CatPhoto, error)
	SetCatPhotoPrimary(ctx context.Context, catID, photoID int) error
	DeleteCatPhoto(ctx context.Context, photoID int) error
	DeleteAllCatPhotos(ctx context.Context, catID int) error
}

type catPhotoRepositoryImpl struct {
	db          *sql.DB
	minioClient *minio.Client
	endpoint    string
	bucketName  string
}

func NewCatPhotoRepository(db *sql.DB, minioClient *minio.Client, endpoint, bucketName string) CatPhotoRepository {
	return &catPhotoRepositoryImpl{
		db:          db,
		minioClient: minioClient,
		endpoint:    endpoint,
		bucketName:  bucketName,
	}
}

func (r *catPhotoRepositoryImpl) AddCatPhoto(ctx context.Context, catID int, req *entities.CatPhotoUploadRequest) (*entities.CatPhotoUploadSuccess, error) {
	// Генерируем уникальное имя файла
	filename := utils.GenerateFilename(req.FileName, catID, "cat")

	// Сохраняем файл в Minio
	_, err := r.minioClient.PutObject(
		ctx,
		r.bucketName,
		filename,
		req.File,
		req.FileSize,
		minio.PutObjectOptions{
			ContentType: req.MimeType,
		})
	if err != nil {
		return nil, err
	}

	// Создаем тело ответа
	res := &entities.CatPhotoUploadSuccess{FileName: filename}

	// Создаем публичный url, формат: http://localhost:9000/bucket-name/filename
	res.Url = fmt.Sprintf("http://%s/%s/%s", r.endpoint, r.bucketName, filename)

	// Сохраняем фото в бд (is_primary = false)
	query := `INSERT INTO cat_photos (cat_id, url, filename, filesize, mime_type) VALUES ($1, $2, $3, $4, $5) RETURNING id;`
	err = r.db.QueryRowContext(ctx, query, catID, res.Url, filename, req.FileSize, req.MimeType).Scan(&res.ID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *catPhotoRepositoryImpl) GetAllCatPhotos(ctx context.Context, catID int) ([]*entities.CatPhotoUrl, error) {
	query := `SELECT id, url, is_primary FROM cat_photos WHERE cat_id = $1 ORDER BY is_primary DESC, id ASC;`

	// Выполняем запрос в бд
	rows, err := r.db.QueryContext(ctx, query, catID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var catPhotos []*entities.CatPhotoUrl

	// Меппинг ответа в структуру
	for rows.Next() {
		catPhoto := &entities.CatPhotoUrl{}
		err = rows.Scan(&catPhoto.ID, &catPhoto.Url, &catPhoto.IsPrimary)
		if err != nil {
			return nil, err
		}
		catPhotos = append(catPhotos, catPhoto)
	}

	return catPhotos, nil
}

func (r *catPhotoRepositoryImpl) GetCatPhotoByID(ctx context.Context, photoID int) (*entities.CatPhoto, error) {
	query := `SELECT url, cat_id, filename, filesize, mime_type, is_primary, created_at FROM cat_photos WHERE id = $1;`

	catPhoto := &entities.CatPhoto{ID: photoID}
	err := r.db.QueryRowContext(ctx, query, photoID).Scan(&catPhoto.Url, &catPhoto.CatID, &catPhoto.FileName, &catPhoto.FileSize, &catPhoto.MimeType, &catPhoto.IsPrimary, &catPhoto.CreatedAt)
	if err != nil {
		return nil, err
	}

	return catPhoto, nil
}

func (r *catPhotoRepositoryImpl) SetCatPhotoPrimary(ctx context.Context, catID, photoID int) error {
	// Создаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx error: %w", err)
	}
	defer tx.Rollback()

	// Проверяем, что фото принадлежит коту
	var exists bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM cat_photos WHERE id = $1 AND cat_id = $2)`, photoID, catID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check photo ownership error: %w", err)
	}
	if !exists {
		return fmt.Errorf("photo %d does not belong to cat %d", photoID, catID)
	}

	// Убираем у всех фото конкретного кота is_primary
	_, err = tx.ExecContext(ctx, `UPDATE cat_photos SET is_primary = false WHERE cat_id = $1;`, catID)
	if err != nil {
		return fmt.Errorf("remove all is_primary error: %w", err)
	}

	// Устанавливаем is_primary для выбранной фото
	_, err = tx.ExecContext(ctx, `UPDATE cat_photos SET is_primary = true WHERE id = $1;`, photoID)
	if err != nil {
		return fmt.Errorf("set is_primary by id error: %w", err)
	}

	// Коммитим транзакцию
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx error: %w", err)
	}

	return nil
}

func (r *catPhotoRepositoryImpl) DeleteCatPhoto(ctx context.Context, photoID int) error {
	// Удаляем файл из бд и получаем filename
	var filename string
	query := `DELETE FROM cat_photos WHERE id = $1 RETURNING filename;`
	err := r.db.QueryRowContext(ctx, query, photoID).Scan(&filename)
	if err != nil {
		return err
	}

	// Удаляем файл из minio
	err = r.minioClient.RemoveObject(ctx, r.bucketName, filename, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (r *catPhotoRepositoryImpl) DeleteAllCatPhotos(ctx context.Context, catID int) error {
	prefix := fmt.Sprintf("cat/%d/", catID)

	// Создаем канал для записи объектов
	objectCh := make(chan minio.ObjectInfo)

	// Создаем горутину
	go func() {
		defer close(objectCh)

		// Записываем в канал все фото кота
		for object := range r.minioClient.ListObjects(ctx, r.bucketName, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		}) {
			if object.Err != nil {
				continue
			}
			objectCh <- object
		}
	}()

	// Удаляем все фото кота
	errorCh := r.minioClient.RemoveObjects(ctx, r.bucketName, objectCh, minio.RemoveObjectsOptions{})
	for removeErr := range errorCh {
		return fmt.Errorf("failed to remove photo %s: %w", removeErr.ObjectName, removeErr.Err)
	}

	return nil
}
