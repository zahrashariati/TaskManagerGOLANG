//manages configuration for the task manager API (database, redis, port)
package config

import "os"

type Config struct {
	DatabaseURL      string
	RedisURL         string
	Port             string
	JWTPrivateKeyPath string
	JWTPublicKeyPath  string
	KafkaBrokerURL   string
}

func Load() *Config { //Return pointer to Config struct - allows modification if needed
	return &Config{ // Creates pointer to Config struct and returns it
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		RedisURL:         getEnv("REDIS_URL", ""),
		Port:             getEnv("PORT", "8080"),
		JWTPrivateKeyPath: getEnv("JWT_PRIVATE_KEY_PATH", "keys/private.pem"),
		JWTPublicKeyPath:  getEnv("JWT_PUBLIC_KEY_PATH", "keys/public.pem"),
		KafkaBrokerURL:   getEnv("KAFKA_BROKER_URL", "localhost:9092"),
	}
}
//You can't "return the Config" directly — you create a Config and return its address with &.


func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

//If the environment variable is set, use it; otherwise, use the default value.