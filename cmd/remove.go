package cmd

import (
	"fmt"

	"github.com/TrixiS/pot/internal/db/dbconn"
	"github.com/TrixiS/pot/internal/db/models"
	"github.com/keybase/go-keychain"
	"github.com/spf13/cobra"
)

var (
	id int
)

var removeCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm"},
	Short:   "Remove a connection",
	RunE:    runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
	conn := models.Connection{}

	db := dbconn.New()
	defer db.Close()

	if err := db.One("ID", id, &conn); err != nil {
		return err
	}

	if err := db.DeleteStruct(&conn); err != nil {
		return err
	}

	if err := keychain.DeleteGenericPasswordItem(PotServiceName, conn.Host); err != nil {
		return err
	}

	fmt.Printf("Connection %s (%d) has been removed\n", conn.Name, conn.ID)
	return nil
}

func init() {
	removeCmd.Flags().IntVarP(&id, "id", "", 0, "Connection ID")
	removeCmd.MarkFlagRequired("id")
	rootCmd.AddCommand(removeCmd)
}
