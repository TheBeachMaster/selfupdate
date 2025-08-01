package config

import "github.com/caarlos0/env/v11"

type Config struct {
	Github GithubConfig
}

type GithubConfig struct {
	AcccessToken string `env:"AccessToken"`
}

func New() *Config {
	return &Config{}
}

func (c *Config) Parse() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil

}
