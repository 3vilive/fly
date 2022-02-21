package config

import "github.com/spf13/viper"

func InitConfig(path string) error {
	viper.SetConfigFile(path)
	return viper.ReadInConfig()
}

func DeinitConfig() error {
	return nil
}
