package workflows

import (
	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the workflows command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "workflows",
		Short: "Manage HubSpot workflows",
		Long:  "Commands for listing and viewing automation workflows.",
	}

	cmd.AddCommand(newListCmd(opts))
	cmd.AddCommand(newGetCmd(opts))

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
					formatBool(workflow.Enabled),
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
				{"Enabled", formatBool(workflow.Enabled)},
				{"Object Type", formatObjectType(workflow.ObjectTypeID)},
				{"Revision ID", workflow.RevisionID},
				{"Created", workflow.CreatedAt},
				{"Updated", workflow.UpdatedAt},
			}

			return v.Render(headers, rows, workflow)
		},
	}
}

func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
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
