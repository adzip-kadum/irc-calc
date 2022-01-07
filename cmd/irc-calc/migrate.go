package main

import (
	"github.com/spf13/cobra"

	pmigrations "github.com/adzip-kadum/irc-calc/migrations"
	"github.com/adzip-kadum/irc-calc/postgres"
)

var migrateVersion *int32

func init() {
	migrateVersion = migrateCmd.Flags().Int32P("version", "v", 0, "Target migrate version")
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migrations",
	RunE: func(*cobra.Command, []string) error {
		return postgres.MigrateTo(Config.Postgres, pmigrations.Migrations, *migrateVersion)
	},
}
