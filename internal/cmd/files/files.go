package files

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the files command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "files",
		Short: "Manage HubSpot files",
		Long:  "Commands for listing and viewing files in the HubSpot File Manager.",
	}

	cmd.AddCommand(newListCmd(opts))
	cmd.AddCommand(newGetCmd(opts))
	cmd.AddCommand(newDeleteCmd(opts))
	cmd.AddCommand(newFoldersCmd(opts))

	parent.AddCommand(cmd)
}

func newListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List files",
		Long:  "List files from the HubSpot File Manager with pagination support.",
		Example: `  # List first 10 files
  hspt files list

  # List with pagination
  hspt files list --limit 50 --after abc123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListFiles(api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No files found")
				return nil
			}

			headers := []string{"ID", "NAME", "TYPE", "SIZE", "ACCESS"}
			rows := make([][]string, 0, len(result.Results))
			for _, file := range result.Results {
				rows = append(rows, []string{
					file.ID,
					file.Name,
					file.Type,
					formatSize(file.Size),
					file.AccessLevel,
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of files to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a file by ID",
		Long:  "Retrieve a single file by its ID from the HubSpot File Manager.",
		Example: `  # Get file by ID
  hspt files get 12345678`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			file, err := client.GetFile(id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("File %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", file.ID},
				{"Name", file.Name},
				{"Path", file.Path},
				{"Type", file.Type},
				{"Extension", file.Extension},
				{"Size", formatSize(file.Size)},
				{"Access", file.AccessLevel},
				{"URL", file.URL},
				{"Archived", formatBool(file.Archived)},
				{"Created", file.CreatedAt},
				{"Updated", file.UpdatedAt},
			}

			return v.Render(headers, rows, file)
		},
	}
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a file",
		Long:  "Delete a file from the HubSpot File Manager.",
		Example: `  # Delete file
  hspt files delete 12345678

  # Delete without confirmation
  hspt files delete 12345678 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if !force {
				v.Warning("This will permanently delete file %s. Use --force to confirm.", id)
				return nil
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			if err := client.DeleteFile(id); err != nil {
				if api.IsNotFound(err) {
					v.Error("File %s not found", id)
					return nil
				}
				return err
			}

			v.Success("File %s deleted", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion without prompt")

	return cmd
}

func newFoldersCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "folders",
		Short: "List folders",
		Long:  "List folders from the HubSpot File Manager.",
		Example: `  # List folders
  hspt files folders`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListFolders(api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No folders found")
				return nil
			}

			headers := []string{"ID", "NAME", "PATH", "ARCHIVED"}
			rows := make([][]string, 0, len(result.Results))
			for _, folder := range result.Results {
				rows = append(rows, []string{
					folder.ID,
					folder.Name,
					folder.Path,
					formatBool(folder.Archived),
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of folders to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
