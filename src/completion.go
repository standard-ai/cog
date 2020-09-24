package main

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCommand = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash completion scripts",
	Long: `Enable tab completion for cog commands in bash.

Installation on Linux:

Completion files are commonly stored in /etc/bash_completion.d/ for
system-wide commands, but can be stored in
~/.local/share/bash-completion/completions for user-specific commands.
Run the commands:

    $ mkdir -p ~/.local/share/bash-completion/completions
    $ cog completion > ~/.local/share/bash-completion/completions/cog

Installation on OS X/Homebrew:

Make sure you have bash-completion brew formula installed, and you
have enabled completions in your bash shell as documented in
https://docs.brew.sh/Shell-Completion. Run the commands:

    $ mkdir -p $(brew --prefix)/etc/bash_completion.d
    $ cog completion > $(brew --prefix)/etc/bash_completion.d/cog

You may have to log out and log back in to your shell session for the
changes to take affect.
`,
	Run: func(cmd *cobra.Command, args []string) {
		rootCmd.GenBashCompletion(os.Stdout)
	},
}
