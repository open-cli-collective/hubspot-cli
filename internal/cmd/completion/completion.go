package completion

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the completion command
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for hspt.

To load completions:

Bash:
  $ source <(hspt completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ hspt completion bash > /etc/bash_completion.d/hspt
  # macOS:
  $ hspt completion bash > $(brew --prefix)/etc/bash_completion.d/hspt

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ hspt completion zsh > "${fpath[1]}/_hspt"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ hspt completion fish | source
  # To load completions for each session, execute once:
  $ hspt completion fish > ~/.config/fish/completions/hspt.fish

PowerShell:
  PS> hspt completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> hspt completion powershell > hspt.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}

	parent.AddCommand(cmd)
}
