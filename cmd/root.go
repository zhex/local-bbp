package cmd

import (
	"github.com/spf13/cobra"
	"github/zhex/bbp/internal/runner"
)

func CreateRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "bbp",
		Short:   "bbp is the Bitbucket Pipelines CLI tool for local development and testing.",
		Long:    `bbp is the Bitbucket Pipelines CLI tool for local development and testing`,
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			proj := cmd.Flag("project").Value.String()
			name := cmd.Flag("name").Value.String()
			r := runner.New(proj)
			r.Run(name)
		},
	}

	rootCmd.Flags().StringP("name", "n", "default", "Name of the workflow to run")
	rootCmd.Flags().StringP("project", "p", ".", "Path to the project directory")
	return rootCmd
}
