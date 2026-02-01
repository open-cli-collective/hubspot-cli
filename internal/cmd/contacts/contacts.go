package contacts

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// DefaultProperties are the default properties to fetch for contacts
var DefaultProperties = []string{"email", "firstname", "lastname", "phone", "company"}

// Register registers the contacts command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "contacts",
		Short: "Manage HubSpot contacts",
		Long:  "Commands for listing, viewing, creating, updating, and searching contacts in HubSpot CRM.",
	}

	cmd.AddCommand(newListCmd(opts))
	cmd.AddCommand(newGetCmd(opts))
	cmd.AddCommand(newCreateCmd(opts))
	cmd.AddCommand(newUpdateCmd(opts))
	cmd.AddCommand(newDeleteCmd(opts))
	cmd.AddCommand(newSearchCmd(opts))

	parent.AddCommand(cmd)
}

func newListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string
	var properties []string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List contacts",
		Long:  "List contacts from HubSpot CRM with pagination support.",
		Example: `  # List first 10 contacts
  hspt contacts list

  # List with custom properties
  hspt contacts list --properties email,firstname,lastname,company

  # List with pagination
  hspt contacts list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = DefaultProperties
			}

			result, err := client.ListObjects(api.ObjectTypeContacts, api.ListOptions{
				Limit:      limit,
				After:      after,
				Properties: properties,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No contacts found")
				return nil
			}

			headers := []string{"ID", "EMAIL", "FIRST NAME", "LAST NAME", "PHONE", "COMPANY"}
			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, []string{
					obj.ID,
					obj.GetProperty("email"),
					obj.GetProperty("firstname"),
					obj.GetProperty("lastname"),
					obj.GetProperty("phone"),
					obj.GetProperty("company"),
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of contacts to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var properties []string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a contact by ID",
		Long:  "Retrieve a single contact by its ID from HubSpot CRM.",
		Example: `  # Get contact by ID
  hspt contacts get 12345

  # Get with specific properties
  hspt contacts get 12345 --properties email,firstname,lastname`,
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

			obj, err := client.GetObject(api.ObjectTypeContacts, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Contact %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Email", obj.GetProperty("email")},
				{"First Name", obj.GetProperty("firstname")},
				{"Last Name", obj.GetProperty("lastname")},
				{"Phone", obj.GetProperty("phone")},
				{"Company", obj.GetProperty("company")},
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
	var email, firstname, lastname, phone, company string
	var props []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new contact",
		Long:  "Create a new contact in HubSpot CRM.",
		Example: `  # Create with common fields
  hspt contacts create --email john@example.com --firstname John --lastname Doe

  # Create with custom properties
  hspt contacts create --email john@example.com --prop lifecyclestage=customer`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if email != "" {
				properties["email"] = email
			}
			if firstname != "" {
				properties["firstname"] = firstname
			}
			if lastname != "" {
				properties["lastname"] = lastname
			}
			if phone != "" {
				properties["phone"] = phone
			}
			if company != "" {
				properties["company"] = company
			}

			// Parse custom properties
			for _, p := range props {
				parts := strings.SplitN(p, "=", 2)
				if len(parts) == 2 {
					properties[parts[0]] = parts[1]
				}
			}

			if len(properties) == 0 {
				return fmt.Errorf("at least one property is required")
			}

			obj, err := client.CreateObject(api.ObjectTypeContacts, properties)
			if err != nil {
				return err
			}

			v.Success("Contact created with ID: %s", obj.ID)

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", obj.ID},
				{"Email", obj.GetProperty("email")},
				{"First Name", obj.GetProperty("firstname")},
				{"Last Name", obj.GetProperty("lastname")},
			}

			return v.Render(headers, rows, obj)
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Contact email address")
	cmd.Flags().StringVar(&firstname, "firstname", "", "Contact first name")
	cmd.Flags().StringVar(&lastname, "lastname", "", "Contact last name")
	cmd.Flags().StringVar(&phone, "phone", "", "Contact phone number")
	cmd.Flags().StringVar(&company, "company", "", "Contact company name")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newUpdateCmd(opts *root.Options) *cobra.Command {
	var email, firstname, lastname, phone, company string
	var props []string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a contact",
		Long:  "Update an existing contact in HubSpot CRM.",
		Example: `  # Update contact name
  hspt contacts update 12345 --firstname Johnny

  # Update custom property
  hspt contacts update 12345 --prop lifecyclestage=customer`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			properties := make(map[string]interface{})
			if email != "" {
				properties["email"] = email
			}
			if firstname != "" {
				properties["firstname"] = firstname
			}
			if lastname != "" {
				properties["lastname"] = lastname
			}
			if phone != "" {
				properties["phone"] = phone
			}
			if company != "" {
				properties["company"] = company
			}

			// Parse custom properties
			for _, p := range props {
				parts := strings.SplitN(p, "=", 2)
				if len(parts) == 2 {
					properties[parts[0]] = parts[1]
				}
			}

			if len(properties) == 0 {
				return fmt.Errorf("at least one property to update is required")
			}

			obj, err := client.UpdateObject(api.ObjectTypeContacts, id, properties)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Contact %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Contact %s updated", obj.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Contact email address")
	cmd.Flags().StringVar(&firstname, "firstname", "", "Contact first name")
	cmd.Flags().StringVar(&lastname, "lastname", "", "Contact last name")
	cmd.Flags().StringVar(&phone, "phone", "", "Contact phone number")
	cmd.Flags().StringVar(&company, "company", "", "Contact company name")
	cmd.Flags().StringArrayVar(&props, "prop", nil, "Custom property in key=value format")

	return cmd
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a contact",
		Long:  "Archive (soft delete) a contact in HubSpot CRM.",
		Example: `  # Delete contact
  hspt contacts delete 12345

  # Delete without confirmation
  hspt contacts delete 12345 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if !force {
				v.Warning("This will archive contact %s. Use --force to confirm.", id)
				return nil
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if err := client.DeleteObject(api.ObjectTypeContacts, id); err != nil {
				if api.IsNotFound(err) {
					v.Error("Contact %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Contact %s archived", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion without prompt")

	return cmd
}

func newSearchCmd(opts *root.Options) *cobra.Command {
	var email, firstname, lastname string
	var query string
	var limit int
	var properties []string

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search contacts",
		Long:  "Search for contacts using filters.",
		Example: `  # Search by email
  hspt contacts search --email john@example.com

  # Search by name
  hspt contacts search --firstname John --lastname Doe

  # Full-text search
  hspt contacts search --query "john"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if len(properties) == 0 {
				properties = DefaultProperties
			}

			// Build filters
			var filters []api.SearchFilter

			if email != "" {
				filters = append(filters, api.SearchFilter{
					PropertyName: "email",
					Operator:     "EQ",
					Value:        email,
				})
			}
			if firstname != "" {
				filters = append(filters, api.SearchFilter{
					PropertyName: "firstname",
					Operator:     "CONTAINS_TOKEN",
					Value:        firstname,
				})
			}
			if lastname != "" {
				filters = append(filters, api.SearchFilter{
					PropertyName: "lastname",
					Operator:     "CONTAINS_TOKEN",
					Value:        lastname,
				})
			}

			req := api.SearchRequest{
				Properties: properties,
				Limit:      limit,
			}

			if len(filters) > 0 {
				req.FilterGroups = []api.SearchFilterGroup{
					{Filters: filters},
				}
			}

			// Note: HubSpot search API doesn't have a direct full-text query parameter
			// for contacts like some other APIs. The filters above handle specific field searches.
			if query != "" && len(filters) == 0 {
				// For simple query, search in email field
				req.FilterGroups = []api.SearchFilterGroup{
					{
						Filters: []api.SearchFilter{
							{
								PropertyName: "email",
								Operator:     "CONTAINS_TOKEN",
								Value:        query,
							},
						},
					},
				}
			}

			result, err := client.SearchObjects(api.ObjectTypeContacts, req)
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No contacts found matching criteria")
				return nil
			}

			headers := []string{"ID", "EMAIL", "FIRST NAME", "LAST NAME", "PHONE", "COMPANY"}
			rows := make([][]string, 0, len(result.Results))
			for _, obj := range result.Results {
				rows = append(rows, []string{
					obj.ID,
					obj.GetProperty("email"),
					obj.GetProperty("firstname"),
					obj.GetProperty("lastname"),
					obj.GetProperty("phone"),
					obj.GetProperty("company"),
				})
			}

			v.Info("Found %d contact(s)", len(result.Results))
			return v.Render(headers, rows, result)
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Search by exact email")
	cmd.Flags().StringVar(&firstname, "firstname", "", "Search by first name (contains)")
	cmd.Flags().StringVar(&lastname, "lastname", "", "Search by last name (contains)")
	cmd.Flags().StringVar(&query, "query", "", "Full-text search query")
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of results")
	cmd.Flags().StringSliceVar(&properties, "properties", nil, "Properties to include (comma-separated)")

	return cmd
}
