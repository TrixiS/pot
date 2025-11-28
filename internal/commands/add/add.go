package add

import (
	"fmt"

	"github.com/TrixiS/pot/internal/db/dbconn"
	"github.com/TrixiS/pot/internal/db/models"
	"github.com/TrixiS/pot/internal/kc"
	"github.com/keybase/go-keychain"
	"github.com/spf13/cobra"
)

var (
	name     string
	host     string
	user     string
	port     uint16
	password string
)

func NewCommand() *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new connection",
		RunE:  runAdd,
	}

	addCmd.Flags().StringVarP(&name, "name", "", "", "Name")
	addCmd.Flags().StringVarP(&host, "host", "", "", "Host")
	addCmd.Flags().StringVarP(&user, "user", "", "", "User")
	addCmd.Flags().Uint16VarP(&port, "port", "", 0, "Port")
	addCmd.Flags().StringVarP(&password, "password", "", "", "Password")

	addCmd.MarkFlagRequired("name")
	addCmd.MarkFlagRequired("host")
	addCmd.MarkFlagRequired("user")
	addCmd.MarkFlagRequired("port")
	addCmd.MarkFlagRequired("password")

	return addCmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	item := keychain.NewGenericPassword(
		kc.ServiceName,
		host,
		user,
		[]byte(password),
		kc.AccessGroup,
	)

	if err := keychain.AddItem(item); err != nil {
		return err
	}

	conn := models.Connection{
		Name: name,
		Host: host,
		User: user,
		Port: port,
	}

	db := dbconn.New()
	err := db.Save(&conn)
	db.Close()

	if err != nil {
		return err
	}

	fmt.Printf("Connection %s has been added with ID %d\n", name, conn.ID)
	return nil
}
