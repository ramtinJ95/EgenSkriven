package commands

import (
	"os"

	"github.com/spf13/cobra"
)

// newCompletionCmd creates the completion command with subcommands for each shell
func newCompletionCmd(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [shell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for EgenSkriven.

To load completions:

Bash:
  # Linux
  $ egenskriven completion bash > /etc/bash_completion.d/egenskriven
  
  # macOS (requires bash-completion@2)
  $ egenskriven completion bash > $(brew --prefix)/etc/bash_completion.d/egenskriven

Zsh:
  # If shell completion is not already enabled, you need to enable it first:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # Add to your ~/.zshrc or run once:
  $ egenskriven completion zsh > "${fpath[1]}/_egenskriven"
  
  # Or for Oh My Zsh:
  $ egenskriven completion zsh > ~/.oh-my-zsh/completions/_egenskriven

Fish:
  $ egenskriven completion fish > ~/.config/fish/completions/egenskriven.fish

PowerShell:
  # Add to your PowerShell profile:
  PS> egenskriven completion powershell >> $PROFILE

After installing, restart your shell or source the completion file.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}

	return cmd
}
