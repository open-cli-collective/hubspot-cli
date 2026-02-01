package forms

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the forms command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "forms",
		Short: "Manage HubSpot forms",
		Long:  "Commands for listing and viewing forms and their submissions in HubSpot.",
	}

	cmd.AddCommand(newListCmd(opts))
	cmd.AddCommand(newGetCmd(opts))
	cmd.AddCommand(newSubmissionsCmd(opts))

	parent.AddCommand(cmd)
}

func newListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List forms",
		Long:  "List forms from HubSpot with pagination support.",
		Example: `  # List first 10 forms
  hspt forms list

  # List with pagination
  hspt forms list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListForms(api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No forms found")
				return nil
			}

			headers := []string{"ID", "NAME", "TYPE", "ARCHIVED", "CREATED"}
			rows := make([][]string, 0, len(result.Results))
			for _, form := range result.Results {
				archived := "No"
				if form.Archived {
					archived = "Yes"
				}
				rows = append(rows, []string{
					form.ID,
					form.Name,
					form.FormType,
					archived,
					form.CreatedAt,
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of forms to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a form by ID",
		Long:  "Retrieve a single form by its ID from HubSpot.",
		Example: `  # Get form by ID
  hspt forms get abc123-def456`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			form, err := client.GetForm(id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Form %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", form.ID},
				{"Name", form.Name},
				{"Type", form.FormType},
				{"Archived", formatBool(form.Archived)},
				{"Created", form.CreatedAt},
				{"Updated", form.UpdatedAt},
			}

			// Add field count
			fieldCount := 0
			for _, group := range form.FieldGroups {
				fieldCount += len(group.Fields)
			}
			rows = append(rows, []string{"Fields", fmt.Sprintf("%d", fieldCount)})

			if err := v.Render(headers, rows, form); err != nil {
				return err
			}

			// Show fields in verbose mode
			if opts.Verbose && len(form.FieldGroups) > 0 {
				v.Println("")
				v.Info("Form Fields:")
				fieldHeaders := []string{"NAME", "LABEL", "TYPE", "REQUIRED"}
				fieldRows := make([][]string, 0)
				for _, group := range form.FieldGroups {
					for _, field := range group.Fields {
						fieldRows = append(fieldRows, []string{
							field.Name,
							field.Label,
							field.FieldType,
							formatBool(field.Required),
						})
					}
				}
				if len(fieldRows) > 0 {
					return v.Render(fieldHeaders, fieldRows, nil)
				}
			}

			return nil
		},
	}
}

func newSubmissionsCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "submissions <form-id>",
		Short: "List form submissions",
		Long:  "List submissions for a specific form.",
		Example: `  # List submissions for a form
  hspt forms submissions abc123-def456

  # With pagination
  hspt forms submissions abc123-def456 --limit 50`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			formID := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.GetFormSubmissions(formID, api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Form %s not found", formID)
					return nil
				}
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No submissions found for form %s", formID)
				return nil
			}

			// For submissions, we show a summary since values vary per form
			headers := []string{"ID", "SUBMITTED AT", "FIELD COUNT"}
			rows := make([][]string, 0, len(result.Results))
			for _, sub := range result.Results {
				rows = append(rows, []string{
					sub.ID,
					sub.SubmittedAt,
					fmt.Sprintf("%d", len(sub.Values)),
				})
			}

			v.Info("Found %d submission(s)", len(result.Results))

			if err := v.Render(headers, rows, result); err != nil {
				return err
			}

			if result.Paging != nil && result.Paging.Next != nil {
				v.Info("\nMore results available. Use --after %s to get the next page.", result.Paging.Next.After)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of submissions to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
