package storage

import "github.com/spf13/viper"

var (
	dbManager    *DatabaseManager
	redisManager *RedisManager
)

func InitStorage() error {
	dbManager = NewDatabaseManager(viper.Sub("database"))
	redisManager = NewRedisManager(viper.Sub("redis"))
	return nil
}

func DeinitStorage() error {
	if dbManager != nil {
		dbManager.Close()
	}
	if redisManager != nil {
		redisManager.Close()
	}

	return nil
}
