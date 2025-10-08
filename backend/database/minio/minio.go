package miniodb

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog"
)

type ConnectConfig struct {
	BackendEndpoint string
	PublicEndpoint  string
	Username        string
	Password        string
	UseSSL          bool
}

type Bucket struct {
	Name   string
	IsOpen bool
}

func InitMinio(cfg *ConnectConfig, l zerolog.Logger, buckets []*Bucket) *minio.Client {
	minioClient, err := ConnectMinio(cfg)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to connect to minio")
	}

	// Инициализируем бакеты
	for _, bucket := range buckets {
		err = CreateBucketIfNotExists(context.Background(), minioClient, bucket.Name)
		if err != nil {
			l.Fatal().Str("bucketName", bucket.Name).Err(err).Msg("Failed to create bucket")
		}

		// Выставляем политику для бакета
		if bucket.IsOpen {
			err = SetOpenBucketPolicy(context.Background(), minioClient, bucket.Name)
			if err != nil {
				l.Fatal().Str("bucketName", bucket.Name).Err(err).Msg("Failed to set bucket policy")
			}
		}
	}

	l.Trace().Msg("Successfully connected to MinIO")
	return minioClient
}

func ConnectMinio(cfg *ConnectConfig) (*minio.Client, error) {
	// Подключение к minio
	minioClient, err := minio.New(cfg.BackendEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Username, cfg.Password, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	// Пинг
	_, err = minioClient.ListBuckets(context.Background())
	if err != nil {
		return nil, fmt.Errorf("MinIO ping failed: %w", err)
	}

	return minioClient, nil
}

func CreateBucketIfNotExists(ctx context.Context, minioClient *minio.Client, bucketName string) error {
	// Создаем bucket
	err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		// Проверяем существует ли bucket
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			return nil
		} else {
			return fmt.Errorf("create bucket error: %s", err)
		}
	}

	return nil
}

func SetOpenBucketPolicy(ctx context.Context, minioClient *minio.Client, bucketName string) error {
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}
		]
	}`, bucketName)
	return minioClient.SetBucketPolicy(ctx, bucketName, policy)
}
