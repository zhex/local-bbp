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
		Short:   "bbp is a simple CI/CD tool",
		Long:    `bbp is a simple CI/CD tool that reads a .ci.yaml file and runs the steps defined in it.`,
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
