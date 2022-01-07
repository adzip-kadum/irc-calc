package main

import (
	"os"

	"github.com/adzip-kadum/irc-calc/log"
)

func main() {
	rc := 0
	defer func() {
		log.Sync()
		os.Exit(rc)
	}()

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		rc = 1
	}
}
