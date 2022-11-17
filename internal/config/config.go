package config

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const configFile = "data/config/config.yml"

type Config struct {
	Token                  string        `yaml:"token"`
	Currencies             []string      `yaml:"currencies"`
	DefaultCurrency        string        `yaml:"currency_default"`
	CurrencyUpdateDuration time.Duration `yaml:"currency_update_period"`
	DBPort                 string        `yaml:"db_port"`
	DBHost                 string        `yaml:"db_host"`
	DBPassword             string        `yaml:"db_password"`
	DBUser                 string        `yaml:"db_user"`
	DBName                 string        `yaml:"db_name"`
	DevMode                string        `yaml:"dev_mode"`
	ServiceName            string        `yaml:"service_name"`
	JaegerHostPort         string        `yaml:"jaeger_host_port"`
	Port                   int64         `yaml:"port"`
	RedisHostPort          string        `yaml:"redis_host_port"`
	RedisPassword          string        `yaml:"redis_password"`
	RedisDb                int           `yaml:"redis_db"`
	BotHost                string        `yaml:"bot_host"`
	BotGrpcPort            int64         `yaml:"bot_grpc_port"`
	BotHttpPort            int64         `yaml:"bot_http_port"`
	KafkaHost              string        `yaml:"kafka_host"`
	KafkaPort              int64         `yaml:"kafka_port"`
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

func (s *Service) DBPort() string {
	return s.config.DBPort
}

func (s *Service) DBPassword() string {
	return s.config.DBPassword
}

func (s *Service) DBUser() string {
	return s.config.DBUser
}

func (s *Service) DBHost() string {
	return s.config.DBHost
}

func (s *Service) DBName() string {
	return s.config.DBName
}

func (s *Service) DevMode() string {
	return s.config.DevMode
}

func (s *Service) ServiceName() string {
	return s.config.ServiceName
}

func (s *Service) JaegerHostPort() string {
	return s.config.JaegerHostPort
}

func (s *Service) Port() int64 {
	return s.config.Port
}

func (s *Service) RedisHostPort() string {
	return s.config.RedisHostPort
}

func (s *Service) RedisDB() int {
	return s.config.RedisDb
}

func (s *Service) RedisPassword() string {
	return s.config.RedisPassword
}

func (s *Service) BotHost() string {
	return s.config.BotHost
}

func (s *Service) BotGrpcPort() int64 {
	return s.config.BotGrpcPort
}

func (s *Service) BotHttpPort() int64 {
	return s.config.BotHttpPort
}

func (s *Service) KafkaHost() string {
	return s.config.KafkaHost
}

func (s *Service) KafkaPort() int64 {
	return s.config.KafkaPort
}
