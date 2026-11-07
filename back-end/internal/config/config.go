package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	SeedData    bool
	CORSOrigins string
}

func Load() Config {
	return Config{
		Port:        env("PORT", "8080"),
		DBHost:      env("DB_HOST", "127.0.0.1"),
		DBPort:      env("DB_PORT", "3306"),
		DBUser:      env("DB_USER", "goodbuy"),
		DBPassword:  env("DB_PASSWORD", "goodbuy"),
		DBName:      env("DB_NAME", "goodbuy"),
		SeedData:    env("SEED_DATA", "true") == "true",
		CORSOrigins: env("CORS_ORIGINS", "*"),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}
