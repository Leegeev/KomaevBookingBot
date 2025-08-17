package config

import (
	"fmt"
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
	Token       string         `mapstructure:"token"`
	OfficeTZ    *time.Location `mapstructure:"office_tz"`
	GroupChatID int64          `mapstructure:"group_chat_id"` // ID группы для проверки админства
	// BotUsername string `mapstructure:"bot_username"` // если нужно, можно добавить
}

type Config struct {
	DB DB `mapstructure:"database"`
	// Server Server `mapstructure:"server"`
	Telegram Telegram `mapstructure:"telegram"`
}

func LoadConfig(logger logger.Logger) (*Config, error) {
	// Загрузка переменных окружения из .env
	if err := godotenv.Load(); err != nil {
		logger.Error("Failed to load .env file, using default environment variables")
	}

	// Поиск файла конфигурации config.yaml
	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Join(filepath.Dir(b), "../../")
	viper.AddConfigPath(basePath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Подключение переменных окружения
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true)

	// Привязка переменных окружения к ключам
	bindEnv()

	// Чтение и парсинг YAML
	if err := viper.ReadInConfig(); err != nil {
		logger.Error("Error reading config.yaml", "error", err)
		return nil, fmt.Errorf("ошибка чтения config.yaml: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		logger.Error("Error unmarshalling config", "error", err)
		return nil, fmt.Errorf("ошибка декодирования config: %w", err)
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
}

func (c *DB) DSN() string {
	return fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
}
