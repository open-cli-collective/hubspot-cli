package graphql

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// tableRenderer is the interface for rendering tabular output
type tableRenderer interface {
	Render([]string, [][]string, interface{}) error
	Info(string, ...interface{})
	Error(string, ...interface{})
}

// Register registers the graphql command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "graphql",
		Short: "Execute GraphQL queries",
		Long:  "Commands for executing GraphQL queries against HubSpot's unified API.",
	}

	cmd.AddCommand(newQueryCmd(opts))
	cmd.AddCommand(newExploreCmd(opts))

	parent.AddCommand(cmd)
}

func newQueryCmd(opts *root.Options) *cobra.Command {
	var queryFile string
	var queryString string

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Execute a GraphQL query",
		Long: `Execute a GraphQL query against HubSpot's unified API.

The query can be provided via --file (path to a .graphql file) or
--query (inline query string). If both are provided, --file takes precedence.`,
		Example: `  # Execute query from file
  hspt graphql query --file query.graphql

  # Execute inline query
  hspt graphql query --query '{ CRM { contact_collection(limit: 10) { items { email } } } }'

  # Query a specific contact
  hspt graphql query --query '{ CRM { contact(id: "123") { firstname lastname email } } }'

  # Query with associations
  hspt graphql query --query '{
    CRM {
      contact_collection(limit: 5) {
        items {
          firstname
          associations {
            company_collection {
              items { name }
            }
          }
        }
      }
    }
  }'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			// Get query from file or string
			var query string
			if queryFile != "" {
				data, err := os.ReadFile(queryFile)
				if err != nil {
					return fmt.Errorf("failed to read query file: %w", err)
				}
				query = string(data)
			} else if queryString != "" {
				query = queryString
			} else {
				return fmt.Errorf("either --file or --query is required")
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ExecuteGraphQL(query, nil)
			if err != nil {
				return err
			}

			// Check for GraphQL errors
			if result.HasErrors() {
				v.Error("GraphQL errors: %s", result.ErrorMessages())
				// Still output the response if there's partial data
				if result.Data == nil {
					return nil
				}
			}

			// Format and output the data
			if result.Data != nil {
				// Pretty print the JSON
				var prettyJSON json.RawMessage
				if err := json.Unmarshal(result.Data, &prettyJSON); err != nil {
					return fmt.Errorf("failed to parse response data: %w", err)
				}

				formatted, err := json.MarshalIndent(prettyJSON, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format response: %w", err)
				}

				fmt.Println(string(formatted))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&queryFile, "file", "f", "", "Path to a GraphQL query file")
	cmd.Flags().StringVarP(&queryString, "query", "q", "", "Inline GraphQL query string")

	return cmd
}

func newExploreCmd(opts *root.Options) *cobra.Command {
	var typeName string
	var fieldName string

	cmd := &cobra.Command{
		Use:   "explore",
		Short: "Explore the GraphQL schema",
		Long: `Explore HubSpot's GraphQL schema using introspection.

Without flags, lists all available types. Use --type to show fields of a
specific type, and --field to show details of a specific field.`,
		Example: `  # List all available types
  hspt graphql explore

  # Show fields of the CRM type
  hspt graphql explore --type CRM

  # Show details of a specific field
  hspt graphql explore --type CRM --field contact_collection

  # Explore Contact fields
  hspt graphql explore --type CRM_contact`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			schema, err := client.IntrospectSchema()
			if err != nil {
				return err
			}

			// If no type specified, list root types
			if typeName == "" {
				return listRootTypes(v, schema)
			}

			// Get the specified type
			t := schema.GetType(typeName)
			if t == nil {
				v.Error("Type %q not found in schema", typeName)
				return nil
			}

			// If no field specified, show type fields
			if fieldName == "" {
				return showTypeFields(v, t)
			}

			// Show specific field details
			return showFieldDetails(v, t, fieldName)
		},
	}

	cmd.Flags().StringVarP(&typeName, "type", "t", "", "Type name to explore")
	cmd.Flags().StringVarP(&fieldName, "field", "f", "", "Field name to show details (requires --type)")

	return cmd
}

func listRootTypes(v tableRenderer, schema *api.IntrospectionSchema) error {
	types := schema.GetRootTypes()
	if len(types) == 0 {
		v.Info("No types found in schema")
		return nil
	}

	headers := []string{"TYPE", "KIND", "FIELDS", "DESCRIPTION"}
	rows := make([][]string, 0, len(types))

	for _, t := range types {
		desc := t.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		rows = append(rows, []string{
			t.Name,
			t.Kind,
			fmt.Sprintf("%d", len(t.Fields)),
			desc,
		})
	}

	return v.Render(headers, rows, types)
}

func showTypeFields(v tableRenderer, t *api.IntrospectionType) error {
	if len(t.Fields) == 0 {
		v.Info("Type %s has no fields", t.Name)
		return nil
	}

	headers := []string{"FIELD", "TYPE", "ARGS", "DESCRIPTION"}
	rows := make([][]string, 0, len(t.Fields))

	for _, f := range t.Fields {
		desc := f.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		argsStr := ""
		if len(f.Args) > 0 {
			argsStr = fmt.Sprintf("%d", len(f.Args))
		}
		rows = append(rows, []string{
			f.Name,
			f.Type.TypeName(),
			argsStr,
			desc,
		})
	}

	return v.Render(headers, rows, t.Fields)
}

func showFieldDetails(v tableRenderer, t *api.IntrospectionType, fieldName string) error {
	var field *api.IntrospectionField
	for i := range t.Fields {
		if t.Fields[i].Name == fieldName {
			field = &t.Fields[i]
			break
		}
	}

	if field == nil {
		v.Error("Field %q not found in type %s", fieldName, t.Name)
		return nil
	}

	headers := []string{"PROPERTY", "VALUE"}
	rows := [][]string{
		{"Name", field.Name},
		{"Type", field.Type.TypeName()},
	}

	if field.Description != "" {
		rows = append(rows, []string{"Description", field.Description})
	}
	if field.IsDeprecated {
		rows = append(rows, []string{"Deprecated", "Yes"})
		if field.DeprecationReason != "" {
			rows = append(rows, []string{"Deprecation Reason", field.DeprecationReason})
		}
	}

	if len(field.Args) > 0 {
		rows = append(rows, []string{"", ""})
		rows = append(rows, []string{"ARGUMENTS", ""})
		for _, arg := range field.Args {
			argDesc := arg.Type.TypeName()
			if arg.DefaultValue != nil {
				argDesc += fmt.Sprintf(" = %s", *arg.DefaultValue)
			}
			rows = append(rows, []string{"  " + arg.Name, argDesc})
		}
	}

	return v.Render(headers, rows, field)
}
