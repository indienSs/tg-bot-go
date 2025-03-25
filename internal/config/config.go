package config

type Config struct {
	TelegramToken string
	Postgres      PostgresConfig
	Redis         RedisConfig
	OpenAI        OpenAIConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type OpenAIConfig struct {
	APIKey      string
	Model       string
	Temperature float32
	MaxTokens   int
}