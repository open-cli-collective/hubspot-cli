package configcmd

import (
	"os"

	"github.com/spf13/cobra"

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

	cmd.AddCommand(newSetCmd(opts))
	cmd.AddCommand(newShowCmd(opts))
	cmd.AddCommand(newClearCmd(opts))
	cmd.AddCommand(newTestCmd(opts))

	parent.AddCommand(cmd)
}

func newSetCmd(opts *root.Options) *cobra.Command {
	var token string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set configuration values",
		Long:  "Set HubSpot credentials.",
		Example: `  # Set access token
  hspt config set --token YOUR_ACCESS_TOKEN

  # Using environment variable instead
  export HUBSPOT_ACCESS_TOKEN=YOUR_ACCESS_TOKEN`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if token != "" {
				cfg.AccessToken = token
			}

			if err := config.Save(cfg); err != nil {
				return err
			}

			v.Success("Configuration saved to %s", config.Path())
			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "HubSpot access token")

	return cmd
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
			maskedToken := ""
			if token != "" {
				if len(token) > 8 {
					maskedToken = token[:4] + "..." + token[len(token)-4:]
				} else {
					maskedToken = "****"
				}
			}

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

func newClearCmd(opts *root.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear stored configuration",
		Long:  "Remove the stored configuration file. Environment variables will still work.",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if err := config.Clear(); err != nil {
				return err
			}

			v.Success("Configuration cleared")
			return nil
		},
	}
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

			// TODO: Add API verification once api package is implemented
			v.Warning("Connection verification not yet implemented")
			v.Println("")
			v.Info("Token is configured (masked): %s...%s", token[:4], token[len(token)-4:])

			return nil
		},
	}
}
