package campaigns

import (
	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the campaigns command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "campaigns",
		Short: "Manage HubSpot marketing campaigns",
		Long:  "Commands for listing and viewing marketing campaigns in HubSpot.",
	}

	cmd.AddCommand(newListCmd(opts))
	cmd.AddCommand(newGetCmd(opts))

	parent.AddCommand(cmd)
}

func newListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List campaigns",
		Long:  "List marketing campaigns from HubSpot with pagination support.",
		Example: `  # List first 10 campaigns
  hspt campaigns list

  # List with pagination
  hspt campaigns list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListCampaigns(api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No campaigns found")
				return nil
			}

			headers := []string{"ID", "NAME", "CREATED", "UPDATED"}
			rows := make([][]string, 0, len(result.Results))
			for _, campaign := range result.Results {
				rows = append(rows, []string{
					campaign.ID,
					campaign.Name,
					campaign.CreatedAt,
					campaign.UpdatedAt,
				})
			}

			if err := v.Render(headers, rows, result); err != nil {
				return err
			}

			if result.Paging != nil && result.Paging.Next != nil {
				v.Info("\nMore results available. Use --after %s to get the next page.", result.Paging.Next.After)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of campaigns to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a campaign by ID",
		Long:  "Retrieve a single campaign by its ID from HubSpot.",
		Example: `  # Get campaign by ID
  hspt campaigns get 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			campaign, err := client.GetCampaign(id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Campaign %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", campaign.ID},
				{"Name", campaign.Name},
				{"Created", campaign.CreatedAt},
				{"Updated", campaign.UpdatedAt},
			}

			return v.Render(headers, rows, campaign)
		},
	}
}
