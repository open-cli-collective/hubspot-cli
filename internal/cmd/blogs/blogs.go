package blogs

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the blogs command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "blogs",
		Short: "Manage HubSpot blog content",
		Long:  "Commands for managing blog posts, authors, and tags.",
	}

	cmd.AddCommand(newPostsCmd(opts))
	cmd.AddCommand(newAuthorsCmd(opts))
	cmd.AddCommand(newTagsCmd(opts))

	parent.AddCommand(cmd)
}

func newPostsCmd(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "posts",
		Short: "Manage blog posts",
		Long:  "Commands for listing, viewing, creating, updating, and deleting blog posts.",
	}

	cmd.AddCommand(newPostsListCmd(opts))
	cmd.AddCommand(newPostsGetCmd(opts))
	cmd.AddCommand(newPostsCreateCmd(opts))
	cmd.AddCommand(newPostsUpdateCmd(opts))
	cmd.AddCommand(newPostsDeleteCmd(opts))

	return cmd
}

func newPostsListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List blog posts",
		Long:  "List blog posts from HubSpot CMS.",
		Example: `  # List blog posts
  hspt blogs posts list

  # List with pagination
  hspt blogs posts list --limit 20`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListBlogPosts(api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No blog posts found")
				return nil
			}

			headers := []string{"ID", "NAME", "SLUG", "STATE", "AUTHOR"}
			rows := make([][]string, 0, len(result.Results))
			for _, post := range result.Results {
				rows = append(rows, []string{
					post.ID,
					truncate(post.Name, 30),
					truncate(post.Slug, 25),
					post.State,
					post.AuthorName,
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of posts to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func newPostsGetCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a blog post by ID",
		Long:  "Retrieve a single blog post by its ID.",
		Example: `  # Get blog post by ID
  hspt blogs posts get 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			post, err := client.GetBlogPost(id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Blog post %s not found", id)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", post.ID},
				{"Name", post.Name},
				{"Slug", post.Slug},
				{"State", post.State},
				{"HTML Title", post.HTMLTitle},
				{"Meta Description", truncate(post.MetaDescription, 50)},
				{"Summary", truncate(post.PostSummary, 50)},
				{"Author", post.AuthorName},
				{"Author ID", post.BlogAuthorID},
				{"Featured Image", truncate(post.FeaturedImage, 40)},
				{"Archived", formatBool(post.Archived)},
				{"Created", post.CreatedAt},
				{"Updated", post.UpdatedAt},
			}

			if post.PublishDate != "" {
				rows = append(rows, []string{"Published", post.PublishDate})
			}

			return v.Render(headers, rows, post)
		},
	}
}

func newPostsCreateCmd(opts *root.Options) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new blog post",
		Long:  "Create a new blog post in HubSpot CMS.",
		Example: `  # Create a blog post from JSON file
  hspt blogs posts create --file post.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if file == "" {
				return fmt.Errorf("--file is required")
			}

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var postData map[string]interface{}
			if err := json.Unmarshal(data, &postData); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			post, err := client.CreateBlogPost(postData)
			if err != nil {
				return err
			}

			v.Success("Blog post created with ID: %s", post.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "JSON file containing post data (required)")

	return cmd
}

func newPostsUpdateCmd(opts *root.Options) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a blog post",
		Long:  "Update an existing blog post in HubSpot CMS.",
		Example: `  # Update a blog post from JSON file
  hspt blogs posts update 12345 --file updates.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			if file == "" {
				return fmt.Errorf("--file is required")
			}

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var updates map[string]interface{}
			if err := json.Unmarshal(data, &updates); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			post, err := client.UpdateBlogPost(id, updates)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Blog post %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Blog post %s updated", post.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "JSON file containing post updates (required)")

	return cmd
}

func newPostsDeleteCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a blog post",
		Long:  "Archive/delete a blog post in HubSpot CMS.",
		Example: `  # Delete a blog post
  hspt blogs posts delete 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()
			id := args[0]

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			err = client.DeleteBlogPost(id)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Blog post %s not found", id)
					return nil
				}
				return err
			}

			v.Success("Blog post %s deleted", id)
			return nil
		},
	}
}

func newAuthorsCmd(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authors",
		Short: "Manage blog authors",
		Long:  "Commands for listing blog authors.",
	}

	cmd.AddCommand(newAuthorsListCmd(opts))

	return cmd
}

func newAuthorsListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List blog authors",
		Long:  "List blog authors from HubSpot CMS.",
		Example: `  # List blog authors
  hspt blogs authors list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListBlogAuthors(api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No blog authors found")
				return nil
			}

			headers := []string{"ID", "NAME", "EMAIL", "DISPLAY NAME"}
			rows := make([][]string, 0, len(result.Results))
			for _, author := range result.Results {
				displayName := author.DisplayName
				if displayName == "" {
					displayName = author.FullName
				}
				rows = append(rows, []string{
					author.ID,
					author.Name,
					author.Email,
					displayName,
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of authors to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
}

func newTagsCmd(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "Manage blog tags",
		Long:  "Commands for listing blog tags.",
	}

	cmd.AddCommand(newTagsListCmd(opts))

	return cmd
}

func newTagsListCmd(opts *root.Options) *cobra.Command {
	var limit int
	var after string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List blog tags",
		Long:  "List blog tags from HubSpot CMS.",
		Example: `  # List blog tags
  hspt blogs tags list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListBlogTags(api.ListOptions{
				Limit: limit,
				After: after,
			})
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No blog tags found")
				return nil
			}

			headers := []string{"ID", "NAME", "SLUG", "LANGUAGE"}
			rows := make([][]string, 0, len(result.Results))
			for _, tag := range result.Results {
				rows = append(rows, []string{
					tag.ID,
					tag.Name,
					tag.Slug,
					tag.Language,
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

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of tags to return")
	cmd.Flags().StringVar(&after, "after", "", "Pagination cursor for the next page")

	return cmd
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
