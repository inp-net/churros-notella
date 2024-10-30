package notella

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Configuration struct {
	Port               int    `env:"PORT" envDefault:"8080"`
	ChurrosDatabaseURL string `env:"DATABASE_URL"`
	VapidPublicKey     string `env:"PUBLIC_VAPID_KEY"`
	VapidPrivateKey    string `env:"VAPID_PRIVATE_KEY"`
	ContactEmail       string `env:"CONTACT_EMAIL"`
}

func LoadConfiguration() (Configuration, error) {
	config := Configuration{}
	err := env.Parse(&config)
	if err != nil {
		return Configuration{}, fmt.Errorf("could not load env variables: %w", err)
	}

	return config, nil
}

var config Configuration

func init() {
	var err error
	config, err = LoadConfiguration()
	if err != nil {
		panic(fmt.Errorf("could not load configuration: %w", err))
	}
}
