package remove

import (
	"fmt"
	"strconv"

	"github.com/TrixiS/pot/internal/db/dbconn"
	"github.com/TrixiS/pot/internal/db/models"
	"github.com/TrixiS/pot/internal/kc"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	removeCmd := &cobra.Command{
		Use:     "remove",
		Aliases: []string{"rm"},
		Short:   "Remove a connection",
		Args:    cobra.ExactArgs(1),
		RunE:    runRemove,
	}

	return removeCmd
}

func runRemove(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])

	if err != nil {
		return err
	}

	db := dbconn.New()
	defer db.Close()

	tx, err := db.Begin(true)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	conn := models.Connection{}

	if err := db.One("ID", id, &conn); err != nil {
		return err
	}

	if err := tx.DeleteStruct(&conn); err != nil {
		return err
	}

	if err := kc.DeletePassword(conn.Host); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("Connection %s (%d) has been removed\n", conn.Name, conn.ID)
	return nil
}
