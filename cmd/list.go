package cmd

import (
	"fmt"
	"os"

	"github.com/TrixiS/pot/internal/db/dbconn"
	"github.com/TrixiS/pot/internal/db/models"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List connections",
	RunE:    runList,
}

func runList(cmd *cobra.Command, args []string) error {
	db := dbconn.New()
	defer db.Close()

	var connections []models.Connection

	if err := db.All(&connections); err != nil {
		return err
	}

	if len(connections) == 0 {
		fmt.Println("No connections added")
		return nil
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Name", "Host", "Port", "User"})
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = false
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:   "ID",
			Colors: text.Colors{text.Bold, text.FgHiCyan},
		},
	})

	for _, conn := range connections {
		t.AppendRow(table.Row{conn.ID, conn.Name, conn.Host, conn.Port, conn.User})
	}

	t.Render()
	return nil
}

func init() {
	rootCmd.AddCommand(listCmd)
}
