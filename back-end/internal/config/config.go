package config

import (
	"net"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Config struct {
	Port        string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	// DBTLS turns on TLS for the database connection. Hosted MySQL services such
	// as TiDB Cloud require it; the local Docker MySQL does not.
	DBTLS       bool
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
		DBTLS:       env("DB_TLS", "false") == "true",
		SeedData:    env("SEED_DATA", "true") == "true",
		CORSOrigins: env("CORS_ORIGINS", "*"),
	}
}

func (c Config) DSN() string {
	// mysql.Config escapes the password, so special characters in it are safe.
	m := mysql.NewConfig()
	m.User = c.DBUser
	m.Passwd = c.DBPassword
	m.Net = "tcp"
	m.Addr = net.JoinHostPort(c.DBHost, c.DBPort)
	m.DBName = c.DBName
	m.ParseTime = true
	m.Loc = time.UTC
	m.Params = map[string]string{"charset": "utf8mb4"}
	if c.DBTLS {
		m.TLSConfig = "true" // verify the server certificate against the system roots
	}
	return m.FormatDSN()
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}
