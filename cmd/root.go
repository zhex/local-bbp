package cmd

import (
	"github.com/spf13/cobra"
	"github/zhex/bbp/pkg/models"
	"github/zhex/bbp/pkg/runner"
	"gopkg.in/yaml.v3"
	"os"
)

func CreateRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "bbp",
		Short:   "bbp is the Bitbucket Pipelines CLI tool for local development and testing.",
		Long:    `bbp is the Bitbucket Pipelines CLI tool for local development and testing`,
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			data, err := os.ReadFile("testdata/bitbucket-pipelines.yml")
			if err != nil {
				panic(err)
			}
			var plan models.Plan
			err = yaml.Unmarshal(data, &plan)
			if err != nil {
				panic(err)
			}
			r := runner.NewRunner(&plan)

			name := cmd.Flag("name").Value.String()
			r.Run(name)
		},
	}

	rootCmd.Flags().StringP("name", "n", "default", "Name of the workflow to run")
	return rootCmd
}
