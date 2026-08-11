package update

import (
	"fmt"

	"github.com/TrixiS/pot/internal/db/dbconn"
	"github.com/TrixiS/pot/internal/db/models"
	"github.com/TrixiS/pot/internal/kc"
	"github.com/spf13/cobra"
)

var (
	id       int
	name     string
	host     string
	user     string
	port     uint16
	password string
)

func NewCommand() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"u", "upd"},
		Short:   "Update connection configuration",
		RunE:    runUpdate,
	}

	updateCmd.Flags().IntVarP(&id, "id", "", 0, "Connection ID")
	updateCmd.MarkFlagRequired("id")

	updateCmd.Flags().StringVarP(&name, "name", "", "", "Name")
	updateCmd.Flags().StringVarP(&user, "user", "", "", "User")
	updateCmd.Flags().Uint16VarP(&port, "port", "", 0, "Port")
	updateCmd.Flags().StringVarP(&password, "password", "", "", "Password")

	return updateCmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	db := dbconn.New()
	defer db.Close()

	conn := models.Connection{}

	if err := db.One("ID", id, &conn); err != nil {
		return err
	}

	currentPassword, err := kc.GetPassword(conn.Host, conn.User)

	if err != nil {
		return err
	}

	if cmd.Flag("name").Changed {
		conn.Name = name
	}

	if cmd.Flag("host").Changed {
		conn.Host = host
	}

	if cmd.Flag("user").Changed {
		conn.User = user
	}

	if cmd.Flag("port").Changed {
		conn.Port = port
	}

	if cmd.Flag("password").Changed {
		currentPassword = password
	}

	tx, err := db.Begin(true)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := tx.Save(&conn); err != nil {
		return err
	}

	if err := kc.AddPassword(host, user, currentPassword); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("Connection %s (%d) has been updated\n", conn.Name, conn.ID)
	return nil
}
