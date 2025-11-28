package migrate

import (
	"fmt"

	"github.com/TrixiS/pot/internal/db/dbconn"
	"github.com/TrixiS/pot/internal/db/models"
	"github.com/TrixiS/pot/internal/kc"
	"github.com/asdine/storm/v3"
	"github.com/keybase/go-keychain"
	"github.com/spf13/cobra"
)

// Mantra db connection struct
type Connection struct {
	ID       int    `storm:"id,increment" json:"id"`
	Name     string `storm:"index"        json:"name"`
	Host     string `storm:"index"        json:"host"`
	Port     uint   `                     json:"port"`
	User     string `                     json:"user"`
	Password string `                     json:"password"`
	Args     string `                     json:"args"`
}

func NewCommand() *cobra.Command {
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate from mantra database file",
		RunE:  runMigrate,
		Args:  cobra.ExactArgs(1),
	}

	return migrateCmd
}

func runMigrate(cmd *cobra.Command, args []string) error {
	mantraDB, err := storm.Open(args[0])

	if err != nil {
		return err
	}

	connections := []Connection{}
	err = mantraDB.All(&connections)
	mantraDB.Close()

	if err != nil {
		return err
	}

	db := dbconn.New()

	for _, conn := range connections {
		db.Save(&models.Connection{
			ID:   conn.ID,
			Name: conn.Name,
			Host: conn.Host,
			User: conn.User,
			Port: uint16(conn.Port),
		})

		item := keychain.NewGenericPassword(
			kc.ServiceName,
			conn.Host,
			conn.User,
			[]byte(conn.Password),
			kc.AccessGroup,
		)

		err := keychain.AddItem(item)

		if err != nil {
			fmt.Printf("Error importing %s (%s): %v\n", conn.Name, conn.Host, err)
		} else {
			fmt.Printf("Imported (%s) %s\n", conn.Name, conn.Host)
		}
	}

	db.Close()
	return nil
}
