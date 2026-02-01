package root

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/internal/config"
	"github.com/open-cli-collective/hubspot-cli/internal/version"
	"github.com/open-cli-collective/hubspot-cli/internal/view"
)

// Options contains global options for commands
type Options struct {
	Output  string
	NoColor bool
	Verbose bool
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

// View returns a configured View instance
func (o *Options) View() *view.View {
	v := view.New(o.Output, o.NoColor)
	v.Out = o.Stdout
	v.Err = o.Stderr
	return v
}

// GetAccessToken returns the access token from config or environment
func (o *Options) GetAccessToken() string {
	return config.GetAccessToken()
}

// NewCmd creates the root command and returns the options struct
func NewCmd() (*cobra.Command, *Options) {
	opts := &Options{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	cmd := &cobra.Command{
		Use:     "hspt",
		Short:   "A CLI for HubSpot",
		Long:    "hspt is a command-line interface for HubSpot CRM.",
		Version: version.Info(),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Setup is done in flag binding
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags - bound to opts struct
	cmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "table", "Output format: table, json, plain")
	cmd.PersistentFlags().BoolVar(&opts.NoColor, "no-color", false, "Disable colored output")
	cmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose output")

	return cmd, opts
}

// RegisterCommands registers subcommands with the root command
func RegisterCommands(root *cobra.Command, opts *Options, registrars ...func(*cobra.Command, *Options)) {
	for _, register := range registrars {
		register(root, opts)
	}
}

// GetOptions extracts Options from a root command
func GetOptions(cmd *cobra.Command) *Options {
	output, _ := cmd.Root().PersistentFlags().GetString("output")
	noColor, _ := cmd.Root().PersistentFlags().GetBool("no-color")
	verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")

	return &Options{
		Output:  output,
		NoColor: noColor,
		Verbose: verbose,
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}
}
