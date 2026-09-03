package cache

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

type MemCacheUserStore[T store.UserStore] struct {
	Store    T
	memCache *cache.Cache
}

func cacheUserID(userid int64) string {
	return strconv.FormatInt(userid, 10)
}

func NewUserStore[T store.UserStore](ctx context.Context, store T) *MemCacheUserStore[T] {
	cfg := config.GetConfig(ctx)
	mem_cache := cache.New(time.Duration(cfg.MemCacheDuration)*time.Second, 1*time.Second)

	return &MemCacheUserStore[T]{
		store,
		mem_cache,
	}
}

func (st *MemCacheUserStore[T]) AddUser(ctx context.Context, userid int64) error {
	user := &domain.User{
		ID: userid,
	}
	if err := st.Store.AddUser(ctx, userid); err != nil {
		return err
	}
	st.memCache.Add(cacheUserID(userid), user, cache.DefaultExpiration)
	return nil
}

func (st *MemCacheUserStore[T]) GetUser(ctx context.Context, userid int64) (*domain.User, error) {
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

func (st *MemCacheUserStore[T]) IsSubscribedUser(ctx context.Context, userid int64, eventid int64) (bool, error) {
	user, err := st.GetUser(ctx, userid)
	if err != nil {
		return false, err
	}
	return slices.Contains(user.SubscribedEvents, eventid), nil
}

func (st *MemCacheUserStore[T]) FindUsersWithEvent(ctx context.Context, eventid int64) []*domain.User {
	return st.Store.FindUsersWithEvent(ctx, eventid)
}

func (st *MemCacheUserStore[T]) AddUserSubscribedEvent(ctx context.Context, userid int64, eventid int64) error {
	err := st.Store.AddUserSubscribedEvent(ctx, userid, eventid)
	if err != nil {
		return err
	}
	st.memCache.Delete(cacheUserID(userid))
	return nil
}

func (st *MemCacheUserStore[T]) RemoveUserSubscribedEvent(ctx context.Context, userid int64, eventid int64) error {
	err := st.Store.RemoveUserSubscribedEvent(ctx, userid, eventid)
	if err != nil {
		return err
	}
	st.memCache.Delete(cacheUserID(userid))
	return nil
}

func (st *MemCacheUserStore[T]) AddUserBookingData(ctx context.Context, userid int64, data *domain.BookingUserData) error {
	err := st.Store.AddUserBookingData(ctx, userid, data)
	if err != nil {
		return err
	}
	st.memCache.Delete(cacheUserID(userid))
	return nil
}
