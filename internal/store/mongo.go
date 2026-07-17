package store

import (
	"context"
	"cu-timepad-bot/internal/config"
	"cu-timepad-bot/internal/domain"
	"strconv"
	"time"

	"github.com/patrickmn/go-cache"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoStore struct {
	Database *mongo.Database
	Users    *mongo.Collection
	memCache *cache.Cache
}

const cacheDuration = 5 * time.Minute

func cacheUserID(userid int64) string {
	return strconv.FormatInt(userid, 10)
}

func New(ctx context.Context, client *mongo.Client) *MongoStore {
	cfg, _ := config.GetConfig(ctx)
	database := client.Database(cfg.DBName)
	users_collection := database.Collection("users")

	return &MongoStore{
		Database: database,
		Users:    users_collection,
	}
}

func (st *MongoStore) AddUser(ctx context.Context, userid int64) error {
	user := domain.User{
		ID: userid,
	}
	_, err := st.Users.InsertOne(ctx, user)
	st.memCache.Add(cacheUserID(userid), &user, cacheDuration)
	return err
}

func (st *MongoStore) GetUser(ctx context.Context, userid int64) (*domain.User, error) {
	if cachedUser, found := st.memCache.Get(cacheUserID(userid)); found {
		return cachedUser.(*domain.User), nil
	}

	filter := bson.D{{Key: "ID", Value: userid}}
	res := st.Users.FindOne(ctx, filter)
	if res.Err() != nil {
		return nil, domain.ErrUserNotFound
	}

	var user domain.User
	res.Decode(user)
	st.memCache.Add(cacheUserID(userid), user, cacheDuration)

	return &user, nil
}

func (st *MongoStore) FindUsersWithEvent(ctx context.Context, event domain.EventID) []domain.User {
	filter := bson.D{{
		Key:   "subscribed_events",
		Value: bson.D{{Key: "$all", Value: bson.A{event}}},
	}}
	cursor, _ := st.Users.Find(ctx, filter)

	users := make([]domain.User, 0, 16)
	cursor.All(ctx, &users)
	return users
}

func (st *MongoStore) AddUserSubscribedEvent(ctx context.Context, userid int64, event domain.EventID) error {
	filter := bson.D{{Key: "id", Value: userid}}
	update := bson.D{{Key: "subscribed_events",
		Value: bson.D{{Key: "$addToSet", Value: event}},
	}}

	_, err := st.Users.UpdateOne(ctx, filter, update)
	if err != nil {
		return domain.ErrUserNotFound
	}
	return nil
}

func (st *MongoStore) RemoveUserSubscribedEvent(ctx context.Context, userid int64, event domain.EventID) error {
	filter := bson.D{{Key: "id", Value: userid}}
	update := bson.D{{Key: "subscribed_events",
		Value: bson.D{{Key: "$pull", Value: event}},
	}}

	_, err := st.Users.UpdateOne(ctx, filter, update)
	if err != nil {
		return domain.ErrUserNotFound
	}
	return nil
}
