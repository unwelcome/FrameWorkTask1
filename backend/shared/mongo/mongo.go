package mongo

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect(connectString string) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectString))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to mongo")
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal().Err(err).Msg("failed to ping mongo")
	}

	return client
}
