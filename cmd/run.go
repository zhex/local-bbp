package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github/zhex/local-bbp/internal/parser"
	"github/zhex/local-bbp/internal/runner"
	"os"
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a pipeline",
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

	cmd.Flags().StringP("name", "n", "default", "Name of the workflow to run")
	cmd.Flags().StringP("secrets-file", "s", "", "Path to the secrets file")

	return cmd
}
