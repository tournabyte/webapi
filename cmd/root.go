package cmd

/*
 * File: cmd/root.go
 *
 * Purpose: define the root command instance for the CLI to the webapi application
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Root command level CLI constants
const (
	configurationFileFlag             = "config"
	configurationFileFlagDefaultValue = "."
	configurationFileFlagHelpMsg      = "Directory to look for the `webapi.json` configuration file."

	rootCmdUsageMsg  = "com.tournabyte.webapi [--config path/to/config] [-h|--help]"
	rootCmdShortHelp = "CLI for controlling the Tournabyte web API process"

	configFileLookupSystem = "/etc/tournabyte"
	configFileLookupLocal  = "$HOME/.local/tournabyte"
	configFileLookupName   = "webapi"
	configFileLookupType   = "json"
)

// Variable `rootCmd` holds a pointer to the root `cobra.Command` struct representing the Tournabyte webapi CLI
var rootCmd *cobra.Command = &cobra.Command{
	Use:               rootCmdUsageMsg,
	Short:             rootCmdShortHelp,
	PersistentPreRunE: initAppConfig,
	RunE:              doRoot,
}

// Function `init` contains the initialization logic to perform on `rootCmd`
func init() {
	rootCmd.AddCommand(serverCmd, validationCmd)
	rootCmd.PersistentFlags().String(
		configurationFileFlag,
		configurationFileFlagDefaultValue,
		configurationFileFlagHelpMsg,
	)
}

// Function `initAppConfig` contains the initialization logic to prepare for application configuration data retrieval
func initAppConfig(cmd *cobra.Command, args []string) error {
	var customConfigDir string
	if f := cmd.Flags().Lookup(configurationFileFlag); f == nil {
		return fmt.Errorf("Command node `%v` did not contain a `--%s` flag", cmd.Name(), configurationFileFlag)
	} else {
		customConfigDir = f.Value.String()
	}

	appConfig = NewAppConfig(
		configFileLookupType,
		configFileLookupName,
		configFileLookupSystem,
		configFileLookupLocal,
		customConfigDir,
	)
	return nil
}

func doRoot(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

// Function `execGreeting` prints the banner message at the start of execution
func execGreeting() {
	log.Printf("Welcome to the Tournabyte Web API!\n\n")
	if info, ok := debug.ReadBuildInfo(); !ok {
		log.Printf("No build information available.\n\n")
	} else {
		log.Printf("Build information:\n%s\n\n", info.String())
	}
}

// Function `execMain` represents the application entry point
//
// Returns:
//   - `int`: a value that acts as the exit code to report to the OS
func execMain() int {
	var (
		EXIT_OK  int = 0
		EXIT_ERR int = 1
	)

	if err := rootCmd.Execute(); err != nil {
		log.Printf("!!! ERROR: %s !!!", err.Error())
		return EXIT_ERR
	}
	return EXIT_OK

}

// Function `Execute` represents the CLI entry point
func Execute() {
	execGreeting()
	os.Exit(execMain())
}
