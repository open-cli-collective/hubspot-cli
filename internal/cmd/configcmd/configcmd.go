package configcmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
	"github.com/open-cli-collective/hubspot-cli/internal/config"
)

// Register registers the config commands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
		Long:  "Commands for managing hspt configuration and credentials.",
	}

	cmd.AddCommand(newShowCmd(opts))
	cmd.AddCommand(newClearCmd(opts))
	cmd.AddCommand(newTestCmd(opts))

	parent.AddCommand(cmd)
}

func newShowCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  "Display the current configuration values (token is masked).",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			token := config.GetAccessToken()

			// Mask the token
			maskedToken := maskToken(token)

			headers := []string{"KEY", "VALUE", "SOURCE"}
			rows := [][]string{
				{"access_token", maskedToken, getTokenSource()},
			}

			data := map[string]string{
				"access_token": maskedToken,
				"path":         config.Path(),
			}

			if err := v.Render(headers, rows, data); err != nil {
				return err
			}

			v.Info("\nConfig file: %s", config.Path())
			return nil
		},
	}
}

func maskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return "********"
	}
	return token[:4] + "********" + token[len(token)-4:]
}

func newClearCmd(opts *root.Options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear stored configuration",
		Long:  "Remove the stored configuration file. Environment variables will still work.",
		Example: `  # Clear with confirmation prompt
  hspt config clear

  # Clear without confirmation
  hspt config clear --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if !force {
				fmt.Print("This will remove all stored credentials. Continue? [y/N]: ")
				var response string
				_, _ = fmt.Scanln(&response)
				response = strings.TrimSpace(strings.ToLower(response))
				if response != "y" && response != "yes" {
					v.Info("Clear cancelled")
					return nil
				}
			}

			if err := config.Clear(); err != nil {
				return err
			}

			v.Success("Configuration cleared")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func getTokenSource() string {
	if os.Getenv("HUBSPOT_ACCESS_TOKEN") != "" {
		return "env (HUBSPOT_ACCESS_TOKEN)"
	}
	cfg, err := config.Load()
	if err != nil {
		return "-"
	}
	if cfg.AccessToken != "" {
		return "config"
	}
	return "-"
}

func newTestCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test connection to HubSpot",
		Long: `Verify that hspt can connect to HubSpot with the current configuration.

This command tests authentication and API access, providing clear
pass/fail status and troubleshooting suggestions on failure.`,
		Example: `  # Test connection
  hspt config test`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			token := config.GetAccessToken()
			if token == "" {
				v.Error("No HubSpot access token configured")
				v.Println("")
				v.Info("Configure with: hspt init")
				v.Info("Or set environment variable: HUBSPOT_ACCESS_TOKEN")
				return nil
			}

			v.Println("Testing connection to HubSpot...")
			v.Println("")

			client, err := opts.APIClient()
			if err != nil {
				v.Error("Failed to create client: %s", err)
				return nil
			}

			owners, err := client.GetOwners()
			if err != nil {
				if api.IsUnauthorized(err) {
					v.Error("Authentication failed: invalid access token")
					v.Println("")
					v.Info("Check your token at: HubSpot Settings > Integrations > Private Apps")
					return nil
				}
				if api.IsForbidden(err) {
					v.Error("Authentication failed: missing required scopes")
					v.Println("")
					v.Info("Ensure your private app has the required scopes enabled")
					return nil
				}
				v.Error("Connection failed: %s", err)
				return nil
			}

			v.Success("Connection successful!")
			v.Println("")
			v.Info("HubSpot account has %d owners", len(owners))
			if len(owners) > 0 {
				v.Info("First owner: %s (%s)", owners[0].FullName(), owners[0].Email)
			}

			return nil
		},
	}
}
