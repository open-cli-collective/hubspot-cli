package companies

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// DefaultProperties are the default properties to fetch for companies
var DefaultProperties = []string{"name", "domain", "industry", "phone", "city", "state", "country"}

// Register registers the companies command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "companies",
		Short: "Manage HubSpot companies",
		Long:  "Commands for listing, viewing, creating, updating, and searching companies in HubSpot CRM.",
	}

	cmd.AddCommand(newListCmd(opts))
	cmd.AddCommand(newGetCmd(opts))
	cmd.AddCommand(newCreateCmd(opts))
	cmd.AddCommand(newUpdateCmd(opts))
	cmd.AddCommand(newDeleteCmd(opts))
	cmd.AddCommand(newSearchCmd(opts))

	parent.AddCommand(cmd)
}

func newListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string
	var properties []string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List companies",
		Long:  "List companies from HubSpot CRM with pagination support.",
		Example: `  # List first 10 companies
  hspt companies list

  # List with custom properties
  hspt companies list --properties name,domain,industry

  # List with pagination
  hspt companies list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = DefaultProperties
			}

			result, err := client.ListObjects(api.ObjectTypeCompanies, api.ListOptions{
				Limit:      limit,
				After:      after,
				Properties: properties,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No companies found")
				return nil
			}

			headers := []string{"ID", "NAME", "DOMAIN", "INDUSTRY", "CITY"}
			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, []string{
					obj.ID,
					obj.GetProperty("name"),
					obj.GetProperty("domain"),
					obj.GetProperty("industry"),
					obj.GetProperty("city"),
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of companies to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var properties []string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a company by ID",
		Long:  "Retrieve a single company by its ID from HubSpot CRM.",
		Example: `  # Get company by ID
  hspt companies get 12345`,
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

			obj, err := client.GetObject(api.ObjectTypeCompanies, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Company %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Name", obj.GetProperty("name")},
				{"Domain", obj.GetProperty("domain")},
				{"Industry", obj.GetProperty("industry")},
				{"Phone", obj.GetProperty("phone")},
				{"City", obj.GetProperty("city")},
				{"State", obj.GetProperty("state")},
				{"Country", obj.GetProperty("country")},
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
	var name, domain, industry, phone, city string
	var props []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new company",
		Long:  "Create a new company in HubSpot CRM.",
		Example: `  # Create with common fields
  hspt companies create --name "Acme Inc" --domain acme.com

  # Create with custom properties
  hspt companies create --name "Acme Inc" --prop numberofemployees=100`,
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
			if domain != "" {
				properties["domain"] = domain
			}
			if industry != "" {
				properties["industry"] = industry
			}
			if phone != "" {
				properties["phone"] = phone
			}
			if city != "" {
				properties["city"] = city
			}

			for _, p := range props {
				parts := strings.SplitN(p, "=", 2)
				if len(parts) == 2 {
					properties[parts[0]] = parts[1]
				}
			}

			if len(properties) == 0 {
				return fmt.Errorf("at least one property is required")
			}

			obj, err := client.CreateObject(api.ObjectTypeCompanies, properties)
			if err != nil {
				return err
			}

			v.Success("Company created with ID: %s", obj.ID)

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Name", obj.GetProperty("name")},
				{"Domain", obj.GetProperty("domain")},
			}

			return v.Render(headers, rows, obj)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Company name")
	cmd.Flags().StringVar(&domain, "domain", "", "Company domain")
	cmd.Flags().StringVar(&industry, "industry", "", "Industry")
	cmd.Flags().StringVar(&phone, "phone", "", "Phone number")
	cmd.Flags().StringVar(&city, "city", "", "City")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newUpdateCmd(opts *root.Options) *cobra.Command {
	var name, domain, industry, phone, city string
	var props []string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a company",
		Long:  "Update an existing company in HubSpot CRM.",
		Args:  cobra.ExactArgs(1),
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
			if domain != "" {
				properties["domain"] = domain
			}
			if industry != "" {
				properties["industry"] = industry
			}
			if phone != "" {
				properties["phone"] = phone
			}
			if city != "" {
				properties["city"] = city
			}

			for _, p := range props {
				parts := strings.SplitN(p, "=", 2)
				if len(parts) == 2 {
					properties[parts[0]] = parts[1]
				}
			}

			if len(properties) == 0 {
				return fmt.Errorf("at least one property to update is required")
			}

			obj, err := client.UpdateObject(api.ObjectTypeCompanies, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Company %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Company %s updated", obj.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Company name")
	cmd.Flags().StringVar(&domain, "domain", "", "Company domain")
	cmd.Flags().StringVar(&industry, "industry", "", "Industry")
	cmd.Flags().StringVar(&phone, "phone", "", "Phone number")
	cmd.Flags().StringVar(&city, "city", "", "City")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a company",
		Long:  "Archive (soft delete) a company in HubSpot CRM.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if !force {
				v.Warning("This will archive company %s. Use --force to confirm.", id)
				return nil
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if err := client.DeleteObject(api.ObjectTypeCompanies, id); err != nil {
				if api.IsNotFound(err) {
					v.Error("Company %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Company %s archived", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion without prompt")

	return cmd
}

func newSearchCmd(opts *root.Options) *cobra.Command {
	var name, domain string
	var limit int
	var properties []string

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search companies",
		Long:  "Search for companies using filters.",
		Example: `  # Search by domain
  hspt companies search --domain acme.com

  # Search by name
  hspt companies search --name "Acme"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = DefaultProperties
			}

			var filters []api.SearchFilter

			if name != "" {
				filters = append(filters, api.SearchFilter{
					PropertyName: "name",
					Operator:     "CONTAINS_TOKEN",
					Value:        name,
				})
			}
			if domain != "" {
				filters = append(filters, api.SearchFilter{
					PropertyName: "domain",
					Operator:     "EQ",
					Value:        domain,
				})
			}

			req := api.SearchRequest{
				Properties: properties,
				Limit:      limit,
			}

			if len(filters) > 0 {
				req.FilterGroups = []api.SearchFilterGroup{
					{Filters: filters},
				}
			}

			result, err := client.SearchObjects(api.ObjectTypeCompanies, req)
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No companies found matching criteria")
				return nil
			}

			headers := []string{"ID", "NAME", "DOMAIN", "INDUSTRY", "CITY"}
			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, []string{
					obj.ID,
					obj.GetProperty("name"),
					obj.GetProperty("domain"),
					obj.GetProperty("industry"),
					obj.GetProperty("city"),
				})
			}

			v.Info("Found %d company(ies)", len(result.Results))
			return v.Render(headers, rows, result)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Search by name (contains)")
	cmd.Flags().StringVar(&domain, "domain", "", "Search by exact domain")
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of results")
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}
