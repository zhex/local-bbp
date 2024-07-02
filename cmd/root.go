package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github/zhex/bbp/internal/parser"
	"github/zhex/bbp/internal/runner"
	"os"
)

func CreateRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "bbp",
		Short:   "bbp is the Bitbucket Pipelines CLI tool for local development and testing.",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			proj := cmd.Flag("project").Value.String()
			name := cmd.Flag("name").Value.String()

			var secrets map[string]string

			secretFile := cmd.Flag("secrets-file").Value.String()
			if secretFile != "" {
				data, err := os.ReadFile(secretFile)
				if err != nil {
					log.Fatalf("Error reading secrets file: %s", err)
				}
				secrets = parser.ParseSecrets(data)
			}

			r := runner.New(proj, secrets)
			r.Run(name)
		},
	}

	rootCmd.Flags().StringP("name", "n", "default", "ProjectName of the workflow to run")
	rootCmd.Flags().StringP("project", "p", ".", "Path to the project directory")
	rootCmd.Flags().StringP("secrets-file", "s", "", "Path to the secrets file")
	return rootCmd
}
