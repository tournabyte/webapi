package app

/*
 * File: app/cli.go
 *
 * Purpose: defining the command line interface (CLI) for the Tournabyte webapi
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/tournabyte/webapi/internal/domains/user"
	"github.com/tournabyte/webapi/internal/utils"
)

// Variable `rootCmd` holds the root cobra command for the CLI
var rootCmd *cobra.Command = &cobra.Command{
	Use:               "tbyte-webapi",
	Short:             "controls the webapi for the Tournabyte platform",
	PersistentPreRunE: initAppContext,
}

// Variable `appConfig` holds the application configuration user for CLI invokations
var appConfig *AppConfig = NewAppConfig("json", "appconf", []string{"/etc/tournabyte/webapi", "$HOME/.local/tournabyte/webapi", "."})

// Function `init` initializes the CLI's subcommands
func init() {
	rootCmd.AddCommand(initServeCmd())
}

// Function `initServeCmd` initializes the serve subcommand for the CLI
//
// Returns:
//   - `*cobra.Command`: pointer to the initialized command ready to be added as a child of the root command
func initServeCmd() *cobra.Command {
	var serveCmd *cobra.Command = &cobra.Command{
		Use:   "serve",
		Short: "start the Tournabyte API webserver",
		RunE:  doServe,
	}

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

	return serveCmd
}

// Function `initAppContext` initializes the application context for command execution
//
// Returns:
//   - `error`: the issue that prevented sourcing application configuration from a file
func initAppContext(cmd *cobra.Command, args []string) error {
	if err := appConfig.PopulateFromFile(); err != nil {
		return err
	}
	if cfg, err := appConfig.UnmarshalOptions(); err != nil {
		return err
	} else {
		if err := initLogs(cfg.Log); err != nil {
			return err
		}
		setAppOpts(cfg)
	}

	return nil
}

// Function `doServe` in the handler function for invocations to the serve subcommand
//
// Returns:
//   - `error`: the issue that arose from executing the serve subcommand run function (nil if execution was ok)
func doServe(cmd *cobra.Command, args []string) error {

	slog.Info("Starting application", slog.String("cmd", cmd.Name()), slog.Any("args", args))
	app, err := NewTournabyteService(GetAppOpts())
	if err != nil {
		return err
	}

	app.With("POST", "/v1/auth/register", utils.ErrorRecovery(), user.CreateUserHandler(app.db))

	return app.Run()
}

// Function `Execute` acts of the CLI entry point for the Tournabyte API webserver
func Execute() {

	if appErr := rootCmd.Execute(); appErr != nil {
		slog.Error("Exit status FAILURE", slog.String("error", appErr.Error()))
	}
	slog.Info("Exit status OK")
}
