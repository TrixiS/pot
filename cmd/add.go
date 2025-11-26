package cmd

import (
	"fmt"

	"github.com/TrixiS/pot/internal/db/dbconn"
	"github.com/TrixiS/pot/internal/db/models"
	"github.com/asdine/storm/v3"
	"github.com/keybase/go-keychain"
	"github.com/spf13/cobra"
)

const (
	PotServiceName = "pot"
	PotAccessGroup = "com.trixis.pot"
)

var (
	name     string
	host     string
	user     string
	port     uint16
	password string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new connection",
	RunE:  runAdd,
}

func runAdd(cmd *cobra.Command, args []string) error {
	item := keychain.NewGenericPassword(
		PotServiceName,
		host,
		user,
		[]byte(password),
		PotAccessGroup,
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

	err := dbconn.With(func(db *storm.DB) error {
		return db.Save(&conn)
	})

	if err != nil {
		return err
	}

	fmt.Printf("Connection %s has been added with ID %d\n", name, conn.ID)
	return nil
}

func init() {
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

	rootCmd.AddCommand(addCmd)
}
