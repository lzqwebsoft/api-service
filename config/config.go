package config

import (
	"encoding/json"
	"fmt"
	"os"

	logger "api-service/utils"
)

type jsonConfig struct {
	ServerPort             string `json:"server_port"`
	DBUser                 string `json:"db_user"`
	DBPass                 string `json:"db_pass"`
	DBHost                 string `json:"db_host"`
	DBPort                 string `json:"db_port"`
	DBName                 string `json:"db_name"`
	DBParams               string `json:"db_params"`
	DBDSN                  string `json:"db_dsn"`
	AccessTokenExpireHours int    `json:"access_token_expire_hours"`
	RefreshTokenExpireDays int    `json:"refresh_token_expire_days"`
}

// Config holds all configuration parameters for the application
type Config struct {
	ServerPort             string
	DBDSN                  string
	AccessTokenExpireHours int
	RefreshTokenExpireDays int
}

// LoadConfig reads configuration from a JSON file, falling back to defaults on error
func LoadConfig() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.json"
	}

	// Initialize with default parameters
	jConfig := jsonConfig{
		ServerPort:             "8080",
		DBUser:                 "root",
		DBPass:                 "root",
		DBHost:                 "127.0.0.1",
		DBPort:                 "3306",
		DBName:                 "api_service",
		DBParams:               "parseTime=true&loc=Local",
		AccessTokenExpireHours: 24,
		RefreshTokenExpireDays: 7,
	}

	// Try reading configuration from the JSON file
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warnf("Config file %s not found. Using default parameters.", configPath)
		} else {
			logger.Warnf("Failed to open config file %s: %v. Using default parameters.", configPath, err)
		}
	} else {
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&jConfig); err != nil {
			logger.Warnf("Failed to parse config JSON in %s: %v. Using default parameters.", configPath, err)
		} else {
			logger.Infof("Successfully loaded configuration from %s", configPath)
		}
	}

	// Determine final DB DSN
	var dbDSN string
	if jConfig.DBDSN != "" {
		dbDSN = jConfig.DBDSN
	} else {
		// Assemble split parameters
		dbDSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
			jConfig.DBUser,
			jConfig.DBPass,
			jConfig.DBHost,
			jConfig.DBPort,
			jConfig.DBName,
			jConfig.DBParams,
		)
	}

	return &Config{
		ServerPort:             jConfig.ServerPort,
		DBDSN:                  dbDSN,
		AccessTokenExpireHours: jConfig.AccessTokenExpireHours,
		RefreshTokenExpireDays: jConfig.RefreshTokenExpireDays,
	}
}
