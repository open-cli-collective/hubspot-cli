package main

import (
	"fmt"
	"os"

	"github.com/open-cli-collective/hubspot-cli/internal/cmd/completion"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/configcmd"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/initcmd"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
	"github.com/open-cli-collective/hubspot-cli/internal/exitcode"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitcode.GeneralError)
	}
}

func run() error {
	rootCmd, opts := root.NewCmd()

	// Register all commands
	initcmd.Register(rootCmd, opts)
	configcmd.Register(rootCmd, opts)
	completion.Register(rootCmd, opts)

	return rootCmd.Execute()
}
