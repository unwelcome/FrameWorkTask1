package mongoDB

import (
	sharedConfig "github.com/unwelcome/FrameWorkTask1/backend/shared/config"
	sharedMongo "github.com/unwelcome/FrameWorkTask1/backend/shared/mongo"
)

type DatabaseRepository struct {
	ApplicationVersion ApplicationVersionRepository
}

func NewDatabaseInstance(cfg sharedConfig.MongoDBConfig) *DatabaseRepository {
	sharedMongo.Setup(cfg.RootConnectionString(), cfg.DbName, cfg.User, cfg.Password)

	client := sharedMongo.Connect(cfg.ConnectionString())
	db := client.Database(cfg.DbName)

	return &DatabaseRepository{
		ApplicationVersion: newApplicationVersionRepository(db),
	}
}
