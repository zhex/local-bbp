package cmd

import (
	"github.com/spf13/cobra"
)

func CreateRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "bbp",
		Short:   "bbp is the Bitbucket Pipelines CLI tool for local development and testing.",
		Version: version,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	rootCmd.PersistentFlags().StringP("project", "p", ".", "Path to the project directory")

	rootCmd.AddCommand(
		newListCmd(),
		newRunCmd(),
		newValidateCmd(),
		newIntegrationsCmd(),
	)

	return rootCmd
}
