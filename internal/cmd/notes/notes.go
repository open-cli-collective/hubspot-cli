package notes

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// DefaultProperties are the default properties to fetch for notes
var DefaultProperties = []string{"hs_note_body", "hs_timestamp", "hubspot_owner_id"}

// Register registers the notes command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "notes",
		Short: "Manage HubSpot notes",
		Long:  "Commands for listing, viewing, creating, updating, and deleting notes (engagement activities) in HubSpot CRM.",
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
		Short: "List notes",
		Long:  "List notes from HubSpot CRM with pagination support.",
		Example: `  # List first 10 notes
  hspt notes list

  # List with pagination
  hspt notes list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = DefaultProperties
			}

			result, err := client.ListObjects(api.ObjectTypeNotes, api.ListOptions{
				Limit:      limit,
				After:      after,
				Properties: properties,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No notes found")
				return nil
			}

			headers := []string{"ID", "BODY", "TIMESTAMP", "OWNER ID"}
			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, []string{
					obj.ID,
					truncate(obj.GetProperty("hs_note_body"), 50),
					obj.GetProperty("hs_timestamp"),
					obj.GetProperty("hubspot_owner_id"),
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of notes to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var properties []string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a note by ID",
		Long:  "Retrieve a single note by its ID from HubSpot CRM.",
		Example: `  # Get note by ID
  hspt notes get 12345`,
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

			obj, err := client.GetObject(api.ObjectTypeNotes, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Note %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Body", obj.GetProperty("hs_note_body")},
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
	var body, timestamp, ownerID string
	var props []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new note",
		Long:  "Create a new note in HubSpot CRM.",
		Example: `  # Create a note
  hspt notes create --body "Meeting notes from today's call"

  # Create with owner
  hspt notes create --body "Follow-up required" --owner-id 12345`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if body != "" {
				properties["hs_note_body"] = body
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
				return fmt.Errorf("at least one property is required (--body is recommended)")
			}

			obj, err := client.CreateObject(api.ObjectTypeNotes, properties)
			if err != nil {
				return err
			}

			v.Success("Note created with ID: %s", obj.ID)

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Body", truncate(obj.GetProperty("hs_note_body"), 80)},
			}

			return v.Render(headers, rows, obj)
		},
	}

	cmd.Flags().StringVar(&body, "body", "", "Note body content")
	cmd.Flags().StringVar(&timestamp, "timestamp", "", "Note timestamp (Unix milliseconds)")
	cmd.Flags().StringVar(&ownerID, "owner-id", "", "HubSpot owner ID")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newUpdateCmd(opts *root.Options) *cobra.Command {
	var body, timestamp, ownerID string
	var props []string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a note",
		Long:  "Update an existing note in HubSpot CRM.",
		Example: `  # Update note body
  hspt notes update 12345 --body "Updated meeting notes"`,
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
				properties["hs_note_body"] = body
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

			obj, err := client.UpdateObject(api.ObjectTypeNotes, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Note %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Note %s updated", obj.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&body, "body", "", "Note body content")
	cmd.Flags().StringVar(&timestamp, "timestamp", "", "Note timestamp (Unix milliseconds)")
	cmd.Flags().StringVar(&ownerID, "owner-id", "", "HubSpot owner ID")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a note",
		Long:  "Archive (soft delete) a note in HubSpot CRM.",
		Example: `  # Delete note
  hspt notes delete 12345

  # Delete without confirmation
  hspt notes delete 12345 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if !force {
				v.Warning("This will archive note %s. Use --force to confirm.", id)
				return nil
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if err := client.DeleteObject(api.ObjectTypeNotes, id); err != nil {
				if api.IsNotFound(err) {
					v.Error("Note %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Note %s archived", id)
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
