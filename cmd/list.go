package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zhex/local-bbp/internal/models"
	"gopkg.in/yaml.v3"
	"os"
	"path"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all the available pipelines",
		Run: func(cmd *cobra.Command, args []string) {
			proj := cmd.Flag("project").Value.String()

			data, err := os.ReadFile(path.Join(proj, "bitbucket-pipelines.yml"))
			if err != nil {
				log.Fatal("Error reading bitbucket-pipelines.yml file: ", err)
			}

			var plan models.Plan
			err = yaml.Unmarshal(data, &plan)
			if err != nil {
				log.Fatal("Error parse bitbucket-pipelines.yml file: ", err)
			}

			for i, name := range plan.GetPipelineNames() {
				fmt.Println(i+1, name)
			}
		},
	}

	return cmd
}
