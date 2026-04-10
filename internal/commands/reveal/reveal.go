package reveal

import (
	"fmt"

	"github.com/TrixiS/pot/internal/commands/connect"
	"github.com/TrixiS/pot/internal/db/dbconn"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reveal",
		Short: "Reveal connection password",
		Args:  cobra.ExactArgs(1),
		RunE:  runReveal,
	}

	return cmd
}

func runReveal(cmd *cobra.Command, args []string) error {
	db := dbconn.New()
	conn, err := connect.GetConnectionByIDString(db, args[0])
	db.Close()

	if err != nil {
		return err
	}

	password, err := connect.GetConnectionPassword(conn)

	if err != nil {
		return err
	}

	fmt.Println(password)
	return nil
}
