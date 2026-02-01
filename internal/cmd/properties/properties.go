package properties

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the properties command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:     "properties",
		Aliases: []string{"props"},
		Short:   "Manage HubSpot properties",
		Long:    "Commands for listing, viewing, creating, and deleting object properties.",
	}

	cmd.AddCommand(newListCmd(opts))
	cmd.AddCommand(newGetCmd(opts))
	cmd.AddCommand(newCreateCmd(opts))
	cmd.AddCommand(newDeleteCmd(opts))

	parent.AddCommand(cmd)
}

func newListCmd(opts *root.Options) *cobra.Command {
	var objectType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List properties",
		Long:  "List all properties for an object type.",
		Example: `  # List contact properties
  hspt properties list --object-type contacts

  # List deal properties
  hspt properties list --object-type deals`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if objectType == "" {
				return fmt.Errorf("--object-type is required")
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListProperties(api.ObjectType(objectType))
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No properties found")
				return nil
			}

			headers := []string{"NAME", "LABEL", "TYPE", "FIELD TYPE", "GROUP"}
			rows := make([][]string, 0, len(result.Results))
			for _, prop := range result.Results {
				rows = append(rows, []string{
					prop.Name,
					prop.Label,
					prop.Type,
					prop.FieldType,
					prop.GroupName,
				})
			}

			v.Info("Found %d properties for %s", len(result.Results), objectType)
			return v.Render(headers, rows, result)
		},
	}

	cmd.Flags().StringVar(&objectType, "object-type", "", "Object type (contacts, companies, deals, tickets, etc.)")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var objectType, name string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a property",
		Long:  "Get details of a specific property.",
		Example: `  # Get the email property for contacts
  hspt properties get --object-type contacts --name email

  # Get a custom property
  hspt properties get --object-type contacts --name my_custom_field`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if objectType == "" || name == "" {
				return fmt.Errorf("--object-type and --name are required")
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			prop, err := client.GetProperty(api.ObjectType(objectType), name)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Property %s not found for %s", name, objectType)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"Name", prop.Name},
				{"Label", prop.Label},
				{"Type", prop.Type},
				{"Field Type", prop.FieldType},
				{"Group", prop.GroupName},
				{"Description", prop.Description},
				{"HubSpot Defined", formatBool(prop.HubspotDefined)},
				{"Calculated", formatBool(prop.Calculated)},
				{"Hidden", formatBool(prop.Hidden)},
				{"Created", prop.CreatedAt},
				{"Updated", prop.UpdatedAt},
			}

			if len(prop.Options) > 0 {
				rows = append(rows, []string{"Options", fmt.Sprintf("%d options", len(prop.Options))})
			}

			return v.Render(headers, rows, prop)
		},
	}

	cmd.Flags().StringVar(&objectType, "object-type", "", "Object type (contacts, companies, deals, tickets, etc.)")
	cmd.Flags().StringVar(&name, "name", "", "Property name")

	return cmd
}

func newCreateCmd(opts *root.Options) *cobra.Command {
	var objectType, name, label, propType, fieldType, groupName, description string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a property",
		Long:  "Create a new custom property for an object type.",
		Example: `  # Create a text property
  hspt properties create --object-type contacts --name my_field --label "My Field" --type string --field-type text

  # Create a number property
  hspt properties create --object-type deals --name custom_amount --label "Custom Amount" --type number --field-type number`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if objectType == "" || name == "" || label == "" || propType == "" || fieldType == "" {
				return fmt.Errorf("--object-type, --name, --label, --type, and --field-type are required")
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			req := api.CreatePropertyRequest{
				Name:      name,
				Label:     label,
				Type:      propType,
				FieldType: fieldType,
				GroupName: groupName,
			}

			if description != "" {
				req.Description = description
			}

			if groupName == "" {
				// Default group name based on object type
				req.GroupName = objectType + "information"
			}

			prop, err := client.CreateProperty(api.ObjectType(objectType), req)
			if err != nil {
				return err
			}

			v.Success("Property %s created for %s", prop.Name, objectType)

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"Name", prop.Name},
				{"Label", prop.Label},
				{"Type", prop.Type},
				{"Field Type", prop.FieldType},
				{"Group", prop.GroupName},
			}

			return v.Render(headers, rows, prop)
		},
	}

	cmd.Flags().StringVar(&objectType, "object-type", "", "Object type (contacts, companies, deals, tickets, etc.)")
	cmd.Flags().StringVar(&name, "name", "", "Property name (internal identifier)")
	cmd.Flags().StringVar(&label, "label", "", "Property label (display name)")
	cmd.Flags().StringVar(&propType, "type", "", "Property type (string, number, date, datetime, enumeration, bool)")
	cmd.Flags().StringVar(&fieldType, "field-type", "", "Field type (text, textarea, number, date, select, checkbox, etc.)")
	cmd.Flags().StringVar(&groupName, "group", "", "Property group name")
	cmd.Flags().StringVar(&description, "description", "", "Property description")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var objectType, name string
	var force bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a property",
		Long:  "Archive (soft delete) a custom property.",
		Example: `  # Delete a custom property
  hspt properties delete --object-type contacts --name my_field --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if objectType == "" || name == "" {
				return fmt.Errorf("--object-type and --name are required")
			}

			if !force {
				v.Warning("This will archive property %s for %s. Use --force to confirm.", name, objectType)
				return nil
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if err := client.DeleteProperty(api.ObjectType(objectType), name); err != nil {
				if api.IsNotFound(err) {
					v.Error("Property %s not found for %s", name, objectType)
					return nil
				}
				return err
			}

			v.Success("Property %s archived for %s", name, objectType)
			return nil
		},
	}

	cmd.Flags().StringVar(&objectType, "object-type", "", "Object type (contacts, companies, deals, tickets, etc.)")
	cmd.Flags().StringVar(&name, "name", "", "Property name")
	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion without prompt")

	return cmd
}

func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
