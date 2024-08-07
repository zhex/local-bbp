package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zhex/local-bbp/internal/common"
	"github.com/zhex/local-bbp/internal/config"
	"github.com/zhex/local-bbp/internal/docker"
	"github.com/zhex/local-bbp/internal/parser"
	"github.com/zhex/local-bbp/internal/runner"
	"os"
	"path/filepath"
	"strings"
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a pipeline",
		Run: func(cmd *cobra.Command, args []string) {
			proj := cmd.Flag("project").Value.String()
			name := cmd.Flag("name").Value.String()
			targetBranch := cmd.Flag("target-branch").Value.String()
			verbose, _ := cmd.Flags().GetBool("verbose")

			if verbose {
				log.SetLevel(log.DebugLevel)
			}

			if !strings.HasPrefix(name, "pr/") {
				targetBranch = ""
			}

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

			arch := common.GetArch()
			if !common.Contains(docker.SupportedArchitectures, arch) {
				log.Fatalf("Unsupported architecture: %s", arch)
			}

			dockerPath := filepath.Join(c.ToolDir, arch, "docker/docker")
			if !common.IsFileExists(dockerPath) {
				log.Info("Downloading linux docker cli binary")
				if err = docker.DownloadDockerCliBinary(c.DockerVersion, c.ToolDir); err != nil {
					log.Fatalf("Error downloading docker cli binary: %s", err)
				}
			}

			fullPath, _ := filepath.Abs(proj)
			if !filepath.IsAbs(c.OutputDir) {
				c.OutputDir = filepath.Join(fullPath, c.OutputDir)
			}
			r := runner.New(fullPath, c, secrets)
			r.Run(name, targetBranch)
		},
	}

	cmd.Flags().StringP("name", "n", "default", "Name of the workflow to run")
	cmd.Flags().StringP("secrets-file", "s", "", "Path to the secrets file")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")
	cmd.Flags().StringP("target-branch", "t", "main", "Target branch for a pull request pipeline. Default is 'main'")

	return cmd
}
