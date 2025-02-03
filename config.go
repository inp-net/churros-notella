package notella

import (
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	ll "github.com/gwennlbh/label-logger-go"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type Configuration struct {
	ChurrosDatabaseURL         string `env:"DATABASE_URL"`
	RedisURL                   string `env:"REDIS_URL"`
	NatsURL                    string `env:"NATS_URL" envDefault:"nats://localhost:4222"`
	VapidPublicKey             string `env:"PUBLIC_VAPID_KEY"`
	VapidPrivateKey            string `env:"VAPID_PRIVATE_KEY"`
	ContactEmail               string `env:"CONTACT_EMAIL"`
	FirebaseServiceAccount     string `env:"FIREBASE_SERVICE_ACCOUNT"`
	StartupScheduleRestoration string `env:"STARTUP_SCHEDULE_RESTORATION" envDefault:"enabled"`
	AppPackageId               string `env:"APP_PACKAGE_ID" envDefault:"app.churros"`
	HealthCheckPort            int    `env:"HEALTH_CHECK_PORT" envDefault:"8080"`
}

func LoadConfiguration() (Configuration, error) {
	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			ll.ErrorDisplay("could not load .env file", err)
		}
		ll.Info("loaded .env file")
	}

	config := Configuration{}
	err := env.Parse(&config)
	if err != nil {
		return Configuration{}, fmt.Errorf("could not load env variables: %w", err)
	}

	if config.StartupScheduleRestoration != "enabled" && config.StartupScheduleRestoration != "disabled" && config.StartupScheduleRestoration != "eager" {
		return Configuration{}, fmt.Errorf("invalid value for STARTUP_SCHEDULE_RESTORATION: %q, should be one of \"enabled\", \"disabled\", \"eager\"", config.StartupScheduleRestoration)
	}

	ll.Log("Loaded", "green", "configuration from environment")

	return config, nil
}

var config Configuration

func init() {
	var err error
	config, err = LoadConfiguration()
	if err != nil {
		panic(fmt.Errorf("could not load configuration: %w", err))
	}

	err = setupFirebaseClient()
	if err != nil {
		ll.ErrorDisplay("could not setup firebase client", err)
	} else {
		ll.Log("Initialized", "cyan", "firebase client")
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr: strings.TrimPrefix(config.RedisURL, "redis://"),
	})
	ll.Log("Initialized", "cyan", "redis client")
}
