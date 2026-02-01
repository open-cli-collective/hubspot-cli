package pages

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the pages command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "pages",
		Short: "Manage HubSpot CMS pages",
		Long:  "Commands for listing, viewing, creating, updating, and deleting site and landing pages.",
	}

	cmd.AddCommand(newListCmd(opts))
	cmd.AddCommand(newGetCmd(opts))
	cmd.AddCommand(newCreateCmd(opts))
	cmd.AddCommand(newUpdateCmd(opts))
	cmd.AddCommand(newDeleteCmd(opts))
	cmd.AddCommand(newCloneCmd(opts))

	parent.AddCommand(cmd)
}

func parsePageType(typeStr string) api.PageType {
	if typeStr == "landing" {
		return api.PageTypeLanding
	}
	return api.PageTypeSite
}

func newListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string
	var pageType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pages",
		Long:  "List site or landing pages from HubSpot CMS.",
		Example: `  # List site pages
  hspt pages list

  # List landing pages
  hspt pages list --type landing

  # List with pagination
  hspt pages list --limit 20`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			pt := parsePageType(pageType)
			result, err := client.ListPages(pt, api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No pages found")
				return nil
			}

			headers := []string{"ID", "NAME", "SLUG", "STATE", "UPDATED"}
			rows := make([][]string, 0, len(result.Results))
			for _, page := range result.Results {
				rows = append(rows, []string{
					page.ID,
					truncate(page.Name, 30),
					truncate(page.Slug, 25),
					page.State,
					page.UpdatedAt,
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of pages to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")
	cmd.Flags().StringVar(&pageType, "type", "site", "Page type: site or landing")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var pageType string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a page by ID",
		Long:  "Retrieve a single page by its ID.",
		Example: `  # Get site page by ID
  hspt pages get 12345

  # Get landing page by ID
  hspt pages get 12345 --type landing`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			pt := parsePageType(pageType)
			page, err := client.GetPage(pt, id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Page %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", page.ID},
				{"Name", page.Name},
				{"Slug", page.Slug},
				{"State", page.State},
				{"HTML Title", page.HTMLTitle},
				{"Meta Description", truncate(page.MetaDescription, 50)},
				{"Domain", page.Domain},
				{"Author", page.AuthorName},
				{"Archived", formatBool(page.Archived)},
				{"Created", page.CreatedAt},
				{"Updated", page.UpdatedAt},
			}

			if page.PublishDate != "" {
				rows = append(rows, []string{"Published", page.PublishDate})
			}

			return v.Render(headers, rows, page)
		},
	}

	cmd.Flags().StringVar(&pageType, "type", "site", "Page type: site or landing")

	return cmd
}

func newCreateCmd(opts *root.Options) *cobra.Command {
	var file string
	var pageType string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new page",
		Long:  "Create a new site or landing page in HubSpot CMS.",
		Example: `  # Create a site page from JSON file
  hspt pages create --file page.json

  # Create a landing page from JSON file
  hspt pages create --file page.json --type landing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if file == "" {
				return fmt.Errorf("--file is required")
			}

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var pageData map[string]interface{}
			if err := json.Unmarshal(data, &pageData); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			pt := parsePageType(pageType)
			page, err := client.CreatePage(pt, pageData)
			if err != nil {
				return err
			}

			v.Success("Page created with ID: %s", page.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "JSON file containing page data (required)")
	cmd.Flags().StringVar(&pageType, "type", "site", "Page type: site or landing")

	return cmd
}

func newUpdateCmd(opts *root.Options) *cobra.Command {
	var file string
	var pageType string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a page",
		Long:  "Update an existing page in HubSpot CMS.",
		Example: `  # Update a site page from JSON file
  hspt pages update 12345 --file updates.json

  # Update a landing page
  hspt pages update 12345 --file updates.json --type landing`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if file == "" {
				return fmt.Errorf("--file is required")
			}

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var updates map[string]interface{}
			if err := json.Unmarshal(data, &updates); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			pt := parsePageType(pageType)
			page, err := client.UpdatePage(pt, id, updates)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Page %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Page %s updated", page.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "JSON file containing page updates (required)")
	cmd.Flags().StringVar(&pageType, "type", "site", "Page type: site or landing")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var pageType string

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a page",
		Long:  "Archive/delete a page in HubSpot CMS.",
		Example: `  # Delete a site page
  hspt pages delete 12345

  # Delete a landing page
  hspt pages delete 12345 --type landing`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			pt := parsePageType(pageType)
			err = client.DeletePage(pt, id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Page %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Page %s deleted", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&pageType, "type", "site", "Page type: site or landing")

	return cmd
}

func newCloneCmd(opts *root.Options) *cobra.Command {
	var pageType string

	cmd := &cobra.Command{
		Use:   "clone <id>",
		Short: "Clone a page",
		Long:  "Create a copy of an existing page in HubSpot CMS.",
		Example: `  # Clone a site page
  hspt pages clone 12345

  # Clone a landing page
  hspt pages clone 12345 --type landing`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			pt := parsePageType(pageType)
			page, err := client.ClonePage(pt, id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Page %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Page cloned with new ID: %s", page.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&pageType, "type", "site", "Page type: site or landing")

	return cmd
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
