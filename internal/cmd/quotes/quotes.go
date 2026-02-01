package quotes

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// DefaultProperties are the default properties to fetch for quotes
var DefaultProperties = []string{"hs_title", "hs_expiration_date", "hs_status", "hs_quote_amount"}

// Register registers the quotes command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "quotes",
		Short: "Manage HubSpot quotes",
		Long:  "Commands for listing, viewing, creating, updating, and deleting quotes in HubSpot CRM.",
	}

	cmd.AddCommand(newListCmd(opts))
	cmd.AddCommand(newGetCmd(opts))
	cmd.AddCommand(newCreateCmd(opts))
	cmd.AddCommand(newUpdateCmd(opts))
	cmd.AddCommand(newDeleteCmd(opts))

	parent.AddCommand(cmd)
}

func newListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string
	var properties []string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List quotes",
		Long:  "List quotes from HubSpot CRM with pagination support.",
		Example: `  # List first 10 quotes
  hspt quotes list

  # List with custom properties
  hspt quotes list --properties hs_title,hs_status,hs_quote_amount

  # List with pagination
  hspt quotes list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = DefaultProperties
			}

			result, err := client.ListObjects(api.ObjectTypeQuotes, api.ListOptions{
				Limit:      limit,
				After:      after,
				Properties: properties,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No quotes found")
				return nil
			}

			headers := []string{"ID", "TITLE", "STATUS", "AMOUNT", "EXPIRES"}
			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, []string{
					obj.ID,
					obj.GetProperty("hs_title"),
					obj.GetProperty("hs_status"),
					obj.GetProperty("hs_quote_amount"),
					obj.GetProperty("hs_expiration_date"),
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of quotes to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var properties []string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a quote by ID",
		Long:  "Retrieve a single quote by its ID from HubSpot CRM.",
		Example: `  # Get quote by ID
  hspt quotes get 12345

  # Get with specific properties
  hspt quotes get 12345 --properties hs_title,hs_status,hs_quote_amount`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = DefaultProperties
			}

			obj, err := client.GetObject(api.ObjectTypeQuotes, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Quote %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Title", obj.GetProperty("hs_title")},
				{"Status", obj.GetProperty("hs_status")},
				{"Amount", obj.GetProperty("hs_quote_amount")},
				{"Expiration Date", obj.GetProperty("hs_expiration_date")},
				{"Created", obj.CreatedAt},
				{"Updated", obj.UpdatedAt},
			}

			return v.Render(headers, rows, obj)
		},
	}

	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}

func newCreateCmd(opts *root.Options) *cobra.Command {
	var title, status, expirationDate string
	var props []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new quote",
		Long:  "Create a new quote in HubSpot CRM.",
		Example: `  # Create with common fields
  hspt quotes create --title "Q1 Proposal" --status DRAFT --expiration-date 2024-12-31

  # Create with custom properties
  hspt quotes create --title "Enterprise Deal" --prop hs_sender_company_name="Acme Corp"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if title != "" {
				properties["hs_title"] = title
			}
			if status != "" {
				properties["hs_status"] = status
			}
			if expirationDate != "" {
				properties["hs_expiration_date"] = expirationDate
			}

			// Parse custom properties
			for _, p := range props {
				parts := strings.SplitN(p, "=", 2)
				if len(parts) == 2 {
					properties[parts[0]] = parts[1]
				}
			}

			if len(properties) == 0 {
				return fmt.Errorf("at least one property is required")
			}

			obj, err := client.CreateObject(api.ObjectTypeQuotes, properties)
			if err != nil {
				return err
			}

			v.Success("Quote created with ID: %s", obj.ID)

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Title", obj.GetProperty("hs_title")},
				{"Status", obj.GetProperty("hs_status")},
				{"Expiration Date", obj.GetProperty("hs_expiration_date")},
			}

			return v.Render(headers, rows, obj)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Quote title")
	cmd.Flags().StringVar(&status, "status", "", "Quote status (DRAFT, APPROVAL_NOT_NEEDED, etc.)")
	cmd.Flags().StringVar(&expirationDate, "expiration-date", "", "Quote expiration date (YYYY-MM-DD)")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newUpdateCmd(opts *root.Options) *cobra.Command {
	var title, status, expirationDate string
	var props []string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a quote",
		Long:  "Update an existing quote in HubSpot CRM.",
		Example: `  # Update quote status
  hspt quotes update 12345 --status APPROVED

  # Update custom property
  hspt quotes update 12345 --prop hs_terms="Net 30"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if title != "" {
				properties["hs_title"] = title
			}
			if status != "" {
				properties["hs_status"] = status
			}
			if expirationDate != "" {
				properties["hs_expiration_date"] = expirationDate
			}

			// Parse custom properties
			for _, p := range props {
				parts := strings.SplitN(p, "=", 2)
				if len(parts) == 2 {
					properties[parts[0]] = parts[1]
				}
			}

			if len(properties) == 0 {
				return fmt.Errorf("at least one property to update is required")
			}

			obj, err := client.UpdateObject(api.ObjectTypeQuotes, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Quote %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Quote %s updated", obj.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Quote title")
	cmd.Flags().StringVar(&status, "status", "", "Quote status (DRAFT, APPROVAL_NOT_NEEDED, etc.)")
	cmd.Flags().StringVar(&expirationDate, "expiration-date", "", "Quote expiration date (YYYY-MM-DD)")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a quote",
		Long:  "Archive (soft delete) a quote in HubSpot CRM.",
		Example: `  # Delete quote
  hspt quotes delete 12345

  # Delete without confirmation
  hspt quotes delete 12345 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if !force {
				v.Warning("This will archive quote %s. Use --force to confirm.", id)
				return nil
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if err := client.DeleteObject(api.ObjectTypeQuotes, id); err != nil {
				if api.IsNotFound(err) {
					v.Error("Quote %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Quote %s archived", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion without prompt")

	return cmd
}
