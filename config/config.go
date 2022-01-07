package config

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	ConfigPath = "./conf"
	ConfigName = "config"
	ConfigFile = ""
	EnvPrefix  = ""
)

func Init(c interface{}) error {
	viper.SetConfigName(ConfigName)
	viper.AddConfigPath(ConfigPath)
	if ConfigFile != "" {
		viper.SetConfigFile(ConfigFile)
	}

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix(EnvPrefix)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return errors.WithStack(err)
	}

	configFileUsed := viper.ConfigFileUsed()

	if c, ok := c.(interface{ SetConfigFileUsed(string) }); ok {
		c.SetConfigFileUsed(configFileUsed)
	}

	return viper.Unmarshal(&c)
}
