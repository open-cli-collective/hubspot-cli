package domains

import (
	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the domains command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "domains",
		Short: "Manage HubSpot domains",
		Long:  "Commands for listing and viewing domains configured in HubSpot.",
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
		Short: "List domains",
		Long:  "List domains configured in HubSpot with pagination support.",
		Example: `  # List all domains
  hspt domains list

  # List with pagination
  hspt domains list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListDomains(api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No domains found")
				return nil
			}

			headers := []string{"ID", "DOMAIN", "SSL", "RESOLVING", "PRIMARY FOR"}
			rows := make([][]string, 0, len(result.Results))
			for _, domain := range result.Results {
				primaryFor := getPrimaryUses(&domain)
				rows = append(rows, []string{
					domain.ID,
					domain.Domain,
					formatBool(domain.IsSslEnabled),
					formatBool(domain.IsResolving),
					primaryFor,
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of domains to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a domain by ID",
		Long:  "Retrieve a single domain by its ID.",
		Example: `  # Get domain by ID
  hspt domains get 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			domain, err := client.GetDomain(id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Domain %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", domain.ID},
				{"Domain", domain.Domain},
				{"SSL Enabled", formatBool(domain.IsSslEnabled)},
				{"SSL Only", formatBool(domain.IsSslOnly)},
				{"Resolving", formatBool(domain.IsResolving)},
				{"Primary Site Page", formatBool(domain.PrimarySitePage)},
				{"Primary Landing Page", formatBool(domain.PrimaryLandingPage)},
				{"Primary Blog Post", formatBool(domain.PrimaryBlogPost)},
				{"Primary Email", formatBool(domain.PrimaryEmail)},
				{"Used for Site Pages", formatBool(domain.IsUsedForSitePage)},
				{"Used for Landing Pages", formatBool(domain.IsUsedForLandingPage)},
				{"Used for Blog Posts", formatBool(domain.IsUsedForBlogPost)},
				{"Used for Email", formatBool(domain.IsUsedForEmail)},
				{"Created", domain.CreatedAt},
				{"Updated", domain.UpdatedAt},
			}

			return v.Render(headers, rows, domain)
		},
	}
}

func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func getPrimaryUses(d *api.Domain) string {
	uses := ""
	if d.PrimarySitePage {
		uses += "Site, "
	}
	if d.PrimaryLandingPage {
		uses += "Landing, "
	}
	if d.PrimaryBlogPost {
		uses += "Blog, "
	}
	if d.PrimaryEmail {
		uses += "Email, "
	}
	if len(uses) > 2 {
		return uses[:len(uses)-2]
	}
	return "-"
}
