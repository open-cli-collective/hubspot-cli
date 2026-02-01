package main

import (
	"fmt"
	"os"

	"github.com/open-cli-collective/hubspot-cli/internal/cmd/associations"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/calls"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/campaigns"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/companies"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/completion"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/configcmd"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/contacts"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/conversations"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/deals"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/domains"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/emails"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/files"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/forms"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/graphql"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/initcmd"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/lineitems"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/marketingemails"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/meetings"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/notes"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/owners"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/pipelines"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/products"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/properties"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/quotes"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/tasks"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/tickets"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/workflows"
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
	products.Register(rootCmd, opts)
	lineitems.Register(rootCmd, opts)
	quotes.Register(rootCmd, opts)

	// CRM engagement commands
	notes.Register(rootCmd, opts)
	calls.Register(rootCmd, opts)
	emails.Register(rootCmd, opts)
	meetings.Register(rootCmd, opts)
	tasks.Register(rootCmd, opts)

	// CRM infrastructure commands
	associations.Register(rootCmd, opts)
	properties.Register(rootCmd, opts)
	pipelines.Register(rootCmd, opts)

	// Marketing commands
	forms.Register(rootCmd, opts)
	campaigns.Register(rootCmd, opts)
	marketingemails.Register(rootCmd, opts)

	// CMS commands
	files.Register(rootCmd, opts)
	domains.Register(rootCmd, opts)

	// Conversations commands
	conversations.Register(rootCmd, opts)

	// Automation commands
	workflows.Register(rootCmd, opts)

	// GraphQL commands
	graphql.Register(rootCmd, opts)

	return rootCmd.Execute()
}
