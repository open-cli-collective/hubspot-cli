package calls

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// DefaultProperties are the default properties to fetch for calls
var DefaultProperties = []string{"hs_call_body", "hs_call_direction", "hs_call_duration", "hs_call_status", "hs_timestamp", "hubspot_owner_id"}

// Register registers the calls command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "calls",
		Short: "Manage HubSpot calls",
		Long:  "Commands for listing, viewing, creating, updating, and deleting calls (engagement activities) in HubSpot CRM.",
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
		Short: "List calls",
		Long:  "List calls from HubSpot CRM with pagination support.",
		Example: `  # List first 10 calls
  hspt calls list

  # List with pagination
  hspt calls list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = DefaultProperties
			}

			result, err := client.ListObjects(api.ObjectTypeCalls, api.ListOptions{
				Limit:      limit,
				After:      after,
				Properties: properties,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No calls found")
				return nil
			}

			headers := []string{"ID", "DIRECTION", "DURATION", "STATUS", "TIMESTAMP"}
			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, []string{
					obj.ID,
					obj.GetProperty("hs_call_direction"),
					obj.GetProperty("hs_call_duration"),
					obj.GetProperty("hs_call_status"),
					obj.GetProperty("hs_timestamp"),
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of calls to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var properties []string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a call by ID",
		Long:  "Retrieve a single call by its ID from HubSpot CRM.",
		Example: `  # Get call by ID
  hspt calls get 12345`,
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

			obj, err := client.GetObject(api.ObjectTypeCalls, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Call %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Body", obj.GetProperty("hs_call_body")},
				{"Direction", obj.GetProperty("hs_call_direction")},
				{"Duration", obj.GetProperty("hs_call_duration")},
				{"Status", obj.GetProperty("hs_call_status")},
				{"Timestamp", obj.GetProperty("hs_timestamp")},
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
	var body, direction, duration, status, timestamp, ownerID string
	var props []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new call",
		Long:  "Create a new call record in HubSpot CRM.",
		Example: `  # Create a call
  hspt calls create --body "Discussion about project" --direction INBOUND --duration 300

  # Create with status
  hspt calls create --body "Sales call" --direction OUTBOUND --status COMPLETED`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if body != "" {
				properties["hs_call_body"] = body
			}
			if direction != "" {
				properties["hs_call_direction"] = direction
			}
			if duration != "" {
				properties["hs_call_duration"] = duration
			}
			if status != "" {
				properties["hs_call_status"] = status
			}
			if timestamp != "" {
				properties["hs_timestamp"] = timestamp
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

			obj, err := client.CreateObject(api.ObjectTypeCalls, properties)
			if err != nil {
				return err
			}

			v.Success("Call created with ID: %s", obj.ID)

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Direction", obj.GetProperty("hs_call_direction")},
				{"Duration", obj.GetProperty("hs_call_duration")},
			}

			return v.Render(headers, rows, obj)
		},
	}

	cmd.Flags().StringVar(&body, "body", "", "Call notes/body")
	cmd.Flags().StringVar(&direction, "direction", "", "Call direction (INBOUND, OUTBOUND)")
	cmd.Flags().StringVar(&duration, "duration", "", "Call duration in seconds")
	cmd.Flags().StringVar(&status, "status", "", "Call status (COMPLETED, BUSY, NO_ANSWER, etc.)")
	cmd.Flags().StringVar(&timestamp, "timestamp", "", "Call timestamp (Unix milliseconds)")
	cmd.Flags().StringVar(&ownerID, "owner-id", "", "HubSpot owner ID")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newUpdateCmd(opts *root.Options) *cobra.Command {
	var body, direction, duration, status, timestamp, ownerID string
	var props []string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a call",
		Long:  "Update an existing call in HubSpot CRM.",
		Example: `  # Update call notes
  hspt calls update 12345 --body "Updated call notes"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if body != "" {
				properties["hs_call_body"] = body
			}
			if direction != "" {
				properties["hs_call_direction"] = direction
			}
			if duration != "" {
				properties["hs_call_duration"] = duration
			}
			if status != "" {
				properties["hs_call_status"] = status
			}
			if timestamp != "" {
				properties["hs_timestamp"] = timestamp
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

			obj, err := client.UpdateObject(api.ObjectTypeCalls, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Call %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Call %s updated", obj.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&body, "body", "", "Call notes/body")
	cmd.Flags().StringVar(&direction, "direction", "", "Call direction (INBOUND, OUTBOUND)")
	cmd.Flags().StringVar(&duration, "duration", "", "Call duration in seconds")
	cmd.Flags().StringVar(&status, "status", "", "Call status (COMPLETED, BUSY, NO_ANSWER, etc.)")
	cmd.Flags().StringVar(&timestamp, "timestamp", "", "Call timestamp (Unix milliseconds)")
	cmd.Flags().StringVar(&ownerID, "owner-id", "", "HubSpot owner ID")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a call",
		Long:  "Archive (soft delete) a call in HubSpot CRM.",
		Example: `  # Delete call
  hspt calls delete 12345

  # Delete without confirmation
  hspt calls delete 12345 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if !force {
				v.Warning("This will archive call %s. Use --force to confirm.", id)
				return nil
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if err := client.DeleteObject(api.ObjectTypeCalls, id); err != nil {
				if api.IsNotFound(err) {
					v.Error("Call %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Call %s archived", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion without prompt")

	return cmd
}
