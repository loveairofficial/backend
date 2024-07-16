package mongo

import (
	"context"
	"loveair/base/data"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	LADB           = "LADB"
	UserCLX        = "UserCLX"
	ChatCLX        = "ChatCLX"
	ReportCLX      = "ReportCLX"
	FeedbackCLX    = "FeedbackCLX"
	TransactionCLX = "TransactionCLX"
	ConfigCLX      = "ConfigCLX"
)

type MongoDB struct {
	client *mongo.Client
}

// InitMongoDBInstance is a constructor function for establishing a connection to a Mongo database
func InitMongoDBInstance(Config map[string]string) (data.Interface, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(Config["url"]))
	// client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(Config["url"]).SetAuth(options.Credential{
	// 	Username: Config["username"],
	// 	Password: Config["password"],
	// }))

	return &MongoDB{
		client,
	}, err
}

// Creting context
func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (mgoConnection *MongoDB) Updater(clx string, filter, update bson.M) error {
	ctx, cancel := getContext()
	defer cancel()

	database := mgoConnection.client.Database(LADB)
	collection := database.Collection(clx)

	d := collection.FindOneAndUpdate(
		ctx,
		filter,
		update,
	)
	return d.Err()
}
