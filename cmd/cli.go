package cmd

/*
 * File: cmd/cli.go
 *
 * Purpose: defining the command line interface (CLI) for the Tournabyte webapi
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

// Variable `rootCmd` holds the root cobra command for the CLI
var rootCmd *cobra.Command = &cobra.Command{
	Use:               "tbyte-webapi",
	Short:             "controls the webapi for the Tournabyte platform",
	PersistentPreRunE: getConfig,
}

// Variable `serveCmd` holds the serve subcommand for the CLI
var serveCmd *cobra.Command = &cobra.Command{
	Use:   "serve",
	Short: "start the Tournabyte API webserver",
	RunE:  doServe,
}

// Variable `appConfig` holds the application configuration user for CLI invokations
var appConfig *AppConfig = NewAppConfig("json", "appconf", []string{"/etc/tournabyte/webapi", "$HOME/.local/tournabyte/webapi", "."})

// Function `init` initializes the CLI's subcommands and processes any flagset overwrites to the application configuration
func init() {

	serveCmd.Flags().Int("port", 8080, "Port for the application to listen on")
	serveCmd.Flags().StringSlice("dbhosts", []string{"localhost:27017"}, "Comma-separated list of hosts for mongo db persistence functionality")
	serveCmd.Flags().String("dbuser", "", "Database identity to access mongo instance")
	serveCmd.Flags().String("dbpass", "", "Database access key for authenticating to mongo instance")
	serveCmd.Flags().String("s3url", "", "Endpoint for the s3 service instance")
	serveCmd.Flags().String("s3access", "", "Access ID for the s3 instance")
	serveCmd.Flags().String("s3secret", "", "Secret key for the s3 instance")

	optionsToFlags := map[string]string{
		"serve.port":       "port",
		"mongodb.hosts":    "dbhosts",
		"mongodb.username": "dbuser",
		"mongodb.password": "dbpass",
		"minio.endpoint":   "s3url",
		"minio.accessKey":  "s3access",
		"minio.secretKey":  "s3secret",
	}

	appConfig.PopulateFromFlagset(serveCmd.Flags(), optionsToFlags)

	rootCmd.AddCommand(serveCmd)
}

// Function `getConfig` retrieves the application configuration from the viper config file settings
//
// Returns:
//   - `error`: the issue that prevented sourcing application configuration from a file
func getConfig(cmd *cobra.Command, args []string) error {

	if err := appConfig.PopulateFromFile(); err != nil {
		return err
	}
	return nil
}

// Function `doServe` in the handler function for invocations to the serve subcommand
//
// Returns:
//   - `error`: the issue that arose from executing the serve subcommand run function (nil if execution was ok)
func doServe(cmd *cobra.Command, args []string) error {
	if cfg, err := appConfig.UnmarshalOptions(); err != nil {
		return err
	} else {
		log.Printf("Serving with config %+v", cfg)
	}

	return nil
}

// Function `Execute` acts of the CLI entry point for the Tournabyte API webserver
func Execute() {
	appErr := rootCmd.ExecuteContext(context.Background())
	if appErr != nil {
		log.Fatalf("Exit status NOT_OK: %s\n", appErr.Error())
	}
	log.Println("Exit status OK")
}
