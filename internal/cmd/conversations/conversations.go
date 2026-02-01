package conversations

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the conversations command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "conversations",
		Short: "Manage HubSpot conversations",
		Long:  "Commands for listing and viewing conversations inboxes, threads, channels, and messages.",
	}

	cmd.AddCommand(newInboxesCmd(opts))
	cmd.AddCommand(newThreadsCmd(opts))
	cmd.AddCommand(newChannelsCmd(opts))
	cmd.AddCommand(newMessagesCmd(opts))

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

func newChannelsCmd(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channels",
		Short: "Manage channels",
		Long:  "Commands for listing and viewing conversation channels.",
	}

	cmd.AddCommand(newChannelsListCmd(opts))
	cmd.AddCommand(newChannelsGetCmd(opts))

	return cmd
}

func newChannelsListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List channels",
		Long:  "List conversation channels from HubSpot.",
		Example: `  # List channels
  hspt conversations channels list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListChannels(api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No channels found")
				return nil
			}

			headers := []string{"ID", "NAME", "TYPE", "CREATED"}
			rows := make([][]string, 0, len(result.Results))
			for _, channel := range result.Results {
				rows = append(rows, []string{
					channel.ID,
					channel.Name,
					channel.Type,
					channel.CreatedAt,
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of channels to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func newChannelsGetCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a channel by ID",
		Long:  "Retrieve a single channel by its ID.",
		Example: `  # Get channel by ID
  hspt conversations channels get 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			channel, err := client.GetChannel(id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Channel %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", channel.ID},
				{"Name", channel.Name},
				{"Type", channel.Type},
				{"Account ID", channel.AccountID},
				{"Created", channel.CreatedAt},
				{"Updated", channel.UpdatedAt},
			}

			return v.Render(headers, rows, channel)
		},
	}
}

func newMessagesCmd(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "messages",
		Short: "Manage messages",
		Long:  "Commands for listing and sending conversation messages.",
	}

	cmd.AddCommand(newMessagesListCmd(opts))
	cmd.AddCommand(newMessagesSendCmd(opts))

	return cmd
}

func newMessagesListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "list <threadId>",
		Short: "List messages in a thread",
		Long:  "List messages for a conversation thread.",
		Example: `  # List messages in a thread
  hspt conversations messages list 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			threadID := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListMessages(threadID, api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Thread %s not found", threadID)
					return nil
				}
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No messages found")
				return nil
			}

			headers := []string{"ID", "TYPE", "DIRECTION", "TEXT", "CREATED"}
			rows := make([][]string, 0, len(result.Results))
			for _, msg := range result.Results {
				text := msg.Text
				if text == "" {
					text = msg.RichText
				}
				rows = append(rows, []string{
					msg.ID,
					msg.Type,
					msg.Direction,
					truncate(text, 50),
					msg.CreatedAt,
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of messages to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func newMessagesSendCmd(opts *root.Options) *cobra.Command {
	var text, channelID, senderID string

	cmd := &cobra.Command{
		Use:   "send <threadId>",
		Short: "Send a message to a thread",
		Long:  "Send a message to a conversation thread.",
		Example: `  # Send a message
  hspt conversations messages send 12345 --text "Hello, how can I help?"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			threadID := args[0]

			if text == "" {
				return fmt.Errorf("--text is required")
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			req := api.SendMessageRequest{
				Type: "MESSAGE",
				Text: text,
			}
			if channelID != "" {
				req.ChannelID = channelID
			}
			if senderID != "" {
				req.SenderID = senderID
			}

			msg, err := client.SendMessage(threadID, req)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Thread %s not found", threadID)
					return nil
				}
				return err
			}

			v.Success("Message sent with ID: %s", msg.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&text, "text", "", "Message text content")
	cmd.Flags().StringVar(&channelID, "channel-id", "", "Channel ID (optional)")
	cmd.Flags().StringVar(&senderID, "sender-id", "", "Sender actor ID (optional)")

	return cmd
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
