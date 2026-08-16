package config

import (
	"errors"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment    string
	HTTPAddr       string
	DatabaseURL    string
	JWTSecret      string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	RequestTimeout time.Duration
	AllowInMemory  bool
	CORSOrigins    []string
	LogLevel       string
}

func Load() (Config, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("HTTP_ADDR", ":8080")
	v.SetDefault("ACCESS_TTL", "15m")
	v.SetDefault("REFRESH_TTL", "168h")
	v.SetDefault("REQUEST_TIMEOUT", "5s")
	v.SetDefault("ALLOW_IN_MEMORY", false)
	v.SetDefault("CORS_ORIGINS", "http://localhost:5173")
	v.SetDefault("LOG_LEVEL", "info")
	_ = v.ReadInConfig()
	accessTTL, err := time.ParseDuration(v.GetString("ACCESS_TTL"))
	if err != nil {
		return Config{}, err
	}
	refreshTTL, err := time.ParseDuration(v.GetString("REFRESH_TTL"))
	if err != nil {
		return Config{}, err
	}
	timeout, err := time.ParseDuration(v.GetString("REQUEST_TIMEOUT"))
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Environment: v.GetString("APP_ENV"), HTTPAddr: v.GetString("HTTP_ADDR"), DatabaseURL: v.GetString("DATABASE_URL"), JWTSecret: v.GetString("JWT_SECRET"),
		AccessTTL: accessTTL, RefreshTTL: refreshTTL, RequestTimeout: timeout, AllowInMemory: v.GetBool("ALLOW_IN_MEMORY"), CORSOrigins: strings.Split(v.GetString("CORS_ORIGINS"), ","), LogLevel: v.GetString("LOG_LEVEL"),
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must contain at least 32 characters")
	}
	if cfg.DatabaseURL == "" && !cfg.AllowInMemory {
		return Config{}, errors.New("DATABASE_URL is required unless ALLOW_IN_MEMORY=true")
	}
	return cfg, nil
}
