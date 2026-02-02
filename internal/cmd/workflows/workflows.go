package workflows

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/shared"
)

// Register registers the workflows command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "workflows",
		Short: "Manage HubSpot workflows",
		Long:  "Commands for listing, viewing, creating, updating, and deleting automation workflows, plus enrollment operations.",
	}

	cmd.AddCommand(newListCmd(opts))
	cmd.AddCommand(newGetCmd(opts))
	cmd.AddCommand(newCreateCmd(opts))
	cmd.AddCommand(newUpdateCmd(opts))
	cmd.AddCommand(newDeleteCmd(opts))
	cmd.AddCommand(newEnrollCmd(opts))
	cmd.AddCommand(newEnrollmentsCmd(opts))

	parent.AddCommand(cmd)
}

func newListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workflows",
		Long:  "List automation workflows from HubSpot.",
		Example: `  # List workflows
  hspt workflows list

  # List with pagination
  hspt workflows list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListWorkflows(api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No workflows found")
				return nil
			}

			headers := []string{"ID", "NAME", "TYPE", "ENABLED", "OBJECT TYPE"}
			rows := make([][]string, 0, len(result.Results))
			for _, workflow := range result.Results {
				rows = append(rows, []string{
					workflow.ID,
					workflow.Name,
					workflow.Type,
					shared.FormatBool(workflow.Enabled),
					formatObjectType(workflow.ObjectTypeID),
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of workflows to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a workflow by ID",
		Long:  "Retrieve a single workflow by its ID.",
		Example: `  # Get workflow by ID
  hspt workflows get 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			workflow, err := client.GetWorkflow(id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Workflow %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", workflow.ID},
				{"Name", workflow.Name},
				{"Type", workflow.Type},
				{"Enabled", shared.FormatBool(workflow.Enabled)},
				{"Object Type", formatObjectType(workflow.ObjectTypeID)},
				{"Revision ID", workflow.RevisionID},
				{"Created", workflow.CreatedAt},
				{"Updated", workflow.UpdatedAt},
			}

			return v.Render(headers, rows, workflow)
		},
	}
}

func formatObjectType(objectTypeID string) string {
	// HubSpot standard object type IDs
	objectTypes := map[string]string{
		"0-1":  "Contacts",
		"0-2":  "Companies",
		"0-3":  "Deals",
		"0-5":  "Tickets",
		"0-7":  "Products",
		"0-8":  "Line Items",
		"0-14": "Quotes",
	}

	if name, ok := objectTypes[objectTypeID]; ok {
		return name
	}
	if objectTypeID == "" {
		return "-"
	}
	return objectTypeID
}

func newCreateCmd(opts *root.Options) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a workflow",
		Long:  "Create a new automation workflow from a JSON file.",
		Example: `  # Create a workflow from JSON file
  hspt workflows create --file workflow.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if file == "" {
				return fmt.Errorf("--file is required")
			}

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var workflowData map[string]interface{}
			if err := json.Unmarshal(data, &workflowData); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			workflow, err := client.CreateWorkflow(workflowData)
			if err != nil {
				return err
			}

			v.Success("Workflow created: %s (ID: %s)", workflow.Name, workflow.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "JSON file containing workflow definition (required)")

	return cmd
}

func newUpdateCmd(opts *root.Options) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a workflow",
		Long:  "Update an existing automation workflow from a JSON file.",
		Example: `  # Update a workflow from JSON file
  hspt workflows update 12345 --file workflow.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if file == "" {
				return fmt.Errorf("--file is required")
			}

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var workflowData map[string]interface{}
			if err := json.Unmarshal(data, &workflowData); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			workflow, err := client.UpdateWorkflow(id, workflowData)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Workflow %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Workflow updated: %s (ID: %s)", workflow.Name, workflow.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "JSON file containing workflow updates (required)")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a workflow",
		Long:  "Delete an automation workflow by ID.",
		Example: `  # Delete a workflow
  hspt workflows delete 12345

  # Delete without confirmation
  hspt workflows delete 12345 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if !force {
				v.Warning("This will delete workflow %s. Use --force to confirm.", id)
				return nil
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			err = client.DeleteWorkflow(id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Workflow %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Workflow %s deleted", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion without prompt")

	return cmd
}

func newEnrollCmd(opts *root.Options) *cobra.Command {
	var objectID string

	cmd := &cobra.Command{
		Use:   "enroll <workflowId>",
		Short: "Enroll an object in a workflow",
		Long:  "Enroll a contact, company, deal, or other object in a workflow.",
		Example: `  # Enroll a contact in a workflow
  hspt workflows enroll 12345 --object-id 67890`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			workflowID := args[0]

			if objectID == "" {
				return fmt.Errorf("--object-id is required")
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			err = client.EnrollInWorkflow(workflowID, objectID)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Workflow %s not found", workflowID)
					return nil
				}
				return err
			}

			v.Success("Object %s enrolled in workflow %s", objectID, workflowID)
			return nil
		},
	}

	cmd.Flags().StringVar(&objectID, "object-id", "", "ID of the object to enroll (required)")

	return cmd
}

func newEnrollmentsCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "enrollments <workflowId>",
		Short: "List workflow enrollments",
		Long:  "List objects enrolled in a workflow.",
		Example: `  # List enrollments for a workflow
  hspt workflows enrollments 12345

  # List with pagination
  hspt workflows enrollments 12345 --limit 50`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			workflowID := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListWorkflowEnrollments(workflowID, api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Workflow %s not found", workflowID)
					return nil
				}
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No enrollments found for workflow %s", workflowID)
				return nil
			}

			headers := []string{"OBJECT ID", "OBJECT TYPE", "STATUS", "ENROLLED AT"}
			rows := make([][]string, 0, len(result.Results))
			for _, enrollment := range result.Results {
				rows = append(rows, []string{
					enrollment.ObjectID,
					enrollment.ObjectType,
					enrollment.Status,
					enrollment.EnrolledAt,
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of enrollments to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}
