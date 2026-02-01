package lineitems

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// DefaultProperties are the default properties to fetch for line items
var DefaultProperties = []string{"name", "quantity", "price", "hs_product_id", "amount"}

// Register registers the line-items command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:     "line-items",
		Aliases: []string{"lineitems"},
		Short:   "Manage HubSpot line items",
		Long:    "Commands for listing, viewing, creating, updating, and deleting line items in HubSpot CRM.",
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
		Short: "List line items",
		Long:  "List line items from HubSpot CRM with pagination support.",
		Example: `  # List first 10 line items
  hspt line-items list

  # List with custom properties
  hspt line-items list --properties name,quantity,price

  # List with pagination
  hspt line-items list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = DefaultProperties
			}

			result, err := client.ListObjects(api.ObjectTypeLineItems, api.ListOptions{
				Limit:      limit,
				After:      after,
				Properties: properties,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No line items found")
				return nil
			}

			headers := []string{"ID", "NAME", "QUANTITY", "PRICE", "AMOUNT", "PRODUCT ID"}
			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, []string{
					obj.ID,
					obj.GetProperty("name"),
					obj.GetProperty("quantity"),
					obj.GetProperty("price"),
					obj.GetProperty("amount"),
					obj.GetProperty("hs_product_id"),
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of line items to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var properties []string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a line item by ID",
		Long:  "Retrieve a single line item by its ID from HubSpot CRM.",
		Example: `  # Get line item by ID
  hspt line-items get 12345

  # Get with specific properties
  hspt line-items get 12345 --properties name,quantity,price`,
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

			obj, err := client.GetObject(api.ObjectTypeLineItems, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Line item %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Name", obj.GetProperty("name")},
				{"Quantity", obj.GetProperty("quantity")},
				{"Price", obj.GetProperty("price")},
				{"Amount", obj.GetProperty("amount")},
				{"Product ID", obj.GetProperty("hs_product_id")},
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
	var name, quantity, price, productID string
	var props []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new line item",
		Long:  "Create a new line item in HubSpot CRM.",
		Example: `  # Create with common fields
  hspt line-items create --name "Widget" --quantity 2 --price 99.99

  # Create linked to a product
  hspt line-items create --name "Widget" --quantity 1 --product-id 12345`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if name != "" {
				properties["name"] = name
			}
			if quantity != "" {
				properties["quantity"] = quantity
			}
			if price != "" {
				properties["price"] = price
			}
			if productID != "" {
				properties["hs_product_id"] = productID
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

			obj, err := client.CreateObject(api.ObjectTypeLineItems, properties)
			if err != nil {
				return err
			}

			v.Success("Line item created with ID: %s", obj.ID)

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Name", obj.GetProperty("name")},
				{"Quantity", obj.GetProperty("quantity")},
				{"Price", obj.GetProperty("price")},
			}

			return v.Render(headers, rows, obj)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Line item name")
	cmd.Flags().StringVar(&quantity, "quantity", "", "Quantity")
	cmd.Flags().StringVar(&price, "price", "", "Unit price")
	cmd.Flags().StringVar(&productID, "product-id", "", "Associated product ID")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newUpdateCmd(opts *root.Options) *cobra.Command {
	var name, quantity, price, productID string
	var props []string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a line item",
		Long:  "Update an existing line item in HubSpot CRM.",
		Example: `  # Update quantity
  hspt line-items update 12345 --quantity 5

  # Update custom property
  hspt line-items update 12345 --prop discount=10`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if name != "" {
				properties["name"] = name
			}
			if quantity != "" {
				properties["quantity"] = quantity
			}
			if price != "" {
				properties["price"] = price
			}
			if productID != "" {
				properties["hs_product_id"] = productID
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

			obj, err := client.UpdateObject(api.ObjectTypeLineItems, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Line item %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Line item %s updated", obj.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Line item name")
	cmd.Flags().StringVar(&quantity, "quantity", "", "Quantity")
	cmd.Flags().StringVar(&price, "price", "", "Unit price")
	cmd.Flags().StringVar(&productID, "product-id", "", "Associated product ID")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a line item",
		Long:  "Archive (soft delete) a line item in HubSpot CRM.",
		Example: `  # Delete line item
  hspt line-items delete 12345

  # Delete without confirmation
  hspt line-items delete 12345 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if !force {
				v.Warning("This will archive line item %s. Use --force to confirm.", id)
				return nil
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if err := client.DeleteObject(api.ObjectTypeLineItems, id); err != nil {
				if api.IsNotFound(err) {
					v.Error("Line item %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Line item %s archived", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion without prompt")

	return cmd
}
