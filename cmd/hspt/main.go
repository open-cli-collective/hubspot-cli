package main

import (
	"fmt"
	"os"

	"github.com/open-cli-collective/hubspot-cli/internal/cmd/campaigns"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/companies"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/completion"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/configcmd"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/contacts"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/deals"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/forms"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/initcmd"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/owners"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/tickets"
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

	// CRM commands
	contacts.Register(rootCmd, opts)
	companies.Register(rootCmd, opts)
	deals.Register(rootCmd, opts)
	tickets.Register(rootCmd, opts)
	owners.Register(rootCmd, opts)

	// Marketing commands
	forms.Register(rootCmd, opts)
	campaigns.Register(rootCmd, opts)

	return rootCmd.Execute()
}
