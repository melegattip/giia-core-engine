package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config interface {
	Get(key string) interface{}
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetFloat64(key string) float64
	IsSet(key string) bool
	Validate(requiredKeys []string) error
}

type ViperConfig struct {
	viper *viper.Viper
}

func New(envPrefix string) (*ViperConfig, error) {
	v := viper.New()

	v.SetConfigType("env")
	v.SetConfigName(".env")
	v.AddConfigPath(".")
	v.AddConfigPath("./")

	v.AutomaticEnv()
	if envPrefix != "" {
		v.SetEnvPrefix(envPrefix)
	}
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	return &ViperConfig{viper: v}, nil
}

func (c *ViperConfig) Get(key string) interface{} {
	return c.viper.Get(key)
}

func (c *ViperConfig) GetString(key string) string {
	return c.viper.GetString(key)
}

func (c *ViperConfig) GetInt(key string) int {
	return c.viper.GetInt(key)
}

func (c *ViperConfig) GetBool(key string) bool {
	return c.viper.GetBool(key)
}

func (c *ViperConfig) GetFloat64(key string) float64 {
	return c.viper.GetFloat64(key)
}

func (c *ViperConfig) IsSet(key string) bool {
	return c.viper.IsSet(key)
}

func (c *ViperConfig) Validate(requiredKeys []string) error {
	var missingKeys []string

	for _, key := range requiredKeys {
		if !c.IsSet(key) {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) > 0 {
		return fmt.Errorf("missing required configuration keys: %s", strings.Join(missingKeys, ", "))
	}

	return nil
}
