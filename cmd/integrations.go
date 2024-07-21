package cmd

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zhex/local-bbp/internal/integrations"
	"os"
	"strings"
)

func newIntegrationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integrations",
		Short: "List all the available integrations in the marketplace",
		Run: func(cmd *cobra.Command, args []string) {
			items, err := integrations.Search()
			if err != nil {
				log.Fatalf("Error searching integrations: %s", err)
			}

			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.SetStyle(table.StyleLight)

			t.AppendHeader(table.Row{"Category", "Name", "Repository", "Version", "Tags"})
			for _, item := range items {
				t.AppendRow(table.Row{item.Category, item.Name, item.RepositoryPath, item.Version, strings.Join(item.Tags, ", ")})
				t.AppendSeparator()
			}
			t.AppendFooter(table.Row{"", fmt.Sprintf("Total: %d", len(items))})

			t.SortBy([]table.SortBy{
				{Name: "Category", Mode: table.Asc},
				{Name: "Name", Mode: table.Asc},
			})
			t.SetColumnConfigs([]table.ColumnConfig{
				{Number: 1, AutoMerge: true},
			})

			t.Render()
		},
	}
	return cmd
}
