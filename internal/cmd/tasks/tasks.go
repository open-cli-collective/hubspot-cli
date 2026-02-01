package tasks

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// DefaultProperties are the default properties to fetch for tasks
var DefaultProperties = []string{"hs_task_subject", "hs_task_body", "hs_task_status", "hs_task_priority", "hs_timestamp", "hubspot_owner_id"}

// Register registers the tasks command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Manage HubSpot tasks",
		Long:  "Commands for listing, viewing, creating, updating, and deleting tasks (engagement activities) in HubSpot CRM.",
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
		Short: "List tasks",
		Long:  "List tasks from HubSpot CRM with pagination support.",
		Example: `  # List first 10 tasks
  hspt tasks list

  # List with pagination
  hspt tasks list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = DefaultProperties
			}

			result, err := client.ListObjects(api.ObjectTypeTasks, api.ListOptions{
				Limit:      limit,
				After:      after,
				Properties: properties,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No tasks found")
				return nil
			}

			headers := []string{"ID", "SUBJECT", "STATUS", "PRIORITY", "TIMESTAMP"}
			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, []string{
					obj.ID,
					truncate(obj.GetProperty("hs_task_subject"), 40),
					obj.GetProperty("hs_task_status"),
					obj.GetProperty("hs_task_priority"),
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of tasks to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var properties []string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a task by ID",
		Long:  "Retrieve a single task by its ID from HubSpot CRM.",
		Example: `  # Get task by ID
  hspt tasks get 12345`,
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

			obj, err := client.GetObject(api.ObjectTypeTasks, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Task %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Subject", obj.GetProperty("hs_task_subject")},
				{"Body", truncate(obj.GetProperty("hs_task_body"), 100)},
				{"Status", obj.GetProperty("hs_task_status")},
				{"Priority", obj.GetProperty("hs_task_priority")},
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
	var subject, body, status, priority, timestamp, ownerID string
	var props []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new task",
		Long:  "Create a new task in HubSpot CRM.",
		Example: `  # Create a task
  hspt tasks create --subject "Follow up with client" --status NOT_STARTED --priority HIGH

  # Create with body
  hspt tasks create --subject "Review proposal" --body "Check pricing section"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if subject != "" {
				properties["hs_task_subject"] = subject
			}
			if body != "" {
				properties["hs_task_body"] = body
			}
			if status != "" {
				properties["hs_task_status"] = status
			}
			if priority != "" {
				properties["hs_task_priority"] = priority
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
				return fmt.Errorf("at least one property is required (--subject is recommended)")
			}

			obj, err := client.CreateObject(api.ObjectTypeTasks, properties)
			if err != nil {
				return err
			}

			v.Success("Task created with ID: %s", obj.ID)

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Subject", obj.GetProperty("hs_task_subject")},
				{"Status", obj.GetProperty("hs_task_status")},
				{"Priority", obj.GetProperty("hs_task_priority")},
			}

			return v.Render(headers, rows, obj)
		},
	}

	cmd.Flags().StringVar(&subject, "subject", "", "Task subject")
	cmd.Flags().StringVar(&body, "body", "", "Task body/description")
	cmd.Flags().StringVar(&status, "status", "", "Task status (NOT_STARTED, IN_PROGRESS, COMPLETED, etc.)")
	cmd.Flags().StringVar(&priority, "priority", "", "Task priority (LOW, MEDIUM, HIGH)")
	cmd.Flags().StringVar(&timestamp, "timestamp", "", "Task due date (Unix milliseconds)")
	cmd.Flags().StringVar(&ownerID, "owner-id", "", "HubSpot owner ID")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newUpdateCmd(opts *root.Options) *cobra.Command {
	var subject, body, status, priority, timestamp, ownerID string
	var props []string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a task",
		Long:  "Update an existing task in HubSpot CRM.",
		Example: `  # Update task status
  hspt tasks update 12345 --status COMPLETED

  # Update task priority
  hspt tasks update 12345 --priority HIGH`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if subject != "" {
				properties["hs_task_subject"] = subject
			}
			if body != "" {
				properties["hs_task_body"] = body
			}
			if status != "" {
				properties["hs_task_status"] = status
			}
			if priority != "" {
				properties["hs_task_priority"] = priority
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

			obj, err := client.UpdateObject(api.ObjectTypeTasks, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Task %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Task %s updated", obj.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&subject, "subject", "", "Task subject")
	cmd.Flags().StringVar(&body, "body", "", "Task body/description")
	cmd.Flags().StringVar(&status, "status", "", "Task status (NOT_STARTED, IN_PROGRESS, COMPLETED, etc.)")
	cmd.Flags().StringVar(&priority, "priority", "", "Task priority (LOW, MEDIUM, HIGH)")
	cmd.Flags().StringVar(&timestamp, "timestamp", "", "Task due date (Unix milliseconds)")
	cmd.Flags().StringVar(&ownerID, "owner-id", "", "HubSpot owner ID")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a task",
		Long:  "Archive (soft delete) a task in HubSpot CRM.",
		Example: `  # Delete task
  hspt tasks delete 12345

  # Delete without confirmation
  hspt tasks delete 12345 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if !force {
				v.Warning("This will archive task %s. Use --force to confirm.", id)
				return nil
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if err := client.DeleteObject(api.ObjectTypeTasks, id); err != nil {
				if api.IsNotFound(err) {
					v.Error("Task %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Task %s archived", id)
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
