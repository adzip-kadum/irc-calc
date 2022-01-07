package main

import (
	"github.com/spf13/cobra"

	"github.com/adzip-kadum/irc-calc/app"
	"github.com/adzip-kadum/irc-calc/log"
	"github.com/adzip-kadum/irc-calc/worker"
)

var someFlag *bool

func init() {
	someFlag = serveCmd.Flags().BoolP("some", "", false, "Some flag")
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start serving",
	RunE: func(*cobra.Command, []string) error {
		app, err := app.New(&Config)
		if err != nil {
			return err
		}

		err = app.Start()
		if err != nil {
			return err
		}

		err = worker.Wait()
		if err != nil {
			log.Error(err)
		}

		return app.Stop()
	},
}
