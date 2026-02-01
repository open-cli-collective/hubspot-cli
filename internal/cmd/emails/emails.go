package emails

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// DefaultProperties are the default properties to fetch for emails
var DefaultProperties = []string{"hs_email_subject", "hs_email_text", "hs_email_direction", "hs_email_status", "hs_timestamp", "hubspot_owner_id"}

// Register registers the emails command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "emails",
		Short: "Manage HubSpot email engagements",
		Long:  "Commands for listing, viewing, creating, updating, and deleting email engagements in HubSpot CRM.",
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
	var properties []string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List email engagements",
		Long:  "List email engagements from HubSpot CRM with pagination support.",
		Example: `  # List first 10 emails
  hspt emails list

  # List with pagination
  hspt emails list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = DefaultProperties
			}

			result, err := client.ListObjects(api.ObjectTypeEmails, api.ListOptions{
				Limit:      limit,
				After:      after,
				Properties: properties,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No emails found")
				return nil
			}

			headers := []string{"ID", "SUBJECT", "DIRECTION", "STATUS", "TIMESTAMP"}
			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, []string{
					obj.ID,
					truncate(obj.GetProperty("hs_email_subject"), 40),
					obj.GetProperty("hs_email_direction"),
					obj.GetProperty("hs_email_status"),
					obj.GetProperty("hs_timestamp"),
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
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var properties []string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get an email by ID",
		Long:  "Retrieve a single email engagement by its ID from HubSpot CRM.",
		Example: `  # Get email by ID
  hspt emails get 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = DefaultProperties
			}

			obj, err := client.GetObject(api.ObjectTypeEmails, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Email %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Subject", obj.GetProperty("hs_email_subject")},
				{"Text", truncate(obj.GetProperty("hs_email_text"), 100)},
				{"Direction", obj.GetProperty("hs_email_direction")},
				{"Status", obj.GetProperty("hs_email_status")},
				{"Timestamp", obj.GetProperty("hs_timestamp")},
				{"Owner ID", obj.GetProperty("hubspot_owner_id")},
				{"Created", obj.CreatedAt},
				{"Updated", obj.UpdatedAt},
			}

			return v.Render(headers, rows, obj)
		},
	}

	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}

func newCreateCmd(opts *root.Options) *cobra.Command {
	var subject, text, direction, status, timestamp, ownerID string
	var props []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new email engagement",
		Long:  "Create a new email engagement record in HubSpot CRM.",
		Example: `  # Create an email record
  hspt emails create --subject "Follow-up" --text "Email body content" --direction EMAIL

  # Create with status
  hspt emails create --subject "Proposal" --direction EMAIL --status SENT`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if subject != "" {
				properties["hs_email_subject"] = subject
			}
			if text != "" {
				properties["hs_email_text"] = text
			}
			if direction != "" {
				properties["hs_email_direction"] = direction
			}
			if status != "" {
				properties["hs_email_status"] = status
			}
			if timestamp != "" {
				properties["hs_timestamp"] = timestamp
			}
			if ownerID != "" {
				properties["hubspot_owner_id"] = ownerID
			}

			for _, p := range props {
				parts := strings.SplitN(p, "=", 2)
				if len(parts) == 2 {
					properties[parts[0]] = parts[1]
				}
			}

			if len(properties) == 0 {
				return fmt.Errorf("at least one property is required")
			}

			obj, err := client.CreateObject(api.ObjectTypeEmails, properties)
			if err != nil {
				return err
			}

			v.Success("Email created with ID: %s", obj.ID)

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Subject", obj.GetProperty("hs_email_subject")},
				{"Direction", obj.GetProperty("hs_email_direction")},
			}

			return v.Render(headers, rows, obj)
		},
	}

	cmd.Flags().StringVar(&subject, "subject", "", "Email subject")
	cmd.Flags().StringVar(&text, "text", "", "Email body text")
	cmd.Flags().StringVar(&direction, "direction", "", "Email direction (EMAIL, INCOMING_EMAIL, FORWARDED_EMAIL)")
	cmd.Flags().StringVar(&status, "status", "", "Email status (SENT, SCHEDULED, BOUNCED, etc.)")
	cmd.Flags().StringVar(&timestamp, "timestamp", "", "Email timestamp (Unix milliseconds)")
	cmd.Flags().StringVar(&ownerID, "owner-id", "", "HubSpot owner ID")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newUpdateCmd(opts *root.Options) *cobra.Command {
	var subject, text, direction, status, timestamp, ownerID string
	var props []string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an email engagement",
		Long:  "Update an existing email engagement in HubSpot CRM.",
		Example: `  # Update email subject
  hspt emails update 12345 --subject "Updated Subject"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if subject != "" {
				properties["hs_email_subject"] = subject
			}
			if text != "" {
				properties["hs_email_text"] = text
			}
			if direction != "" {
				properties["hs_email_direction"] = direction
			}
			if status != "" {
				properties["hs_email_status"] = status
			}
			if timestamp != "" {
				properties["hs_timestamp"] = timestamp
			}
			if ownerID != "" {
				properties["hubspot_owner_id"] = ownerID
			}

			for _, p := range props {
				parts := strings.SplitN(p, "=", 2)
				if len(parts) == 2 {
					properties[parts[0]] = parts[1]
				}
			}

			if len(properties) == 0 {
				return fmt.Errorf("at least one property to update is required")
			}

			obj, err := client.UpdateObject(api.ObjectTypeEmails, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Email %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Email %s updated", obj.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&subject, "subject", "", "Email subject")
	cmd.Flags().StringVar(&text, "text", "", "Email body text")
	cmd.Flags().StringVar(&direction, "direction", "", "Email direction (EMAIL, INCOMING_EMAIL, FORWARDED_EMAIL)")
	cmd.Flags().StringVar(&status, "status", "", "Email status (SENT, SCHEDULED, BOUNCED, etc.)")
	cmd.Flags().StringVar(&timestamp, "timestamp", "", "Email timestamp (Unix milliseconds)")
	cmd.Flags().StringVar(&ownerID, "owner-id", "", "HubSpot owner ID")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an email engagement",
		Long:  "Archive (soft delete) an email engagement in HubSpot CRM.",
		Example: `  # Delete email
  hspt emails delete 12345

  # Delete without confirmation
  hspt emails delete 12345 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if !force {
				v.Warning("This will archive email %s. Use --force to confirm.", id)
				return nil
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if err := client.DeleteObject(api.ObjectTypeEmails, id); err != nil {
				if api.IsNotFound(err) {
					v.Error("Email %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Email %s archived", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion without prompt")

	return cmd
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
