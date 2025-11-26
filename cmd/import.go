package cmd

import (
	"fmt"

	"github.com/TrixiS/pot/internal/db/dbconn"
	"github.com/TrixiS/pot/internal/db/models"
	"github.com/asdine/storm/v3"
	"github.com/keybase/go-keychain"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import mantra connection database",
	RunE:  runImport,
	Args:  cobra.ExactArgs(1),
}

type Connection struct {
	ID       int    `storm:"id,increment" json:"id"`
	Name     string `storm:"index"        json:"name"`
	Host     string `storm:"index"        json:"host"`
	Port     uint   `                     json:"port"`
	User     string `                     json:"user"`
	Password string `                     json:"password"`
	Args     string `                     json:"args"`
}

func runImport(cmd *cobra.Command, args []string) error {
	mantraDBFilepath := args[0]

	mantraDB, err := storm.Open(mantraDBFilepath)

	if err != nil {
		return err
	}

	defer mantraDB.Close()

	connections := []Connection{}

	if err := mantraDB.All(&connections); err != nil {
		return err
	}

	db := dbconn.New()
	defer db.Close()

	for _, conn := range connections {
		db.Save(&models.Connection{
			ID:   conn.ID,
			Name: conn.Name,
			Host: conn.Host,
			User: conn.User,
			Port: uint16(conn.Port),
		})

		item := keychain.NewGenericPassword(
			PotServiceName,
			conn.Host,
			conn.User,
			[]byte(conn.Password),
			PotAccessGroup,
		)

		err := keychain.AddItem(item)

		if err != nil {
			fmt.Printf("Error importing %s (%s): %v\n", conn.Name, conn.Host, err)
		} else {
			fmt.Printf("Imported (%s) %s\n", conn.Name, conn.Host)
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(importCmd)
}
