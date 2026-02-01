package products

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// DefaultProperties are the default properties to fetch for products
var DefaultProperties = []string{"name", "price", "description", "hs_sku"}

// Register registers the products command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "products",
		Short: "Manage HubSpot products",
		Long:  "Commands for listing, viewing, creating, updating, and deleting products in HubSpot CRM.",
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
		Short: "List products",
		Long:  "List products from HubSpot CRM with pagination support.",
		Example: `  # List first 10 products
  hspt products list

  # List with custom properties
  hspt products list --properties name,price,hs_sku

  # List with pagination
  hspt products list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = DefaultProperties
			}

			result, err := client.ListObjects(api.ObjectTypeProducts, api.ListOptions{
				Limit:      limit,
				After:      after,
				Properties: properties,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No products found")
				return nil
			}

			headers := []string{"ID", "NAME", "PRICE", "SKU", "DESCRIPTION"}
			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, []string{
					obj.ID,
					obj.GetProperty("name"),
					obj.GetProperty("price"),
					obj.GetProperty("hs_sku"),
					truncate(obj.GetProperty("description"), 40),
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of products to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var properties []string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a product by ID",
		Long:  "Retrieve a single product by its ID from HubSpot CRM.",
		Example: `  # Get product by ID
  hspt products get 12345

  # Get with specific properties
  hspt products get 12345 --properties name,price,hs_sku`,
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

			obj, err := client.GetObject(api.ObjectTypeProducts, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Product %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Name", obj.GetProperty("name")},
				{"Price", obj.GetProperty("price")},
				{"SKU", obj.GetProperty("hs_sku")},
				{"Description", obj.GetProperty("description")},
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
	var name, price, description, sku string
	var props []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new product",
		Long:  "Create a new product in HubSpot CRM.",
		Example: `  # Create with common fields
  hspt products create --name "Widget" --price 99.99 --sku ABC123

  # Create with custom properties
  hspt products create --name "Widget" --prop hs_cost_of_goods_sold=50`,
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
			if price != "" {
				properties["price"] = price
			}
			if description != "" {
				properties["description"] = description
			}
			if sku != "" {
				properties["hs_sku"] = sku
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

			obj, err := client.CreateObject(api.ObjectTypeProducts, properties)
			if err != nil {
				return err
			}

			v.Success("Product created with ID: %s", obj.ID)

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Name", obj.GetProperty("name")},
				{"Price", obj.GetProperty("price")},
				{"SKU", obj.GetProperty("hs_sku")},
			}

			return v.Render(headers, rows, obj)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Product name")
	cmd.Flags().StringVar(&price, "price", "", "Product price")
	cmd.Flags().StringVar(&description, "description", "", "Product description")
	cmd.Flags().StringVar(&sku, "sku", "", "Product SKU")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newUpdateCmd(opts *root.Options) *cobra.Command {
	var name, price, description, sku string
	var props []string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a product",
		Long:  "Update an existing product in HubSpot CRM.",
		Example: `  # Update product price
  hspt products update 12345 --price 149.99

  # Update custom property
  hspt products update 12345 --prop hs_cost_of_goods_sold=75`,
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
			if price != "" {
				properties["price"] = price
			}
			if description != "" {
				properties["description"] = description
			}
			if sku != "" {
				properties["hs_sku"] = sku
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

			obj, err := client.UpdateObject(api.ObjectTypeProducts, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Product %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Product %s updated", obj.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Product name")
	cmd.Flags().StringVar(&price, "price", "", "Product price")
	cmd.Flags().StringVar(&description, "description", "", "Product description")
	cmd.Flags().StringVar(&sku, "sku", "", "Product SKU")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a product",
		Long:  "Archive (soft delete) a product in HubSpot CRM.",
		Example: `  # Delete product
  hspt products delete 12345

  # Delete without confirmation
  hspt products delete 12345 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if !force {
				v.Warning("This will archive product %s. Use --force to confirm.", id)
				return nil
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if err := client.DeleteObject(api.ObjectTypeProducts, id); err != nil {
				if api.IsNotFound(err) {
					v.Error("Product %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Product %s archived", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion without prompt")

	return cmd
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
