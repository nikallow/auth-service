package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env     string   `env:"ENV" env-default:"local"`
	Service Service  `env-prefix:"SERVICE_"`
	Logger  Logger   `env-prefix:"LOG_"`
	HTTP    HTTP     `env-prefix:"HTTP_"`
	GRPC    GRPC     `env-prefix:"GRPC_"`
	PG      Postgres `env-prefix:"PG_"`
	Redis   Redis    `env-prefix:"REDIS_"`
	JWT     JWT      `env-prefix:"JWT_"`
}

type Service struct {
	Name string `env:"NAME" env-default:"auth-service"`
}

type Logger struct {
	Level  string `env:"LEVEL"  env-default:"info"`
	Format string `env:"FORMAT" env-default:"text"`
}

type HTTP struct {
	Port          string `env:"PORT" env-default:"8000"`
	SecureCookies bool   `env:"SECURE_COOKIES" env-default:"false"`
}

type GRPC struct {
	Port string `env:"PORT" env-default:"9000"`
}

type Postgres struct {
	Host     string `env:"HOST"     env-default:"localhost"`
	Port     string `env:"PORT"     env-default:"5432"`
	User     string `env:"USER"     env-default:"postgres"`
	Password string `env:"PASSWORD" env-default:"postgres"`
	DBName   string `env:"DB_NAME"  env-default:"auth_service"`
	SSLMode  string `env:"SSLMODE"  env-default:"disable"`
}

type Redis struct {
	Addr     string `env:"ADDR"     env-default:"localhost:6379"`
	Password string `env:"PASSWORD" env-default:""`
	DB       int    `env:"DB"       env-default:"0"`
}

type JWT struct {
	Secret         string `env:"SECRET"           env-required:"true"`
	AccessTTLMin   int    `env:"ACCESS_TTL_MIN"   env-default:"15"`
	RefreshTTLHour int    `env:"REFRESH_TTL_HOUR" env-default:"168"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("read config from env: %w", err)
	}

	return &cfg, nil
}

func (h HTTP) Address() string {
	return ":" + h.Port
}

func (p Postgres) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		p.Host,
		p.Port,
		p.User,
		p.Password,
		p.DBName,
		p.SSLMode,
	)
}
