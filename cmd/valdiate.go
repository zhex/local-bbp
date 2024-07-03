package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github/zhex/local-bbp/internal/validator"
	"os"
	"path"
)

func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validate",
		Short:   "Validate the bitbucket-pipelines.yml file",
		Example: `bbp validate`,
		Run: func(cmd *cobra.Command, args []string) {
			proj := cmd.Flag("project").Value.String()
			file := path.Join(proj, "bitbucket-pipelines.yml")
			data, err := os.ReadFile(file)
			if err != nil {
				log.Fatalf("Failed reading yaml file: %s", err)
			}
			err = validator.ValidatePipelineYaml(data)
			if err != nil {
				validator.PrintError(err)
				os.Exit(1)
			} else {
				log.Info("Validation successful")
			}
		},
	}

	return cmd
}
