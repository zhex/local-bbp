package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zhex/local-bbp/internal/config"
	"github.com/zhex/local-bbp/internal/parser"
	"github.com/zhex/local-bbp/internal/runner"
	"os"
	"path/filepath"
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

			c, err := config.LoadConfig()
			if err != nil {
				log.Fatalf("Error loading config: %s", err)
			}

			fullPath, _ := filepath.Abs(proj)
			if !filepath.IsAbs(c.OutputDir) {
				c.OutputDir = filepath.Join(fullPath, c.OutputDir)
			}
			r := runner.New(fullPath, c, secrets)
			r.Run(name)
		},
	}

	cmd.Flags().StringP("name", "n", "default", "Name of the workflow to run")
	cmd.Flags().StringP("secrets-file", "s", "", "Path to the secrets file")

	return cmd
}
