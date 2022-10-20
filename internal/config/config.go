package config

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const configFile = "data/config.yml"

type Config struct {
	Token                  string        `yaml:"token"`
	Currencies             []string      `yaml:"currencies"`
	DefaultCurrency        string        `yaml:"currency_default"`
	CurrencyUpdateDuration time.Duration `yaml:"currency_update_period"`
}

type Service struct {
	config Config
}

func New() (*Service, error) {
	s := &Service{}

	rawYAML, err := os.ReadFile(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "reading config file")
	}

	err = yaml.Unmarshal(rawYAML, &s.config)
	if err != nil {
		return nil, errors.Wrap(err, "parsing yaml")
	}

	return s, nil
}

func (s *Service) Token() string {
	return s.config.Token
}

func (s *Service) Currencies() []string {
	return s.config.Currencies
}

func (s *Service) DefaultCurrency() string {
	return s.config.DefaultCurrency
}

func (s *Service) CurrencyUpdateDuration() time.Duration {
	return s.config.CurrencyUpdateDuration
}
