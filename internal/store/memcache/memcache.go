package memcache

import (
	"context"
	"cu-timepad-bot/internal/config"
	"cu-timepad-bot/internal/domain"
	"cu-timepad-bot/internal/store"
	"slices"
	"strconv"
	"time"

	"github.com/patrickmn/go-cache"
)

type MemCacheStore[T store.Store] struct {
	Store    T
	memCache *cache.Cache
}

func cacheUserID(userid int64) string {
	return strconv.FormatInt(userid, 10)
}

func New[T store.Store](ctx context.Context, store T) *MemCacheStore[T] {
	cfg := config.GetConfig(ctx)
	mem_cache := cache.New(time.Duration(cfg.MemCacheDuration)*time.Second, 1*time.Second)

	return &MemCacheStore[T]{
		store,
		mem_cache,
	}
}

func (st *MemCacheStore[T]) AddUser(ctx context.Context, userid int64) error {
	user := &domain.User{
		ID: userid,
	}
	if err := st.Store.AddUser(ctx, userid); err != nil {
		return err
	}
	st.memCache.Add(cacheUserID(userid), user, cache.DefaultExpiration)
	return nil
}

func (st *MemCacheStore[T]) GetUser(ctx context.Context, userid int64) (*domain.User, error) {
	if cachedUser, found := st.memCache.Get(cacheUserID(userid)); found {
		return cachedUser.(*domain.User), nil
	}

	user, err := st.Store.GetUser(ctx, userid)
	if err != nil {
		return nil, err
	}

	st.memCache.Add(cacheUserID(userid), user, cache.DefaultExpiration)

	return user, nil
}

func (st *MemCacheStore[T]) IsSubcribedUser(ctx context.Context, userid int64, event domain.EventID) (bool, error) {
	user, err := st.GetUser(ctx, userid)
	if err != nil {
		return false, err
	}
	return slices.Contains(user.SubscribedEvents, event), nil
}

func (st *MemCacheStore[T]) FindUsersWithEvent(ctx context.Context, eventid domain.EventID) []domain.User {
	return st.Store.FindUsersWithEvent(ctx, eventid)
}

func (st *MemCacheStore[T]) AddUserSubscribedEvent(ctx context.Context, userid int64, eventid domain.EventID) error {
	err := st.Store.AddUserSubscribedEvent(ctx, userid, eventid)
	if err != nil {
		return err
	}
	st.memCache.Delete(cacheUserID(userid))
	return nil
}

func (st *MemCacheStore[T]) RemoveUserSubscribedEvent(ctx context.Context, userid int64, eventid domain.EventID) error {
	err := st.Store.RemoveUserSubscribedEvent(ctx, userid, eventid)
	if err != nil {
		return err
	}
	st.memCache.Delete(cacheUserID(userid))
	return nil
}
