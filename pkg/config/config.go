package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/joho/godotenv"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
	"github.com/spf13/viper"
)

type DB struct {
	Host       string        `mapstructure:"host"`
	Port       string        `mapstructure:"port"`
	User       string        `mapstructure:"user"`
	Password   string        `mapstructure:"password"`
	DBName     string        `mapstructure:"name"`
	SSLMode    string        `mapstructure:"sslmode"`
	RetryCount int           `mapstructure:"retry_count"`
	RetryDelay time.Duration `mapstructure:"retry_delay"`
}

type Telegram struct {
	Token       string `mapstructure:"token"`
	OfficeTZ    *time.Location
	tzinString  string `mapstructure:"office_tz"`
	GroupChatID int64  `mapstructure:"group_chat_id"` // ID группы для проверки админства
	AdminID     int64  `mapstructure:"admin_id"`      // ID админа для уведомлений
}

type Config struct {
	DB       DB       `mapstructure:"database"`
	Telegram Telegram `mapstructure:"telegram"`
}

// pkg/config/config.go
func LoadConfig(logger logger.Logger) (*Config, error) {
	// .env опционален — если нет файла, просто логируем и продолжаем
	_ = godotenv.Load()

	viper.SetConfigType("yaml")

	// 1) Явный путь из окружения
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		viper.SetConfigFile(p)
	} else {
		viper.AddConfigPath(".")    // текущая рабочая директория
		viper.AddConfigPath("/src") // твой WORKDIR в Docker
		if _, b, _, ok := runtime.Caller(0); ok && filepath.IsAbs(b) {
			basePath := filepath.Join(filepath.Dir(b), "../../")
			viper.AddConfigPath(basePath)
			logger.Info("Using config file from base path", "path", basePath)
		}
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true)
	bindEnv()

	if err := viper.ReadInConfig(); err != nil {
		logger.Error("Error reading config.yaml", "error", err)
		return nil, fmt.Errorf("ошибка чтения config.yaml: %w", err)
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		logger.Error("Error unmarshalling config", "error", err)
		return nil, fmt.Errorf("ошибка декодирования config: %w", err)
	}
	locStr := cfg.Telegram.tzinString
	if locStr != "" {
		loc, err := time.LoadLocation(locStr)
		if err != nil {
			logger.Error("не удалось загрузить часовой пояс", "error", err)
			return nil, fmt.Errorf("не удалось загрузить часовой пояс: %w", err)
		}
		cfg.Telegram.OfficeTZ = loc
	}
	return &cfg, nil
}

func bindEnv() {
	// Привязка переменных окружения к ключам viper
	// Database
	_ = viper.BindEnv("database.user", "POSTGRES_USER")
	_ = viper.BindEnv("database.password", "POSTGRES_PASSWORD")
	_ = viper.BindEnv("database.name", "POSTGRES_DB")
	_ = viper.BindEnv("database.host", "POSTGRES_HOST")
	_ = viper.BindEnv("database.port", "POSTGRES_PORT")

	// Telegram
	_ = viper.BindEnv("telegram.token", "TELEGRAM_TOKEN")
	_ = viper.BindEnv("telegram.tzinString", "TELEGRAM_OFFICE_TZ")
	_ = viper.BindEnv("telegram.group_chat_id", "TELEGRAM_GROUP_CHAT_ID")
	_ = viper.BindEnv("telegram.admin_id", "TELEGRAM_ADMIN_ID")

}

func (c *DB) DSN() string {
	return fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
}
