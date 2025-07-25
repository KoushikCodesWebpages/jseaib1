package config

import "github.com/spf13/viper"

type ServerConfig struct {
    ServerPort                      int
    ServerHost                      string
    LogLevel                        string
    RateLimit                       int
    Environment                     string

}

func LoadServerConfig() (*ServerConfig, error) {
    ServerConfig := &ServerConfig{

        ServerPort:                 viper.GetInt("SERVER_PORT"),
        ServerHost:                 viper.GetString("SERVER_HOST"),
        LogLevel:                   viper.GetString("LOG_LEVEL"),
        RateLimit:                  viper.GetInt("RATE_LIMIT"),
        Environment:                viper.GetString("ENVIRONMENT"),

    }

    return ServerConfig, nil
}
