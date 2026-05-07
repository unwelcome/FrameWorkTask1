package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect(connectString string) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(connectString))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to mongo")
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal().Err(err).Msg("failed to ping mongo")
	}

	return client
}

// Setup создаёт пользователя с правами readWrite на указанную базу данных.
// Идемпотентен: если пользователь уже существует — пропускает создание.
func Setup(adminConnString, dbName, username, password string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(adminConnString))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to mongo as admin")
	}
	defer func() { _ = client.Disconnect(ctx) }()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal().Err(err).Msg("failed to ping mongo as admin")
	}

	cmd := bson.D{
		{Key: "createUser", Value: username},
		{Key: "pwd", Value: password},
		{Key: "roles", Value: bson.A{
			bson.D{
				{Key: "role", Value: "readWrite"},
				{Key: "db", Value: dbName},
			},
		}},
	}

	err = client.Database(dbName).RunCommand(ctx, cmd).Err()
	if err != nil {
		var cmdErr mongo.CommandError
		if errors.As(err, &cmdErr) && cmdErr.Code == 51003 {
			log.Info().Str("user", username).Str("db", dbName).Msg("mongo user already exists, skipping")
			return
		}
		log.Fatal().Err(err).Msg("failed to create mongo user")
	}

	log.Info().Str("user", username).Str("db", dbName).Msg("mongo user created")
}
