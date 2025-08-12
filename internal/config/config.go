package config

import (
	"os"
	"strconv"
	"strings"
	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort       int
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string
	KafkaBrokers     []string
	KafkaTopic       string
	KafkaGroupID     string
}

func Load() (*Config, error) {
	// загрузка данных из окружения 
	_ = godotenv.Load()

	port := 8080
	if p := os.Getenv("SERVER_PORT"); p != "" {
		if pv, err := strconv.Atoi(p); err == nil {
			port = pv
		}
	}

	brokersEnv := os.Getenv("KAFKA_BROKERS")
	var brokers []string
	if brokersEnv != "" {
		for _, b := range strings.Split(brokersEnv, ",") {
			if tb := strings.TrimSpace(b); tb != "" {
				brokers = append(brokers, tb)
			}
		}
	}

	return &Config{
		ServerPort:       port,
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresDB:       os.Getenv("POSTGRES_DB"),
		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresPort:     os.Getenv("POSTGRES_PORT"),
		KafkaBrokers:     brokers,
		KafkaTopic:       os.Getenv("KAFKA_TOPIC"),
		KafkaGroupID:     os.Getenv("KAFKA_GROUP_ID"),
	}, nil
}
