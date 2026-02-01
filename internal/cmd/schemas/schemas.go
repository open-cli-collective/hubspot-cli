package schemas

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the schemas command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "schemas",
		Short: "Manage custom object schemas",
		Long:  "Commands for listing, viewing, creating, and deleting custom object schemas. Requires Operations Hub Professional or Enterprise.",
	}

	cmd.AddCommand(newListCmd(opts))
	cmd.AddCommand(newGetCmd(opts))
	cmd.AddCommand(newCreateCmd(opts))
	cmd.AddCommand(newDeleteCmd(opts))

	parent.AddCommand(cmd)
}

func newListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom object schemas",
		Long:  "List all custom object schemas in your HubSpot account.",
		Example: `  # List schemas
  hspt schemas list

  # List with pagination
  hspt schemas list --limit 20`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListSchemas(api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No custom object schemas found")
				return nil
			}

			headers := []string{"NAME", "FULLY QUALIFIED NAME", "LABELS", "PROPERTIES"}
			rows := make([][]string, 0, len(result.Results))
			for _, schema := range result.Results {
				labels := ""
				if schema.Labels.Singular != "" {
					labels = fmt.Sprintf("%s / %s", schema.Labels.Singular, schema.Labels.Plural)
				}
				rows = append(rows, []string{
					schema.Name,
					schema.FullyQualifiedName,
					labels,
					fmt.Sprintf("%d", len(schema.Properties)),
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of schemas to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <fullyQualifiedName>",
		Short: "Get a custom object schema",
		Long:  "Retrieve a single custom object schema by its fully qualified name.",
		Example: `  # Get schema by fully qualified name
  hspt schemas get p_my_custom_object`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			fqn := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			schema, err := client.GetSchema(fqn)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Schema %s not found", fqn)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"Name", schema.Name},
				{"Fully Qualified Name", schema.FullyQualifiedName},
				{"Object Type ID", schema.ObjectTypeID},
				{"Singular Label", schema.Labels.Singular},
				{"Plural Label", schema.Labels.Plural},
				{"Primary Display Property", schema.PrimaryDisplayProperty},
				{"Archived", formatBool(schema.Archived)},
				{"Created", schema.CreatedAt},
				{"Updated", schema.UpdatedAt},
			}

			if len(schema.Properties) > 0 {
				propNames := make([]string, 0, len(schema.Properties))
				for _, prop := range schema.Properties {
					propNames = append(propNames, prop.Name)
				}
				rows = append(rows, []string{"Properties", strings.Join(propNames, ", ")})
			}

			if len(schema.AssociatedObjects) > 0 {
				rows = append(rows, []string{"Associated Objects", strings.Join(schema.AssociatedObjects, ", ")})
			}

			return v.Render(headers, rows, schema)
		},
	}
}

func newCreateCmd(opts *root.Options) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a custom object schema",
		Long:  "Create a new custom object schema from a JSON file. Requires Operations Hub Professional or Enterprise.",
		Example: `  # Create a schema from JSON file
  hspt schemas create --file schema.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if file == "" {
				return fmt.Errorf("--file is required")
			}

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var schemaData map[string]interface{}
			if err := json.Unmarshal(data, &schemaData); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			schema, err := client.CreateSchema(schemaData)
			if err != nil {
				return err
			}

			v.Success("Schema created: %s (fully qualified name: %s)", schema.Name, schema.FullyQualifiedName)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "JSON file containing schema definition (required)")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <fullyQualifiedName>",
		Short: "Delete a custom object schema",
		Long:  "Delete a custom object schema by its fully qualified name. This will also delete all objects of that type.",
		Example: `  # Delete schema by fully qualified name
  hspt schemas delete p_my_custom_object`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			fqn := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			err = client.DeleteSchema(fqn)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Schema %s not found", fqn)
					return nil
				}
				return err
			}

			v.Success("Schema %s deleted", fqn)
			return nil
		},
	}
}

func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
