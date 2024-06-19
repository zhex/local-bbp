package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func CreateRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "bbp",
		Short:   "bbp is a simple CI/CD tool",
		Long:    `bbp is a simple CI/CD tool that reads a .ci.yaml file and runs the steps defined in it.`,
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ci is a simple CI/CD tool")
		},
	}
	return rootCmd
}
