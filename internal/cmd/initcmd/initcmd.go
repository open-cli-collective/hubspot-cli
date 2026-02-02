package initcmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
	"github.com/open-cli-collective/hubspot-cli/internal/config"
)

// Register registers the init command
func Register(parent *cobra.Command, opts *root.Options) {
	var token string
	var noVerify bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize hspt with guided setup",
		Long: `Interactive setup wizard for configuring hspt.

Prompts for your HubSpot access token, then optionally verifies
the connection before saving the configuration.

Get your access token from: HubSpot Settings > Integrations > Private Apps`,
		Example: `  # Interactive setup
  hspt init

  # Non-interactive setup
  hspt init --token YOUR_ACCESS_TOKEN

  # Skip connection verification
  hspt init --no-verify`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(opts, token, noVerify)
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "HubSpot access token")
	cmd.Flags().BoolVar(&noVerify, "no-verify", false, "Skip connection verification")

	parent.AddCommand(cmd)
}

func runInit(opts *root.Options, prefillToken string, noVerify bool) error {
	configPath := config.Path()

	// Load existing config for pre-population
	existingCfg, _ := config.Load()
	if existingCfg == nil {
		existingCfg = &config.Config{}
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		var overwrite bool
		err := huh.NewConfirm().
			Title("Configuration already exists").
			Description(fmt.Sprintf("Overwrite %s?", configPath)).
			Value(&overwrite).
			Run()
		if err != nil {
			return err
		}
		if !overwrite {
			fmt.Println("Initialization cancelled.")
			return nil
		}
	}

	cfg := &config.Config{}

	// Pre-fill from existing config, then override with CLI flags
	// Priority: CLI flag > existing config value
	if prefillToken != "" {
		cfg.AccessToken = prefillToken
	} else if existingCfg.AccessToken != "" {
		cfg.AccessToken = existingCfg.AccessToken
	}

	// Build the form
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Access Token").
				Description("Get one from: HubSpot Settings > Integrations > Private Apps").
				EchoMode(huh.EchoModePassword).
				Value(&cfg.AccessToken).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("access token is required")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	// Verify connection unless --no-verify
	if !noVerify {
		fmt.Print("Verifying connection... ")
		client, err := api.New(api.ClientConfig{
			AccessToken: cfg.AccessToken,
			Verbose:     opts.Verbose,
		})
		if err != nil {
			fmt.Println("failed!")
			return fmt.Errorf("failed to create client: %w", err)
		}

		owners, err := client.GetOwners()
		if err != nil {
			fmt.Println("failed!")
			fmt.Println()
			fmt.Println("Troubleshooting:")
			fmt.Println("  - Check your access token is correct")
			fmt.Println("  - Ensure your private app has the required scopes")
			fmt.Println()
			fmt.Println("To get a new access token:")
			fmt.Println("  HubSpot Settings > Integrations > Private Apps")
			return fmt.Errorf("connection verification failed: %w", err)
		}
		fmt.Println("success!")
		fmt.Println()
		fmt.Printf("HubSpot account has %d owners\n", len(owners))
		if len(owners) > 0 {
			fmt.Printf("First owner: %s (%s)\n", owners[0].FullName(), owners[0].Email)
		}
	}

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("\nConfiguration saved to %s\n", configPath)
	fmt.Println("\nYou're all set! Try running:")
	fmt.Println("  hspt config show")

	return nil
}
