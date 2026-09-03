package mongo

import (
	"context"
	"cu-timepad-bot/internal/config"
	"cu-timepad-bot/internal/domain"
	"cu-timepad-bot/internal/store"
	"errors"
	"slices"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoUserStore struct {
	Users *mongo.Collection
}

func NewUserStore(ctx context.Context, client *mongo.Client) (*MongoUserStore, error) {
	cfg := config.GetConfig(ctx)
	database := client.Database(cfg.DBName)
	users_collection := database.Collection("users")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	users_collection.Indexes().CreateOne(ctx, indexModel)

	return &MongoUserStore{
		Users: users_collection,
	}, nil
}

func (st *MongoUserStore) AddUser(ctx context.Context, userid int64) error {
	user := &domain.User{
		ID: userid,
	}
	_, err := st.Users.InsertOne(ctx, user)
	return err
}

func (st *MongoUserStore) GetUser(ctx context.Context, userid int64) (*domain.User, error) {
	filter := bson.D{{Key: "id", Value: userid}}
	res := st.Users.FindOne(ctx, filter)
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, store.ErrNotFound
		}
		return nil, store.ErrTransient
	}

	var user *domain.User
	err := res.Decode(&user)
	if err != nil {
		return nil, store.ErrTransient
	}

	return user, nil
}

func (st *MongoUserStore) IsSubscribedUser(ctx context.Context, userid int64, eventid int64) (bool, error) {
	user, err := st.GetUser(ctx, userid)
	if err != nil {
		return false, err
	}
	return slices.Contains(user.SubscribedEvents, eventid), nil
}

func (st *MongoUserStore) FindUsersWithEvent(ctx context.Context, eventid int64) []*domain.User {
	filter := bson.D{{
		Key:   "subscribed_events",
		Value: bson.D{{Key: "$all", Value: bson.A{eventid}}},
	}}
	cursor, _ := st.Users.Find(ctx, filter)

	users := make([]*domain.User, 0, 16)
	cursor.All(ctx, &users)
	return users
}

func (st *MongoUserStore) AddUserSubscribedEvent(ctx context.Context, userid int64, eventid int64) error {
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

func (st *MongoUserStore) RemoveUserSubscribedEvent(ctx context.Context, userid int64, eventid int64) error {
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

func (st *MongoUserStore) AddUserBookingData(ctx context.Context, userid int64, data *domain.BookingUserData) error {
	filter := bson.D{{Key: "id", Value: userid}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "bookinguserdata", Value: data}}}}

	_, err := st.Users.UpdateOne(ctx, filter, update)
	if err != nil {
		return domain.ErrUserNotFound
	}
	return nil
}
