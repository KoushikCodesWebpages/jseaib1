package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server  *ServerConfig
	Cloud   *CloudConfig
	Project *ProjectConfig
}

var Cfg *Config

func InitConfig() error {
	// RemoveSystemEnv()

	// Load the .env file only if not running in Railway
	if _, exists := os.LookupEnv("RAILWAY_ENVIRONMENT"); !exists {
		err := godotenv.Load()
		if err != nil {
			log.Println("No .env file found, relying on system environment variables")
		}
	}

	// Always enable Viper to read from environment variables
	viper.AutomaticEnv()

	// Load different configuration components
	server, err := LoadServerConfig()
	if err != nil {
		return fmt.Errorf("error loading server config: %v", err)
	}
	cloud, err := LoadCloudConfig()
	if err != nil {
		return fmt.Errorf("error loading cloud config: %v", err)
	}
	project, err := LoadProjectConfig()
	if err != nil {
		return fmt.Errorf("error loading project config: %v", err)
	}

	// Set the global config variable
	Cfg = &Config{
		Server:  server,
		Cloud:   cloud,
		Project: project,
	}

	return nil
}

// Optional: Clear environment variables for testing or CLI tools
func RemoveSystemEnv() {
	for _, pair := range os.Environ() {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			os.Unsetenv(kv[0])
		}
	}
}
