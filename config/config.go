package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP     HTTPConfig     `yaml:"http"`
	Postgres PostgresConfig `yaml:"postgres"`
	Logger   LoggerConfig   `yaml:"logger"`
	JWT      JWTConfig      `yaml:"jwt"`
	MinIO    MinIOConfig    `yaml:"minio"`
}

type HTTPConfig struct {
	Addr            string        `yaml:"addr" env:"HTTP_ADDR"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"HTTP_SHUTDOWN_TIMEOUT"`
}

type PostgresConfig struct {
	Host     string        `yaml:"host" env:"POSTGRES_HOST"`
	Port     string        `yaml:"port" env:"POSTGRES_PORT"`
	User     string        `yaml:"user" env:"POSTGRES_USER"`
	Password string        `yaml:"password" env:"POSTGRES_PASSWORD"`
	DB       string        `yaml:"db" env:"POSTGRES_DB"`
	SSLMode  string        `yaml:"sslmode" env:"POSTGRES_SSLMODE"`
	Timeout  time.Duration `yaml:"timeout" env:"POSTGRES_TIMEOUT"`
}

type LoggerConfig struct {
	Level  string `yaml:"level" env:"LOGGER_LEVEL"`
	Folder string `yaml:"folder" env:"LOGGER_FOLDER"`
}

type JWTConfig struct {
	AccessSecret      string        `yaml:"access_secret" env:"JWT_ACCESS_SECRET"`
	RefreshSecret     string        `yaml:"refresh_secret" env:"JWT_REFRESH_SECRET"`
	AccessExpiration  time.Duration `yaml:"access_expiration" env:"JWT_ACCESS_EXPIRATION"`
	RefreshExpiration time.Duration `yaml:"refresh_expiration" env:"JWT_REFRESH_EXPIRATION"`
}

type MinIOConfig struct {
	Endpoint  string        `yaml:"endpoint" env:"MINIO_ENDPOINT"`
	AccessKey string        `yaml:"access_key" env:"MINIO_ACCESS_KEY"`
	SecretKey string        `yaml:"secret_key" env:"MINIO_SECRET_KEY"`
	Bucket    string        `yaml:"bucket" env:"MINIO_BUCKET"`
	UseSSL    bool          `yaml:"use_ssl" env:"MINIO_USE_SSL"`
	Timeout   time.Duration `yaml:"timeout" env:"MINIO_TIMEOUT"`
}

func NewConfig(configPath string) (*Config, error) {
	cfg := &Config{}

	if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config %q:\n%w", configPath, err)
	}

	return cfg, nil
}
