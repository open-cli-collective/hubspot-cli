package conversations

import (
	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the conversations command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "conversations",
		Short: "Manage HubSpot conversations",
		Long:  "Commands for listing and viewing conversations inboxes and threads.",
	}

	cmd.AddCommand(newInboxesCmd(opts))
	cmd.AddCommand(newThreadsCmd(opts))

	parent.AddCommand(cmd)
}

func newInboxesCmd(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inboxes",
		Short: "Manage inboxes",
		Long:  "Commands for listing and viewing conversation inboxes.",
	}

	cmd.AddCommand(newInboxesListCmd(opts))
	cmd.AddCommand(newInboxesGetCmd(opts))

	return cmd
}

func newInboxesListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List inboxes",
		Long:  "List conversation inboxes from HubSpot.",
		Example: `  # List inboxes
  hspt conversations inboxes list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListInboxes(api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No inboxes found")
				return nil
			}

			headers := []string{"ID", "NAME", "ARCHIVED", "CREATED"}
			rows := make([][]string, 0, len(result.Results))
			for _, inbox := range result.Results {
				rows = append(rows, []string{
					inbox.ID,
					inbox.Name,
					formatBool(inbox.Archived),
					inbox.CreatedAt,
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of inboxes to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func newInboxesGetCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get an inbox by ID",
		Long:  "Retrieve a single inbox by its ID.",
		Example: `  # Get inbox by ID
  hspt conversations inboxes get 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			inbox, err := client.GetInbox(id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Inbox %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", inbox.ID},
				{"Name", inbox.Name},
				{"Archived", formatBool(inbox.Archived)},
				{"Created", inbox.CreatedAt},
				{"Updated", inbox.UpdatedAt},
			}

			return v.Render(headers, rows, inbox)
		},
	}
}

func newThreadsCmd(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "threads",
		Short: "Manage threads",
		Long:  "Commands for listing and viewing conversation threads.",
	}

	cmd.AddCommand(newThreadsListCmd(opts))
	cmd.AddCommand(newThreadsGetCmd(opts))

	return cmd
}

func newThreadsListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List threads",
		Long:  "List conversation threads from HubSpot.",
		Example: `  # List threads
  hspt conversations threads list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListThreads(api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No threads found")
				return nil
			}

			headers := []string{"ID", "STATUS", "INBOX ID", "CONTACT ID", "CREATED"}
			rows := make([][]string, 0, len(result.Results))
			for _, thread := range result.Results {
				rows = append(rows, []string{
					thread.ID,
					thread.Status,
					thread.InboxID,
					thread.AssociatedContactID,
					thread.CreatedAt,
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of threads to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func newThreadsGetCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a thread by ID",
		Long:  "Retrieve a single thread by its ID.",
		Example: `  # Get thread by ID
  hspt conversations threads get 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			thread, err := client.GetThread(id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Thread %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", thread.ID},
				{"Status", thread.Status},
				{"Inbox ID", thread.InboxID},
				{"Contact ID", thread.AssociatedContactID},
				{"Archived", formatBool(thread.Archived)},
				{"Created", thread.CreatedAt},
				{"Updated", thread.UpdatedAt},
			}

			if thread.ClosedAt != "" {
				rows = append(rows, []string{"Closed", thread.ClosedAt})
			}

			return v.Render(headers, rows, thread)
		},
	}
}

func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
