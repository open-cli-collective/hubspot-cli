package deals

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// DefaultProperties are the default properties to fetch for deals
var DefaultProperties = []string{"dealname", "amount", "dealstage", "pipeline", "closedate", "hubspot_owner_id"}

// Register registers the deals command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "deals",
		Short: "Manage HubSpot deals",
		Long:  "Commands for listing, viewing, creating, updating, and searching deals in HubSpot CRM.",
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
		Short: "List deals",
		Long:  "List deals from HubSpot CRM with pagination support.",
		Example: `  # List first 10 deals
  hspt deals list

  # List with custom properties
  hspt deals list --properties dealname,amount,dealstage

  # List with pagination
  hspt deals list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = DefaultProperties
			}

			result, err := client.ListObjects(api.ObjectTypeDeals, api.ListOptions{
				Limit:      limit,
				After:      after,
				Properties: properties,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No deals found")
				return nil
			}

			headers := []string{"ID", "NAME", "AMOUNT", "STAGE", "PIPELINE", "CLOSE DATE"}
			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, []string{
					obj.ID,
					obj.GetProperty("dealname"),
					obj.GetProperty("amount"),
					obj.GetProperty("dealstage"),
					obj.GetProperty("pipeline"),
					obj.GetProperty("closedate"),
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of deals to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var properties []string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a deal by ID",
		Long:  "Retrieve a single deal by its ID from HubSpot CRM.",
		Example: `  # Get deal by ID
  hspt deals get 12345`,
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

			obj, err := client.GetObject(api.ObjectTypeDeals, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Deal %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Name", obj.GetProperty("dealname")},
				{"Amount", obj.GetProperty("amount")},
				{"Stage", obj.GetProperty("dealstage")},
				{"Pipeline", obj.GetProperty("pipeline")},
				{"Close Date", obj.GetProperty("closedate")},
				{"Owner ID", obj.GetProperty("hubspot_owner_id")},
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
	var dealname, amount, dealstage, pipeline, closedate, ownerID string
	var props []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new deal",
		Long:  "Create a new deal in HubSpot CRM.",
		Example: `  # Create with common fields
  hspt deals create --name "New Enterprise Deal" --amount 50000 --stage qualifiedtobuy

  # Create with pipeline and close date
  hspt deals create --name "Q1 Deal" --pipeline default --closedate 2024-03-31`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if dealname != "" {
				properties["dealname"] = dealname
			}
			if amount != "" {
				properties["amount"] = amount
			}
			if dealstage != "" {
				properties["dealstage"] = dealstage
			}
			if pipeline != "" {
				properties["pipeline"] = pipeline
			}
			if closedate != "" {
				properties["closedate"] = closedate
			}
			if ownerID != "" {
				properties["hubspot_owner_id"] = ownerID
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

			obj, err := client.CreateObject(api.ObjectTypeDeals, properties)
			if err != nil {
				return err
			}

			v.Success("Deal created with ID: %s", obj.ID)

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Name", obj.GetProperty("dealname")},
				{"Amount", obj.GetProperty("amount")},
				{"Stage", obj.GetProperty("dealstage")},
			}

			return v.Render(headers, rows, obj)
		},
	}

	cmd.Flags().StringVar(&dealname, "name", "", "Deal name")
	cmd.Flags().StringVar(&amount, "amount", "", "Deal amount")
	cmd.Flags().StringVar(&dealstage, "stage", "", "Deal stage")
	cmd.Flags().StringVar(&pipeline, "pipeline", "", "Pipeline ID")
	cmd.Flags().StringVar(&closedate, "closedate", "", "Close date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&ownerID, "owner", "", "Owner ID")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newUpdateCmd(opts *root.Options) *cobra.Command {
	var dealname, amount, dealstage, pipeline, closedate, ownerID string
	var props []string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a deal",
		Long:  "Update an existing deal in HubSpot CRM.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if dealname != "" {
				properties["dealname"] = dealname
			}
			if amount != "" {
				properties["amount"] = amount
			}
			if dealstage != "" {
				properties["dealstage"] = dealstage
			}
			if pipeline != "" {
				properties["pipeline"] = pipeline
			}
			if closedate != "" {
				properties["closedate"] = closedate
			}
			if ownerID != "" {
				properties["hubspot_owner_id"] = ownerID
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

			obj, err := client.UpdateObject(api.ObjectTypeDeals, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Deal %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Deal %s updated", obj.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&dealname, "name", "", "Deal name")
	cmd.Flags().StringVar(&amount, "amount", "", "Deal amount")
	cmd.Flags().StringVar(&dealstage, "stage", "", "Deal stage")
	cmd.Flags().StringVar(&pipeline, "pipeline", "", "Pipeline ID")
	cmd.Flags().StringVar(&closedate, "closedate", "", "Close date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&ownerID, "owner", "", "Owner ID")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a deal",
		Long:  "Archive (soft delete) a deal in HubSpot CRM.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if !force {
				v.Warning("This will archive deal %s. Use --force to confirm.", id)
				return nil
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if err := client.DeleteObject(api.ObjectTypeDeals, id); err != nil {
				if api.IsNotFound(err) {
					v.Error("Deal %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Deal %s archived", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion without prompt")

	return cmd
}

func newSearchCmd(opts *root.Options) *cobra.Command {
	var dealname, dealstage, pipeline string
	var limit int
	var properties []string

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search deals",
		Long:  "Search for deals using filters.",
		Example: `  # Search by name
  hspt deals search --name "Enterprise"

  # Search by stage
  hspt deals search --stage closedwon`,
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

			if dealname != "" {
				filters = append(filters, api.SearchFilter{
					PropertyName: "dealname",
					Operator:     "CONTAINS_TOKEN",
					Value:        dealname,
				})
			}
			if dealstage != "" {
				filters = append(filters, api.SearchFilter{
					PropertyName: "dealstage",
					Operator:     "EQ",
					Value:        dealstage,
				})
			}
			if pipeline != "" {
				filters = append(filters, api.SearchFilter{
					PropertyName: "pipeline",
					Operator:     "EQ",
					Value:        pipeline,
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

			result, err := client.SearchObjects(api.ObjectTypeDeals, req)
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No deals found matching criteria")
				return nil
			}

			headers := []string{"ID", "NAME", "AMOUNT", "STAGE", "PIPELINE", "CLOSE DATE"}
			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, []string{
					obj.ID,
					obj.GetProperty("dealname"),
					obj.GetProperty("amount"),
					obj.GetProperty("dealstage"),
					obj.GetProperty("pipeline"),
					obj.GetProperty("closedate"),
				})
			}

			v.Info("Found %d deal(s)", len(result.Results))
			return v.Render(headers, rows, result)
		},
	}

	cmd.Flags().StringVar(&dealname, "name", "", "Search by name (contains)")
	cmd.Flags().StringVar(&dealstage, "stage", "", "Search by exact deal stage")
	cmd.Flags().StringVar(&pipeline, "pipeline", "", "Search by pipeline ID")
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of results")
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}
