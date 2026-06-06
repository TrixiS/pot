package run

import (
	"fmt"
	"os"
	"strings"

	"github.com/TrixiS/pot/internal/commands/connect"
	"github.com/TrixiS/pot/internal/db/dbconn"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run",
		Short:              "Run a command on remote host",
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		RunE:               runRun,
	}

	return cmd
}

func runRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("connection id or name is required")
	}

	db := dbconn.New()
	conn, err := connect.GetConnectionByIDString(db, args[0])
	db.Close()

	if err != nil {
		return err
	}

	sshConn, err := connect.DialSSH(conn)

	if err != nil {
		return err
	}

	defer sshConn.Close()

	session, err := sshConn.NewSession()

	if err != nil {
		return err
	}

	defer session.Close()

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	return session.Run(strings.Join(args[1:], " "))
}
