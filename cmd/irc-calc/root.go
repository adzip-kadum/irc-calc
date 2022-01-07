package main

import (
	"github.com/spf13/cobra"

	"github.com/adzip-kadum/irc-calc/app"
	"github.com/adzip-kadum/irc-calc/config"
	"github.com/adzip-kadum/irc-calc/log"
	"github.com/adzip-kadum/irc-calc/version"
)

var Config app.Config

func init() {
	rootCmd.PersistentFlags().StringVar(&config.ConfigFile, "conf", "", "Config file")
}

var rootCmd = &cobra.Command{
	Use: version.Project,
	PersistentPreRunE: func(*cobra.Command, []string) error {
		if err := config.Init(&Config); err != nil {
			return err
		}
		return log.Init(Config.Logger)
	},
}
