package meetings

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// DefaultProperties are the default properties to fetch for meetings
var DefaultProperties = []string{"hs_meeting_title", "hs_meeting_body", "hs_meeting_start_time", "hs_meeting_end_time", "hs_meeting_outcome", "hubspot_owner_id"}

// Register registers the meetings command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "meetings",
		Short: "Manage HubSpot meetings",
		Long:  "Commands for listing, viewing, creating, updating, and deleting meetings (engagement activities) in HubSpot CRM.",
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
		Short: "List meetings",
		Long:  "List meetings from HubSpot CRM with pagination support.",
		Example: `  # List first 10 meetings
  hspt meetings list

  # List with pagination
  hspt meetings list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = DefaultProperties
			}

			result, err := client.ListObjects(api.ObjectTypeMeetings, api.ListOptions{
				Limit:      limit,
				After:      after,
				Properties: properties,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No meetings found")
				return nil
			}

			headers := []string{"ID", "TITLE", "START TIME", "END TIME", "OUTCOME"}
			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, []string{
					obj.ID,
					truncate(obj.GetProperty("hs_meeting_title"), 40),
					obj.GetProperty("hs_meeting_start_time"),
					obj.GetProperty("hs_meeting_end_time"),
					obj.GetProperty("hs_meeting_outcome"),
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of meetings to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var properties []string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a meeting by ID",
		Long:  "Retrieve a single meeting by its ID from HubSpot CRM.",
		Example: `  # Get meeting by ID
  hspt meetings get 12345`,
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

			obj, err := client.GetObject(api.ObjectTypeMeetings, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Meeting %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Title", obj.GetProperty("hs_meeting_title")},
				{"Body", truncate(obj.GetProperty("hs_meeting_body"), 100)},
				{"Start Time", obj.GetProperty("hs_meeting_start_time")},
				{"End Time", obj.GetProperty("hs_meeting_end_time")},
				{"Outcome", obj.GetProperty("hs_meeting_outcome")},
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
	var title, body, startTime, endTime, outcome, ownerID string
	var props []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new meeting",
		Long:  "Create a new meeting record in HubSpot CRM.",
		Example: `  # Create a meeting
  hspt meetings create --title "Sales Demo" --start-time 1704067200000 --end-time 1704070800000

  # Create with outcome
  hspt meetings create --title "Discovery Call" --outcome SCHEDULED`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if title != "" {
				properties["hs_meeting_title"] = title
			}
			if body != "" {
				properties["hs_meeting_body"] = body
			}
			if startTime != "" {
				properties["hs_meeting_start_time"] = startTime
			}
			if endTime != "" {
				properties["hs_meeting_end_time"] = endTime
			}
			if outcome != "" {
				properties["hs_meeting_outcome"] = outcome
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

			obj, err := client.CreateObject(api.ObjectTypeMeetings, properties)
			if err != nil {
				return err
			}

			v.Success("Meeting created with ID: %s", obj.ID)

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Title", obj.GetProperty("hs_meeting_title")},
				{"Start Time", obj.GetProperty("hs_meeting_start_time")},
			}

			return v.Render(headers, rows, obj)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Meeting title")
	cmd.Flags().StringVar(&body, "body", "", "Meeting description/notes")
	cmd.Flags().StringVar(&startTime, "start-time", "", "Meeting start time (Unix milliseconds)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Meeting end time (Unix milliseconds)")
	cmd.Flags().StringVar(&outcome, "outcome", "", "Meeting outcome (SCHEDULED, COMPLETED, RESCHEDULED, etc.)")
	cmd.Flags().StringVar(&ownerID, "owner-id", "", "HubSpot owner ID")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newUpdateCmd(opts *root.Options) *cobra.Command {
	var title, body, startTime, endTime, outcome, ownerID string
	var props []string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a meeting",
		Long:  "Update an existing meeting in HubSpot CRM.",
		Example: `  # Update meeting outcome
  hspt meetings update 12345 --outcome COMPLETED`,
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
				properties["hs_meeting_title"] = title
			}
			if body != "" {
				properties["hs_meeting_body"] = body
			}
			if startTime != "" {
				properties["hs_meeting_start_time"] = startTime
			}
			if endTime != "" {
				properties["hs_meeting_end_time"] = endTime
			}
			if outcome != "" {
				properties["hs_meeting_outcome"] = outcome
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

			obj, err := client.UpdateObject(api.ObjectTypeMeetings, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Meeting %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Meeting %s updated", obj.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Meeting title")
	cmd.Flags().StringVar(&body, "body", "", "Meeting description/notes")
	cmd.Flags().StringVar(&startTime, "start-time", "", "Meeting start time (Unix milliseconds)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Meeting end time (Unix milliseconds)")
	cmd.Flags().StringVar(&outcome, "outcome", "", "Meeting outcome (SCHEDULED, COMPLETED, RESCHEDULED, etc.)")
	cmd.Flags().StringVar(&ownerID, "owner-id", "", "HubSpot owner ID")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a meeting",
		Long:  "Archive (soft delete) a meeting in HubSpot CRM.",
		Example: `  # Delete meeting
  hspt meetings delete 12345

  # Delete without confirmation
  hspt meetings delete 12345 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if !force {
				v.Warning("This will archive meeting %s. Use --force to confirm.", id)
				return nil
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if err := client.DeleteObject(api.ObjectTypeMeetings, id); err != nil {
				if api.IsNotFound(err) {
					v.Error("Meeting %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Meeting %s archived", id)
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
