package hubdb

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the hubdb command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "hubdb",
		Short: "Manage HubDB tables and rows",
		Long:  "Commands for managing HubDB tables and rows. HubDB uses a draft/publish workflow.",
	}

	cmd.AddCommand(newTablesCmd(opts))
	cmd.AddCommand(newRowsCmd(opts))

	parent.AddCommand(cmd)
}

func newTablesCmd(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tables",
		Short: "Manage HubDB tables",
		Long:  "Commands for listing, viewing, creating, publishing, and deleting HubDB tables.",
	}

	cmd.AddCommand(newTablesListCmd(opts))
	cmd.AddCommand(newTablesGetCmd(opts))
	cmd.AddCommand(newTablesCreateCmd(opts))
	cmd.AddCommand(newTablesDeleteCmd(opts))
	cmd.AddCommand(newTablesPublishCmd(opts))

	return cmd
}

func newTablesListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List HubDB tables",
		Long:  "List HubDB tables from HubSpot CMS.",
		Example: `  # List HubDB tables
  hspt hubdb tables list

  # List with pagination
  hspt hubdb tables list --limit 20`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListHubDBTables(api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No HubDB tables found")
				return nil
			}

			headers := []string{"ID", "NAME", "LABEL", "ROWS", "PUBLISHED"}
			rows := make([][]string, 0, len(result.Results))
			for _, table := range result.Results {
				rows = append(rows, []string{
					table.ID,
					table.Name,
					table.Label,
					fmt.Sprintf("%d", table.RowCount),
					formatBool(table.Published),
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of tables to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func newTablesGetCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <tableIdOrName>",
		Short: "Get a HubDB table",
		Long:  "Retrieve a single HubDB table by ID or name.",
		Example: `  # Get table by ID
  hspt hubdb tables get 12345

  # Get table by name
  hspt hubdb tables get my_table`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			tableIDOrName := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			table, err := client.GetHubDBTable(tableIDOrName)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("HubDB table %s not found", tableIDOrName)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", table.ID},
				{"Name", table.Name},
				{"Label", table.Label},
				{"Row Count", fmt.Sprintf("%d", table.RowCount)},
				{"Published", formatBool(table.Published)},
				{"Public API Access", formatBool(table.AllowPublicAPIAccess)},
				{"Created", table.CreatedAt},
				{"Updated", table.UpdatedAt},
			}

			if table.PublishedAt != "" {
				rows = append(rows, []string{"Published At", table.PublishedAt})
			}

			if len(table.Columns) > 0 {
				rows = append(rows, []string{"Columns", fmt.Sprintf("%d", len(table.Columns))})
			}

			return v.Render(headers, rows, table)
		},
	}
}

func newTablesCreateCmd(opts *root.Options) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new HubDB table",
		Long:  "Create a new HubDB table from a JSON file.",
		Example: `  # Create a HubDB table from JSON file
  hspt hubdb tables create --file table.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if file == "" {
				return fmt.Errorf("--file is required")
			}

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var tableData map[string]interface{}
			if err := json.Unmarshal(data, &tableData); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			table, err := client.CreateHubDBTable(tableData)
			if err != nil {
				return err
			}

			v.Success("HubDB table created with ID: %s (name: %s)", table.ID, table.Name)
			v.Info("Note: Table is in draft mode. Use 'hspt hubdb tables publish %s' to publish.", table.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "JSON file containing table definition (required)")

	return cmd
}

func newTablesDeleteCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <tableIdOrName>",
		Short: "Delete a HubDB table",
		Long:  "Delete a HubDB table by ID or name.",
		Example: `  # Delete table by ID
  hspt hubdb tables delete 12345

  # Delete table by name
  hspt hubdb tables delete my_table`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			tableIDOrName := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			err = client.DeleteHubDBTable(tableIDOrName)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("HubDB table %s not found", tableIDOrName)
					return nil
				}
				return err
			}

			v.Success("HubDB table %s deleted", tableIDOrName)
			return nil
		},
	}
}

func newTablesPublishCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "publish <tableIdOrName>",
		Short: "Publish a HubDB table",
		Long:  "Publish a HubDB table draft to make changes live.",
		Example: `  # Publish table by ID
  hspt hubdb tables publish 12345

  # Publish table by name
  hspt hubdb tables publish my_table`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			tableIDOrName := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			table, err := client.PublishHubDBTable(tableIDOrName)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("HubDB table %s not found", tableIDOrName)
					return nil
				}
				return err
			}

			v.Success("HubDB table %s published", table.ID)
			return nil
		},
	}
}

func newRowsCmd(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rows",
		Short: "Manage HubDB table rows",
		Long:  "Commands for listing, viewing, creating, updating, and deleting HubDB table rows.",
	}

	cmd.AddCommand(newRowsListCmd(opts))
	cmd.AddCommand(newRowsGetCmd(opts))
	cmd.AddCommand(newRowsCreateCmd(opts))
	cmd.AddCommand(newRowsUpdateCmd(opts))
	cmd.AddCommand(newRowsDeleteCmd(opts))

	return cmd
}

func newRowsListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "list <tableIdOrName>",
		Short: "List rows in a HubDB table",
		Long:  "List rows from a HubDB table.",
		Example: `  # List rows in a table
  hspt hubdb rows list my_table

  # List with pagination
  hspt hubdb rows list my_table --limit 50`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			tableIDOrName := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListHubDBRows(tableIDOrName, api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("HubDB table %s not found", tableIDOrName)
					return nil
				}
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No rows found in table %s", tableIDOrName)
				return nil
			}

			headers := []string{"ID", "PATH", "NAME", "UPDATED"}
			rows := make([][]string, 0, len(result.Results))
			for _, row := range result.Results {
				rows = append(rows, []string{
					row.ID,
					row.Path,
					row.Name,
					row.UpdatedAt,
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of rows to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func newRowsGetCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <tableIdOrName> <rowId>",
		Short: "Get a row from a HubDB table",
		Long:  "Retrieve a single row from a HubDB table.",
		Example: `  # Get row by ID
  hspt hubdb rows get my_table 12345`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			tableIDOrName := args[0]
			rowID := args[1]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			row, err := client.GetHubDBRow(tableIDOrName, rowID)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Row %s not found in table %s", rowID, tableIDOrName)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", row.ID},
				{"Path", row.Path},
				{"Name", row.Name},
				{"Created", row.CreatedAt},
				{"Updated", row.UpdatedAt},
			}

			// Add row values
			for key, val := range row.Values {
				valStr := fmt.Sprintf("%v", val)
				rows = append(rows, []string{key, truncate(valStr, 50)})
			}

			return v.Render(headers, rows, row)
		},
	}
}

func newRowsCreateCmd(opts *root.Options) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create <tableIdOrName>",
		Short: "Create a row in a HubDB table",
		Long:  "Create a new row in a HubDB table draft. Remember to publish the table to make changes live.",
		Example: `  # Create a row from JSON file
  hspt hubdb rows create my_table --file row.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			tableIDOrName := args[0]

			if file == "" {
				return fmt.Errorf("--file is required")
			}

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var rowData map[string]interface{}
			if err := json.Unmarshal(data, &rowData); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			row, err := client.CreateHubDBRow(tableIDOrName, rowData)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("HubDB table %s not found", tableIDOrName)
					return nil
				}
				return err
			}

			v.Success("Row created with ID: %s", row.ID)
			v.Info("Note: Row is in draft mode. Use 'hspt hubdb tables publish %s' to publish.", tableIDOrName)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "JSON file containing row data (required)")

	return cmd
}

func newRowsUpdateCmd(opts *root.Options) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "update <tableIdOrName> <rowId>",
		Short: "Update a row in a HubDB table",
		Long:  "Update a row in a HubDB table draft. Remember to publish the table to make changes live.",
		Example: `  # Update a row from JSON file
  hspt hubdb rows update my_table 12345 --file updates.json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			tableIDOrName := args[0]
			rowID := args[1]

			if file == "" {
				return fmt.Errorf("--file is required")
			}

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var updates map[string]interface{}
			if err := json.Unmarshal(data, &updates); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			row, err := client.UpdateHubDBRow(tableIDOrName, rowID, updates)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Row %s not found in table %s", rowID, tableIDOrName)
					return nil
				}
				return err
			}

			v.Success("Row %s updated", row.ID)
			v.Info("Note: Changes are in draft mode. Use 'hspt hubdb tables publish %s' to publish.", tableIDOrName)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "JSON file containing row updates (required)")

	return cmd
}

func newRowsDeleteCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <tableIdOrName> <rowId>",
		Short: "Delete a row from a HubDB table",
		Long:  "Delete a row from a HubDB table draft. Remember to publish the table to make changes live.",
		Example: `  # Delete a row
  hspt hubdb rows delete my_table 12345`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			tableIDOrName := args[0]
			rowID := args[1]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			err = client.DeleteHubDBRow(tableIDOrName, rowID)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Row %s not found in table %s", rowID, tableIDOrName)
					return nil
				}
				return err
			}

			v.Success("Row %s deleted from draft", rowID)
			v.Info("Note: Deletion is in draft mode. Use 'hspt hubdb tables publish %s' to publish.", tableIDOrName)
			return nil
		},
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
