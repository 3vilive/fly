package storage

import "github.com/spf13/viper"

var (
	dbManager *DatabaseManager
)

func InitDatabase() error {
	dbManager = NewDatabaseManager(viper.Sub("database"))
	return nil
}

func DeinitDatabase() error {
	if dbManager == nil {
		return nil
	}
	return dbManager.Close()
}
