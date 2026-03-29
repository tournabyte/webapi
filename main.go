package main

/*
 * File: main.go
 *
 * Purpose: main entry point into the Tournabyte API application
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import cli "github.com/tournabyte/webapi/cmd"

// Function `main` is the antry point to the Tournabyte web API application
func main() {
	cli.Execute()
}
