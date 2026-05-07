package mongoDB

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/unwelcome/FrameWorkTask1/backend/application/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
)

const applicationVersionsCollection = "application_versions"

type ApplicationVersionRepository interface {
	SaveApplicationVersion(ctx context.Context, version *entities.ApplicationVersion) Error.CodeError
	GetApplicationVersions(ctx context.Context, applicationUUID string) ([]*entities.ApplicationVersion, Error.CodeError)
}

type applicationVersionRepository struct {
	collection *mongo.Collection
}

func newApplicationVersionRepository(db *mongo.Database) ApplicationVersionRepository {
	collection := db.Collection(applicationVersionsCollection)

	// Индекс по application_uuid для быстрой выборки истории
	_, _ = collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "application_uuid", Value: 1}, {Key: "version", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	return &applicationVersionRepository{collection: collection}
}

// SaveApplicationVersion сохраняет снимок заявки перед изменением
func (r *applicationVersionRepository) SaveApplicationVersion(ctx context.Context, version *entities.ApplicationVersion) Error.CodeError {
	_, err := r.collection.InsertOne(ctx, version)
	if err != nil {
		return Error.CodeError{Code: 0, Err: err}
	}
	return Error.CodeError{Code: -1, Err: nil}
}

// GetApplicationVersions возвращает всю историю версий заявки, отсортированную по версии
func (r *applicationVersionRepository) GetApplicationVersions(ctx context.Context, applicationUUID string) ([]*entities.ApplicationVersion, Error.CodeError) {
	filter := bson.D{{Key: "application_uuid", Value: applicationUUID}}
	opts := options.Find().SetSort(bson.D{{Key: "version", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, Error.CodeError{Code: 0, Err: err}
	}
	defer cursor.Close(ctx)

	versions := make([]*entities.ApplicationVersion, 0)
	if err = cursor.All(ctx, &versions); err != nil {
		return nil, Error.CodeError{Code: 0, Err: err}
	}

	return versions, Error.CodeError{Code: -1, Err: nil}
}
