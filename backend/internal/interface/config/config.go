package config

import (
	"os"
)

// Config はアプリケーションの設定を管理する
type Config struct {
	GeminiAPIKey     string
	GRPCPort         string
	HTTPPort         string
	CORSAllowOrigins string
	LogLevel         string
}

// LoadConfig は環境変数から設定を読み込む
func LoadConfig() *Config {
	return &Config{
		GeminiAPIKey:     getEnv("GEMINI_API_KEY", ""),
		GRPCPort:         getEnv("GRPC_PORT", "8080"),
		HTTPPort:         getEnv("HTTP_PORT", "8081"),
		CORSAllowOrigins: getEnv("CORS_ALLOW_ORIGINS", "*"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
	}
}

// getEnv は環境変数を取得し、存在しない場合はデフォルト値を返す
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
