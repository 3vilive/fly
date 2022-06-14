package storage

import (
	"context"
	"sync"

	"github.com/3vilive/fly/pkg/flylog"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type WrapRedis struct {
	client *redis.Client
}

func (w *WrapRedis) Close() error {
	if w == nil {
		return nil
	}

	return w.client.Close()
}

type RedisManager struct {
	config   *viper.Viper
	redisMap map[string]*WrapRedis
	mu       sync.Mutex
	closed   atomicBool
}

func NewRedisManager(config *viper.Viper) *RedisManager {
	return &RedisManager{
		config:   config,
		redisMap: make(map[string]*WrapRedis),
	}
}

func (m *RedisManager) GetRedis(name string) (*redis.Client, error) {
	if m.closed.isSet() {
		return nil, errors.New("redis manager closed")
	}

	wrapClient, ok := m.redisMap[name]
	if ok {
		return wrapClient.client, nil
	}

	// init db
	configOfRedis := m.config.Sub(name)
	if configOfRedis == nil {
		return nil, errors.New("no redis config")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// check again, maybe other goroutine init redis
	if wrapClient, ok := m.redisMap[name]; ok {
		return wrapClient.client, nil
	}

	// init redis client
	client := redis.NewClient(&redis.Options{
		Addr:     configOfRedis.GetString("address"),
		Password: configOfRedis.GetString("password"),
		DB:       configOfRedis.GetInt("db"),
	})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		return nil, errors.Wrap(err, "ping redis error")
	}

	m.redisMap[name] = &WrapRedis{
		client: client,
	}

	return client, nil
}

func (m *RedisManager) Close() error {
	if m.closed.isSet() {
		return errors.New("database manager closed")
	}
	m.closed.setTrue()

	m.mu.Lock()
	defer m.mu.Unlock()
	for name, wrapRedis := range m.redisMap {
		if err := wrapRedis.Close(); err != nil {
			flylog.Error("close redis error", zap.String("redis", name), zap.Error(err))
		}

		delete(m.redisMap, name)
	}

	return nil
}

func GetRedis(name string) (*redis.Client, error) {
	if redisManager == nil {
		return nil, errors.New("redis manager not init")
	}
	return redisManager.GetRedis(name)
}
