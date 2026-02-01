package initcmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

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

func runInit(opts *root.Options, token string, noVerify bool) error {
	v := opts.View()
	reader := bufio.NewReader(opts.Stdin)

	v.Println("HubSpot CLI Setup")
	v.Println("")

	// Check for existing config
	existingCfg, _ := config.Load()
	if existingCfg.AccessToken != "" {
		v.Warning("Existing configuration found at %s", config.Path())
		v.Println("")

		overwrite, err := promptYesNo(reader, "Overwrite existing configuration?", false)
		if err != nil {
			return err
		}
		if !overwrite {
			v.Info("Setup cancelled")
			return nil
		}
		v.Println("")
	}

	// Prompt for token if not provided
	if token == "" {
		v.Println("Enter your HubSpot access token")
		v.Println("  Get one from: HubSpot Settings > Integrations > Private Apps")
		v.Println("")

		var err error
		token, err = promptRequired(reader, "Access Token")
		if err != nil {
			return err
		}
	}

	v.Println("")

	// Verify connection unless --no-verify
	if !noVerify {
		v.Println("Testing connection...")
		// TODO: Add API verification once api package is implemented
		v.Warning("Connection verification not yet implemented")
		v.Println("")
	}

	// Save configuration
	cfg := &config.Config{
		AccessToken: token,
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	v.Success("Configuration saved to %s", config.Path())
	v.Println("")
	v.Println("Try it out:")
	v.Println("  hspt config show")

	return nil
}

func promptRequired(reader *bufio.Reader, label string) (string, error) {
	for {
		fmt.Printf("%s: ", label)
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		input = strings.TrimSpace(input)
		if input != "" {
			return input, nil
		}
		fmt.Printf("  %s is required\n", label)
	}
}

func promptYesNo(reader *bufio.Reader, question string, defaultYes bool) (bool, error) {
	suffix := " [y/N]: "
	if defaultYes {
		suffix = " [Y/n]: "
	}

	fmt.Print(question + suffix)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return defaultYes, nil
	}
	return input == "y" || input == "yes", nil
}
