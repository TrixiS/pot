package dbconn

import (
	"os"
	"path"

	"github.com/asdine/storm/v3"
)

var dbFilepath = path.Join(initPotDir(), "pot.db")

func initPotDir() string {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		panic(err)
	}

	potDir := path.Join(homeDir, ".pot")

	if err := os.MkdirAll(potDir, 0o770); err != nil {
		panic(err)
	}

	return potDir
}

func New() *storm.DB {
	db, err := storm.Open(dbFilepath)

	if err != nil {
		panic(err)
	}

	return db
}
