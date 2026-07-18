package mongo

import (
	"context"
	"cu-timepad-bot/internal/config"
	"cu-timepad-bot/internal/domain"
	"slices"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoStore struct {
	Database *mongo.Database
	Users    *mongo.Collection
}

func New(ctx context.Context, client *mongo.Client) (*MongoStore, error) {
	cfg := config.GetConfig(ctx)
	database := client.Database(cfg.DBName)
	users_collection := database.Collection("users")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	users_collection.Indexes().CreateOne(ctx, indexModel)

	return &MongoStore{
		Database: database,
		Users:    users_collection,
	}, nil
}

func (st *MongoStore) AddUser(ctx context.Context, userid int64) error {
	user := &domain.User{
		ID: userid,
	}
	_, err := st.Users.InsertOne(ctx, user)
	return err
}

func (st *MongoStore) GetUser(ctx context.Context, userid int64) (*domain.User, error) {
	filter := bson.D{{Key: "id", Value: userid}}
	res := st.Users.FindOne(ctx, filter)
	if res.Err() != nil {
		return nil, domain.ErrUserNotFound
	}

	var user *domain.User
	res.Decode(&user)

	return user, nil
}

func (st *MongoStore) IsSubcribedUser(ctx context.Context, userid int64, eventid domain.EventID) (bool, error) {
	user, err := st.GetUser(ctx, userid)
	if err != nil {
		return false, err
	}
	return slices.Contains(user.SubscribedEvents, eventid), nil
}

func (st *MongoStore) FindUsersWithEvent(ctx context.Context, eventid domain.EventID) []domain.User {
	filter := bson.D{{
		Key:   "subscribed_events",
		Value: bson.D{{Key: "$all", Value: bson.A{eventid}}},
	}}
	cursor, _ := st.Users.Find(ctx, filter)

	users := make([]domain.User, 0, 16)
	cursor.All(ctx, &users)
	return users
}

func (st *MongoStore) AddUserSubscribedEvent(ctx context.Context, userid int64, eventid domain.EventID) error {
	filter := bson.D{{Key: "id", Value: userid}}
	update := bson.D{{Key: "$addToSet",
		Value: bson.D{{Key: "subscribed_events", Value: eventid}},
	}}

	_, err := st.Users.UpdateOne(ctx, filter, update)
	if err != nil {
		return domain.ErrUserNotFound
	}
	return nil
}

func (st *MongoStore) RemoveUserSubscribedEvent(ctx context.Context, userid int64, eventid domain.EventID) error {
	filter := bson.D{{Key: "id", Value: userid}}
	update := bson.D{{Key: "$pull",
		Value: bson.D{{Key: "subscribed_events", Value: eventid}},
	}}

	_, err := st.Users.UpdateOne(ctx, filter, update)
	if err != nil {
		return domain.ErrUserNotFound
	}
	return nil
}
