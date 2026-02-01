package marketingemails

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the marketing emails command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:     "marketing-emails",
		Aliases: []string{"emails", "me"},
		Short:   "Manage HubSpot marketing emails",
		Long:    "Commands for listing, viewing, creating, updating, and deleting marketing emails.",
	}

	cmd.AddCommand(newListCmd(opts))
	cmd.AddCommand(newGetCmd(opts))
	cmd.AddCommand(newCreateCmd(opts))
	cmd.AddCommand(newUpdateCmd(opts))
	cmd.AddCommand(newDeleteCmd(opts))

	parent.AddCommand(cmd)
}

func newListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List marketing emails",
		Long:  "List marketing emails from HubSpot.",
		Example: `  # List marketing emails
  hspt marketing-emails list

  # List with pagination
  hspt marketing-emails list --limit 20`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListMarketingEmails(api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No marketing emails found")
				return nil
			}

			headers := []string{"ID", "NAME", "SUBJECT", "STATE", "TYPE"}
			rows := make([][]string, 0, len(result.Results))
			for _, email := range result.Results {
				rows = append(rows, []string{
					email.ID,
					truncate(email.Name, 30),
					truncate(email.Subject, 30),
					email.State,
					email.Type,
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of emails to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a marketing email by ID",
		Long:  "Retrieve a single marketing email by its ID.",
		Example: `  # Get email by ID
  hspt marketing-emails get 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			email, err := client.GetMarketingEmail(id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Marketing email %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", email.ID},
				{"Name", email.Name},
				{"Subject", email.Subject},
				{"Type", email.Type},
				{"State", email.State},
				{"From Name", email.FromName},
				{"Reply To", email.ReplyTo},
				{"Campaign ID", email.CampaignID},
				{"Archived", formatBool(email.Archived)},
				{"Created", email.CreatedAt},
				{"Updated", email.UpdatedAt},
			}

			if email.PublishDate != "" {
				rows = append(rows, []string{"Published", email.PublishDate})
			}

			return v.Render(headers, rows, email)
		},
	}
}

func newCreateCmd(opts *root.Options) *cobra.Command {
	var name, subject, fromName, replyTo, emailType string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new marketing email",
		Long:  "Create a new marketing email in HubSpot.",
		Example: `  # Create a marketing email
  hspt marketing-emails create --name "Welcome Email" --subject "Welcome to our newsletter"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if name == "" {
				return fmt.Errorf("--name is required")
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			data := map[string]interface{}{
				"name": name,
			}
			if subject != "" {
				data["subject"] = subject
			}
			if fromName != "" {
				data["fromName"] = fromName
			}
			if replyTo != "" {
				data["replyTo"] = replyTo
			}
			if emailType != "" {
				data["type"] = emailType
			}

			email, err := client.CreateMarketingEmail(data)
			if err != nil {
				return err
			}

			v.Success("Marketing email created with ID: %s", email.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Email name (required)")
	cmd.Flags().StringVar(&subject, "subject", "", "Email subject line")
	cmd.Flags().StringVar(&fromName, "from-name", "", "Sender display name")
	cmd.Flags().StringVar(&replyTo, "reply-to", "", "Reply-to email address")
	cmd.Flags().StringVar(&emailType, "type", "", "Email type (e.g., REGULAR, AUTOMATED)")

	return cmd
}

func newUpdateCmd(opts *root.Options) *cobra.Command {
	var name, subject, fromName, replyTo string
	var jsonData string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a marketing email",
		Long:  "Update an existing marketing email in HubSpot.",
		Example: `  # Update email subject
  hspt marketing-emails update 12345 --subject "New Subject"

  # Update with JSON data
  hspt marketing-emails update 12345 --json '{"name": "Updated Name"}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			var updates map[string]interface{}

			if jsonData != "" {
				if err := json.Unmarshal([]byte(jsonData), &updates); err != nil {
					return fmt.Errorf("invalid JSON: %w", err)
				}
			} else {
				updates = make(map[string]interface{})
				if name != "" {
					updates["name"] = name
				}
				if subject != "" {
					updates["subject"] = subject
				}
				if fromName != "" {
					updates["fromName"] = fromName
				}
				if replyTo != "" {
					updates["replyTo"] = replyTo
				}
			}

			if len(updates) == 0 {
				return fmt.Errorf("no updates specified; use --name, --subject, --from-name, --reply-to, or --json")
			}

			email, err := client.UpdateMarketingEmail(id, updates)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Marketing email %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Marketing email %s updated", email.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Email name")
	cmd.Flags().StringVar(&subject, "subject", "", "Email subject line")
	cmd.Flags().StringVar(&fromName, "from-name", "", "Sender display name")
	cmd.Flags().StringVar(&replyTo, "reply-to", "", "Reply-to email address")
	cmd.Flags().StringVar(&jsonData, "json", "", "JSON object with fields to update")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a marketing email",
		Long:  "Archive/delete a marketing email in HubSpot.",
		Example: `  # Delete email by ID
  hspt marketing-emails delete 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			err = client.DeleteMarketingEmail(id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Marketing email %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Marketing email %s deleted", id)
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
