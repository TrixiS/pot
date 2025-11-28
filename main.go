package main

import (
	"github.com/TrixiS/pot/internal/commands/add"
	"github.com/TrixiS/pot/internal/commands/connect"
	"github.com/TrixiS/pot/internal/commands/list"
	"github.com/TrixiS/pot/internal/commands/migrate"
	"github.com/TrixiS/pot/internal/commands/remove"
	"github.com/TrixiS/pot/internal/commands/update"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "pot",
		Short: "SSH tool",
	}

	rootCmd.AddCommand(add.NewCommand())
	rootCmd.AddCommand(remove.NewCommand())
	rootCmd.AddCommand(list.NewCommand())
	rootCmd.AddCommand(update.NewCommand())
	rootCmd.AddCommand(connect.NewCommand())
	rootCmd.AddCommand(migrate.NewCommand())

	rootCmd.Execute()
}
