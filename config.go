package notella

import (
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	ll "github.com/ewen-lbh/label-logger-go"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type Configuration struct {
	Port                   int    `env:"PORT" envDefault:"8080"`
	ChurrosDatabaseURL     string `env:"DATABASE_URL"`
	RedisURL               string `env:"REDIS_URL"`
	VapidPublicKey         string `env:"PUBLIC_VAPID_KEY"`
	VapidPrivateKey        string `env:"VAPID_PRIVATE_KEY"`
	ContactEmail           string `env:"CONTACT_EMAIL"`
	FirebaseServiceAccount string `env:"FIREBASE_SERVICE_ACCOUNT"`
	AppPackageId           string `env:"APP_PACKAGE_ID" envDefault:"app.churros"`
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
		panic(fmt.Errorf("could not setup firebase client: %w", err))
	}
	ll.Log("Initialized", "cyan", "firebase client")

	redisClient = redis.NewClient(&redis.Options{
		Addr: strings.TrimPrefix(config.RedisURL, "redis://"),
	})
	ll.Log("Initialized", "cyan", "redis client")
}
