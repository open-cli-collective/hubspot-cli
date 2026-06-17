package shared

import (
	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// SearchCmdConfig describes the object-specific pieces of a `search` subcommand.
// Everything else (flag wiring, filter/sort parsing, the search request,
// pagination, and rendering) is shared across object types.
type SearchCmdConfig struct {
	// ObjectType is the HubSpot CRM object type to search.
	ObjectType api.ObjectType
	// Noun is the singular human-readable object name (e.g. "task", "email")
	// used in the empty-result and result-count messages.
	Noun string
	// Short and Long are the cobra command descriptions.
	Short string
	Long  string
	// Example is the cobra command example text.
	Example string
	// DefaultProperties are fetched when the user does not pass --properties.
	DefaultProperties []string
	// Headers are the table column headers.
	Headers []string
	// Row maps a result object to a table row. It is called once per result and
	// must return values aligned with Headers.
	Row func(obj api.CRMObject) []string
}

// NewSearchCmd builds a `search` subcommand for a CRM object type. The
// object-specific behavior is supplied via cfg; the command wiring, filter/sort
// parsing (via ParseFilters/ParseSort), request building, pagination, and
// rendering are shared across object types.
func NewSearchCmd(opts *root.Options, cfg SearchCmdConfig) *cobra.Command {
	var filterArgs []string
	var sortArgs []string
	var limit int
	var after string
	var properties []string

	cmd := &cobra.Command{
		Use:     "search",
		Short:   cfg.Short,
		Long:    cfg.Long,
		Example: cfg.Example,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = cfg.DefaultProperties
			}

			filters, err := ParseFilters(filterArgs)
			if err != nil {
				return err
			}

			sorts, err := ParseSort(sortArgs)
			if err != nil {
				return err
			}

			req := api.SearchRequest{
				Properties: properties,
				Limit:      limit,
				After:      after,
				Sorts:      sorts,
			}
			if len(filters) > 0 {
				req.FilterGroups = []api.SearchFilterGroup{
					{Filters: filters},
				}
			}

			result, err := client.SearchObjects(cfg.ObjectType, req)
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No %ss found matching criteria", cfg.Noun)
				return nil
			}

			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, cfg.Row(obj))
			}

			v.Info("Found %d %s(s)", len(result.Results), cfg.Noun)
			if err := v.Render(cfg.Headers, rows, result); err != nil {
				return err
			}

			if result.Paging != nil && result.Paging.Next != nil {
				v.Info("\nMore results available. Use --after %s to get the next page.", result.Paging.Next.After)
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVar(&filterArgs, "filter", nil, "Filter condition (e.g. prop=value, prop>=value, prop:OPERATOR:value); repeatable")
	cmd.Flags().StringArrayVar(&sortArgs, "sort", nil, "Sort condition (e.g. hs_timestamp:asc or hs_timestamp:desc); repeatable")
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of results")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}
