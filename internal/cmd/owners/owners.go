package owners

import (
	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the owners command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "owners",
		Short: "Manage HubSpot owners",
		Long:  "Commands for listing and viewing owners (users) in HubSpot.",
	}

	cmd.AddCommand(newListCmd(opts))
	cmd.AddCommand(newGetCmd(opts))

	parent.AddCommand(cmd)
}

func newListCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List owners",
		Long:  "List all owners (users) in HubSpot.",
		Example: `  # List all owners
  hspt owners list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			owners, err := client.GetOwners()
			if err != nil {
				return err
			}

			if len(owners) == 0 {
				v.Info("No owners found")
				return nil
			}

			headers := []string{"ID", "EMAIL", "NAME", "ARCHIVED"}
			rows := make([][]string, 0, len(owners))
			for _, owner := range owners {
				archived := "No"
				if owner.Archived {
					archived = "Yes"
				}
				rows = append(rows, []string{
					owner.ID,
					owner.Email,
					owner.FullName(),
					archived,
				})
			}

			return v.Render(headers, rows, owners)
		},
	}
}

func newGetCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get an owner by ID",
		Long:  "Retrieve a single owner by their ID.",
		Example: `  # Get owner by ID
  hspt owners get 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			owner, err := client.GetOwner(id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Owner %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", owner.ID},
				{"Email", owner.Email},
				{"Name", owner.FullName()},
				{"First Name", owner.FirstName},
				{"Last Name", owner.LastName},
				{"Archived", formatBool(owner.Archived)},
			}

			if len(owner.Teams) > 0 {
				for _, team := range owner.Teams {
					primary := ""
					if team.Primary {
						primary = " (primary)"
					}
					rows = append(rows, []string{"Team", team.Name + primary})
				}
			}

			return v.Render(headers, rows, owner)
		},
	}
}

func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
