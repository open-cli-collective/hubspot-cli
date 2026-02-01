package graphql

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the graphql command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "graphql",
		Short: "Execute GraphQL queries",
		Long:  "Commands for executing GraphQL queries against HubSpot's unified API.",
	}

	cmd.AddCommand(newQueryCmd(opts))

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
