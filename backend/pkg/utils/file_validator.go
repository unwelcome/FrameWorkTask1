package utils

import (
	"crypto/rand"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"mime/multipart"
	"path/filepath"
)

func GetFilesFromFormData(c *fiber.Ctx, key string, maxFilesCount int) ([]*multipart.FileHeader, error) {

	// Парсим multipart/formData
	form, err := c.MultipartForm()
	if err != nil {
		return nil, fmt.Errorf("failed to parse multipart form")
	}

	// Получаем файлы
	files := form.File[key]
	if len(files) == 0 {
		return nil, fmt.Errorf("no file found in form")
	}

	if len(files) > maxFilesCount {
		return nil, fmt.Errorf("too many files")
	}

	return files, nil
}

func CheckFileSize(file *multipart.FileHeader, maxSize int) error {
	// Проверяем содержимое файла
	if file.Size == 0 {
		return fmt.Errorf("file is empty")
	}

	// Проверяем размер файла
	if file.Size > int64(maxSize) {
		return fmt.Errorf("file is too large")
	}

	return nil
}

func IsImageFile(fileHeader *multipart.FileHeader) bool {
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}

	contentType := fileHeader.Header.Get("Content-Type")
	return allowedTypes[contentType]
}

func GenerateFilename(fileName string, id int, prefix string) string {
	// Извлекаем расширение файла
	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = ".jpg"
	}

	// Генерируем уникальный ID для имени файла
	uniqueID := generateUniqueID()

	// Формируем название файла cat/{catID}/{uuid}{ext}
	return fmt.Sprintf("%s/%d/%s%s", prefix, id, uniqueID, ext)
}

func generateUniqueID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}
