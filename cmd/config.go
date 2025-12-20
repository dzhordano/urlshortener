package cmd

import (
	"fmt"
	"time"
)

const (
	EnvironmentProduction  = "production"
	EnvironmentDevelopment = "development"
)

type Config struct {
	Environment string
	ServiceName string
	HTTP        HTTPConfig
	DB          DBConfig
	RDB         RedisConfig
	JaegerURL   string
}

type HTTPConfig struct {
	Host string
	Port string
}

func (c *HTTPConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Name,
	)
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	TTL      time.Duration
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
