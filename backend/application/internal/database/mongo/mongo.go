package mongoDB

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"

	sharedConfig "github.com/unwelcome/FrameWorkTask1/backend/shared/config"
	sharedMongo "github.com/unwelcome/FrameWorkTask1/backend/shared/mongo"
)

type DatabaseRepository struct {
	ApplicationVersion ApplicationVersionRepository
	client             *mongo.Client
}

func (r *DatabaseRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx, nil)
}

func NewDatabaseInstance(cfg sharedConfig.MongoDBConfig) *DatabaseRepository {
	sharedMongo.Setup(cfg.RootConnectionString(), cfg.DbName, cfg.User, cfg.Password)

	client := sharedMongo.Connect(cfg.ConnectionString())
	db := client.Database(cfg.DbName)

	return &DatabaseRepository{
		ApplicationVersion: newApplicationVersionRepository(db),
		client:             client,
	}
}
