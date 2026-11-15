package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL     string
	CurationPath    string
	TextScoresPath  string
	HTTPAddr        string
	StartupTimeout  time.Duration
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	if err := loadEnvFile(); err != nil {
		return Config{}, err
	}
	cfg := Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		HTTPAddr:       "127.0.0.1:8080",
		CurationPath:   "../data/catalog_curation.json",
		TextScoresPath: "../data/ranking_text_scores.json",
	}
	if value, ok := os.LookupEnv("HTTP_ADDR"); ok {
		cfg.HTTPAddr = value
	}
	if value, ok := os.LookupEnv("CATALOG_CURATION_PATH"); ok {
		cfg.CurationPath = value
	}
	if value, ok := os.LookupEnv("TEXT_SCORES_PATH"); ok {
		cfg.TextScoresPath = value
	}
	var err error
	cfg.StartupTimeout, err = duration("STARTUP_TIMEOUT", 30*time.Second)
	if err != nil {
		return Config{}, err
	}
	cfg.ShutdownTimeout, err = duration("SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return errors.New("DATABASE_URL is required")
	}
	if strings.TrimSpace(c.TextScoresPath) == "" {
		return errors.New("TEXT_SCORES_PATH is required")
	}
	_, port, err := net.SplitHostPort(c.HTTPAddr)
	if err != nil {
		return errors.New("HTTP_ADDR must be host:port")
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 0 || number > 65535 {
		return errors.New("HTTP_ADDR must contain a numeric port in 0..65535")
	}
	if c.StartupTimeout <= 0 {
		return errors.New("STARTUP_TIMEOUT must be positive")
	}
	if c.ShutdownTimeout <= 0 {
		return errors.New("SHUTDOWN_TIMEOUT must be positive")
	}
	return nil
}

func duration(name string, fallback time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", name)
	}
	return parsed, nil
}

func loadEnvFile() error {
	for _, path := range []string{".env", "../.env"} {
		err := godotenv.Load(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("cannot read or parse %s", path)
		}
		return nil
	}
	return nil
}
