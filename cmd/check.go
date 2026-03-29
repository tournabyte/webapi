package cmd

/*
 * File: cmd/check.go
 *
 * Purpose: define the check subcommand instance for the CLI to the webapi application
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// Check subcommand level CLI constants
const (
	checkCmdUsageMsg  = "check"
	checkCmdShortHelp = "Validates the Tournabyte web API configuration"

	syntaxCheckFlag             = "syntax"
	syntaxCheckFlagDefaultValue = true
	syntaxCheckFlagHelpMsg      = "Enable/disable syntax checking of the configuration file (can parse as JSON)."

	bindCheckflag         = "load"
	bindCheckDefaultValue = false
	bindCheckflagHelpMsg  = "Enable/disable unmarshalling configuration file to program memory. Implies the --syntax flag."
)

var (
	checkSyntax  *bool
	checkLoading *bool
)

var validationCmd *cobra.Command = &cobra.Command{
	Use:   checkCmdUsageMsg,
	Short: checkCmdShortHelp,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if appConfig == nil {
			return errors.New("Application configuration manager is not present")
		}
		return nil
	},
	RunE: doCheck,
}

func init() {
	checkSyntax = validationCmd.Flags().Bool(
		syntaxCheckFlag,
		syntaxCheckFlagDefaultValue,
		syntaxCheckFlagHelpMsg,
	)

	checkLoading = validationCmd.Flags().Bool(
		bindCheckflag,
		bindCheckDefaultValue,
		bindCheckflagHelpMsg,
	)
}

func doCheck(cmd *cobra.Command, args []string) error {
	if *checkLoading {
		return appConfig.Bind()
	}
	if *checkSyntax {
		return appConfig.populateFromDisk()
	}

	return fmt.Errorf("No check flag was set [--%s|--%s]. No work was done.", syntaxCheckFlag, bindCheckflag)
}
