package update

import (
	"fmt"

	"github.com/TrixiS/pot/internal/db/dbconn"
	"github.com/TrixiS/pot/internal/db/models"
	"github.com/TrixiS/pot/internal/kc"
	"github.com/keybase/go-keychain"
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
	updateCmd.Flags().StringVarP(&host, "host", "", "", "Host")
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

	queryItem := keychain.NewItem()
	queryItem.SetSecClass(keychain.SecClassGenericPassword)
	queryItem.SetService(kc.ServiceName)
	queryItem.SetAccount(conn.Host)
	queryItem.SetLabel(conn.User)
	queryItem.SetAccessGroup(kc.AccessGroup)
	queryItem.SetReturnData(true)

	updateItem := keychain.NewItem()
	updateItem.SetSecClass(keychain.SecClassGenericPassword)
	updateItem.SetService(kc.ServiceName)
	updateItem.SetAccount(conn.Host)
	updateItem.SetLabel(conn.User)
	updateItem.SetAccessGroup(kc.AccessGroup)

	if cmd.Flag("name").Changed {
		conn.Name = name
	}

	if cmd.Flag("host").Changed {
		conn.Host = host
		updateItem.SetAccount(host)
	}

	if cmd.Flag("user").Changed {
		conn.User = user
		updateItem.SetLabel(user)
	}

	if cmd.Flag("port").Changed {
		conn.Port = port
	}

	if cmd.Flag("password").Changed {
		updateItem.SetData([]byte(password))
	}

	tx, err := db.Begin(true)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := tx.Save(&conn); err != nil {
		return err
	}

	if err := keychain.UpdateItem(queryItem, updateItem); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("Connection %s (%d) has been updated\n", conn.Name, conn.ID)
	return nil
}
