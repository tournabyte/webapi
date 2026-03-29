package cmd

/*
 * File: cmd/server.go
 *
 * Purpose: define the server subcommand instance for the CLI to the webapi application
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"errors"
	"log"

	"github.com/spf13/cobra"
)

// Server command level CLI constants
const (
	serverCmdUsageMsg  = "server"
	serverCmdShortHelp = "Starts the Tournabyte web API server process"

	listeningPortFlag            = "port"
	listeningPortDefaultValue    = 8080
	listeningPortHelpMsg         = "The port the API server process should listen on"
	listeningPortFlagOverrideKey = "serve.port"
)

var (
	serverPort *int
)

// Variable `serverCmd` holds a pointer to the server `cobra.Command` struct representing the server subcommand CLI
var serverCmd *cobra.Command = &cobra.Command{
	Use:   serverCmdUsageMsg,
	Short: serverCmdShortHelp,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if appConfig == nil {
			return errors.New("Application configuration manager is not present")
		}
		return appConfig.Bind(
			OverrideFromFlag(listeningPortFlagOverrideKey, cmd.Flag(listeningPortFlag)),
		)
	},
	RunE: doServe,
}

func init() {
	serverPort = serverCmd.Flags().Int(
		listeningPortFlag,
		listeningPortDefaultValue,
		listeningPortHelpMsg,
	)
}

func doServe(cmd *cobra.Command, args []string) error {
	log.Printf("Starting server process...")
	return nil
}
