package tickets

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// DefaultProperties are the default properties to fetch for tickets
var DefaultProperties = []string{"subject", "content", "hs_pipeline", "hs_pipeline_stage", "hs_ticket_priority", "hubspot_owner_id"}

// Register registers the tickets command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "tickets",
		Short: "Manage HubSpot tickets",
		Long:  "Commands for listing, viewing, creating, updating, and searching tickets in HubSpot CRM.",
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
		Short: "List tickets",
		Long:  "List tickets from HubSpot CRM with pagination support.",
		Example: `  # List first 10 tickets
  hspt tickets list

  # List with custom properties
  hspt tickets list --properties subject,hs_pipeline_stage,hs_ticket_priority

  # List with pagination
  hspt tickets list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = DefaultProperties
			}

			result, err := client.ListObjects(api.ObjectTypeTickets, api.ListOptions{
				Limit:      limit,
				After:      after,
				Properties: properties,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No tickets found")
				return nil
			}

			headers := []string{"ID", "SUBJECT", "STAGE", "PRIORITY", "PIPELINE"}
			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, []string{
					obj.ID,
					obj.GetProperty("subject"),
					obj.GetProperty("hs_pipeline_stage"),
					obj.GetProperty("hs_ticket_priority"),
					obj.GetProperty("hs_pipeline"),
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of tickets to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var properties []string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a ticket by ID",
		Long:  "Retrieve a single ticket by its ID from HubSpot CRM.",
		Example: `  # Get ticket by ID
  hspt tickets get 12345`,
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

			obj, err := client.GetObject(api.ObjectTypeTickets, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Ticket %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Subject", obj.GetProperty("subject")},
				{"Content", obj.GetProperty("content")},
				{"Pipeline", obj.GetProperty("hs_pipeline")},
				{"Stage", obj.GetProperty("hs_pipeline_stage")},
				{"Priority", obj.GetProperty("hs_ticket_priority")},
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
	var subject, content, pipeline, stage, priority, ownerID string
	var props []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new ticket",
		Long:  "Create a new ticket in HubSpot CRM.",
		Example: `  # Create with common fields
  hspt tickets create --subject "Login issue" --priority HIGH

  # Create with content and pipeline
  hspt tickets create --subject "Bug report" --content "Details here..." --pipeline 0`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if subject != "" {
				properties["subject"] = subject
			}
			if content != "" {
				properties["content"] = content
			}
			if pipeline != "" {
				properties["hs_pipeline"] = pipeline
			}
			if stage != "" {
				properties["hs_pipeline_stage"] = stage
			}
			if priority != "" {
				properties["hs_ticket_priority"] = priority
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

			obj, err := client.CreateObject(api.ObjectTypeTickets, properties)
			if err != nil {
				return err
			}

			v.Success("Ticket created with ID: %s", obj.ID)

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Subject", obj.GetProperty("subject")},
				{"Stage", obj.GetProperty("hs_pipeline_stage")},
				{"Priority", obj.GetProperty("hs_ticket_priority")},
			}

			return v.Render(headers, rows, obj)
		},
	}

	cmd.Flags().StringVar(&subject, "subject", "", "Ticket subject")
	cmd.Flags().StringVar(&content, "content", "", "Ticket content/description")
	cmd.Flags().StringVar(&pipeline, "pipeline", "", "Pipeline ID")
	cmd.Flags().StringVar(&stage, "stage", "", "Pipeline stage")
	cmd.Flags().StringVar(&priority, "priority", "", "Priority (LOW, MEDIUM, HIGH)")
	cmd.Flags().StringVar(&ownerID, "owner", "", "Owner ID")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newUpdateCmd(opts *root.Options) *cobra.Command {
	var subject, content, pipeline, stage, priority, ownerID string
	var props []string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a ticket",
		Long:  "Update an existing ticket in HubSpot CRM.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if subject != "" {
				properties["subject"] = subject
			}
			if content != "" {
				properties["content"] = content
			}
			if pipeline != "" {
				properties["hs_pipeline"] = pipeline
			}
			if stage != "" {
				properties["hs_pipeline_stage"] = stage
			}
			if priority != "" {
				properties["hs_ticket_priority"] = priority
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

			obj, err := client.UpdateObject(api.ObjectTypeTickets, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Ticket %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Ticket %s updated", obj.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&subject, "subject", "", "Ticket subject")
	cmd.Flags().StringVar(&content, "content", "", "Ticket content/description")
	cmd.Flags().StringVar(&pipeline, "pipeline", "", "Pipeline ID")
	cmd.Flags().StringVar(&stage, "stage", "", "Pipeline stage")
	cmd.Flags().StringVar(&priority, "priority", "", "Priority (LOW, MEDIUM, HIGH)")
	cmd.Flags().StringVar(&ownerID, "owner", "", "Owner ID")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a ticket",
		Long:  "Archive (soft delete) a ticket in HubSpot CRM.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if !force {
				v.Warning("This will archive ticket %s. Use --force to confirm.", id)
				return nil
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if err := client.DeleteObject(api.ObjectTypeTickets, id); err != nil {
				if api.IsNotFound(err) {
					v.Error("Ticket %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Ticket %s archived", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion without prompt")

	return cmd
}

func newSearchCmd(opts *root.Options) *cobra.Command {
	var subject, stage, priority, pipeline string
	var limit int
	var properties []string

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search tickets",
		Long:  "Search for tickets using filters.",
		Example: `  # Search by subject
  hspt tickets search --subject "login"

  # Search by priority
  hspt tickets search --priority HIGH`,
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

			if subject != "" {
				filters = append(filters, api.SearchFilter{
					PropertyName: "subject",
					Operator:     "CONTAINS_TOKEN",
					Value:        subject,
				})
			}
			if stage != "" {
				filters = append(filters, api.SearchFilter{
					PropertyName: "hs_pipeline_stage",
					Operator:     "EQ",
					Value:        stage,
				})
			}
			if priority != "" {
				filters = append(filters, api.SearchFilter{
					PropertyName: "hs_ticket_priority",
					Operator:     "EQ",
					Value:        priority,
				})
			}
			if pipeline != "" {
				filters = append(filters, api.SearchFilter{
					PropertyName: "hs_pipeline",
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

			result, err := client.SearchObjects(api.ObjectTypeTickets, req)
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No tickets found matching criteria")
				return nil
			}

			headers := []string{"ID", "SUBJECT", "STAGE", "PRIORITY", "PIPELINE"}
			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, []string{
					obj.ID,
					obj.GetProperty("subject"),
					obj.GetProperty("hs_pipeline_stage"),
					obj.GetProperty("hs_ticket_priority"),
					obj.GetProperty("hs_pipeline"),
				})
			}

			v.Info("Found %d ticket(s)", len(result.Results))
			return v.Render(headers, rows, result)
		},
	}

	cmd.Flags().StringVar(&subject, "subject", "", "Search by subject (contains)")
	cmd.Flags().StringVar(&stage, "stage", "", "Search by pipeline stage")
	cmd.Flags().StringVar(&priority, "priority", "", "Search by priority (LOW, MEDIUM, HIGH)")
	cmd.Flags().StringVar(&pipeline, "pipeline", "", "Search by pipeline ID")
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of results")
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}
