package associations

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the associations command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:     "associations",
		Aliases: []string{"assoc"},
		Short:   "Manage HubSpot associations",
		Long:    "Commands for listing, creating, and deleting associations between CRM objects.",
	}

	cmd.AddCommand(newListCmd(opts))
	cmd.AddCommand(newCreateCmd(opts))
	cmd.AddCommand(newDeleteCmd(opts))

	parent.AddCommand(cmd)
}

func newListCmd(opts *root.Options) *cobra.Command {
	var fromType, toType, fromID string
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List associations",
		Long:  "List associations from one object to another type.",
		Example: `  # List companies associated with a contact
  hspt associations list --from-type contacts --from-id 123 --to-type companies

  # List deals associated with a company
  hspt associations list --from-type companies --from-id 456 --to-type deals`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if fromType == "" || toType == "" || fromID == "" {
				return fmt.Errorf("--from-type, --from-id, and --to-type are required")
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListAssociations(
				api.ObjectType(fromType),
				fromID,
				api.ObjectType(toType),
				api.ListOptions{
					Limit: limit,
					After: after,
				},
			)
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No associations found")
				return nil
			}

			headers := []string{"TO OBJECT ID", "ASSOCIATION TYPE", "CATEGORY"}
			rows := make([][]string, 0, len(result.Results))
			for _, assoc := range result.Results {
				for _, at := range assoc.AssociationTypes {
					label := at.Label
					if label == "" {
						label = fmt.Sprintf("Type %d", at.TypeID)
					}
					rows = append(rows, []string{
						assoc.ToObjectID,
						label,
						at.Category,
					})
				}
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

	cmd.Flags().StringVar(&fromType, "from-type", "", "Source object type (contacts, companies, deals, etc.)")
	cmd.Flags().StringVar(&fromID, "from-id", "", "Source object ID")
	cmd.Flags().StringVar(&toType, "to-type", "", "Target object type (contacts, companies, deals, etc.)")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of associations to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func newCreateCmd(opts *root.Options) *cobra.Command {
	var fromType, toType, fromID, toID string
	var typeID int

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an association",
		Long:  "Create an association between two CRM objects.",
		Example: `  # Associate a contact with a company (default association type)
  hspt associations create --from-type contacts --from-id 123 --to-type companies --to-id 456

  # Associate with a specific type ID
  hspt associations create --from-type contacts --from-id 123 --to-type companies --to-id 456 --type-id 1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if fromType == "" || toType == "" || fromID == "" || toID == "" {
				return fmt.Errorf("--from-type, --from-id, --to-type, and --to-id are required")
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			err = client.CreateAssociation(
				api.ObjectType(fromType),
				fromID,
				api.ObjectType(toType),
				toID,
				typeID,
			)
			if err != nil {
				return err
			}

			v.Success("Association created between %s %s and %s %s", fromType, fromID, toType, toID)
			return nil
		},
	}

	cmd.Flags().StringVar(&fromType, "from-type", "", "Source object type (contacts, companies, deals, etc.)")
	cmd.Flags().StringVar(&fromID, "from-id", "", "Source object ID")
	cmd.Flags().StringVar(&toType, "to-type", "", "Target object type (contacts, companies, deals, etc.)")
	cmd.Flags().StringVar(&toID, "to-id", "", "Target object ID")
	cmd.Flags().IntVar(&typeID, "type-id", 1, "Association type ID (default: 1 for standard association)")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var fromType, toType, fromID, toID string
	var force bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an association",
		Long:  "Delete an association between two CRM objects.",
		Example: `  # Delete association between contact and company
  hspt associations delete --from-type contacts --from-id 123 --to-type companies --to-id 456 --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if fromType == "" || toType == "" || fromID == "" || toID == "" {
				return fmt.Errorf("--from-type, --from-id, --to-type, and --to-id are required")
			}

			if !force {
				v.Warning("This will delete the association between %s %s and %s %s. Use --force to confirm.", fromType, fromID, toType, toID)
				return nil
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			err = client.DeleteAssociation(
				api.ObjectType(fromType),
				fromID,
				api.ObjectType(toType),
				toID,
			)
			if err != nil {
				return err
			}

			v.Success("Association deleted between %s %s and %s %s", fromType, fromID, toType, toID)
			return nil
		},
	}

	cmd.Flags().StringVar(&fromType, "from-type", "", "Source object type (contacts, companies, deals, etc.)")
	cmd.Flags().StringVar(&fromID, "from-id", "", "Source object ID")
	cmd.Flags().StringVar(&toType, "to-type", "", "Target object type (contacts, companies, deals, etc.)")
	cmd.Flags().StringVar(&toID, "to-id", "", "Target object ID")
	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion without prompt")

	return cmd
}
