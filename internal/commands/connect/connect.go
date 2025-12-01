package connect

import (
	"fmt"
	"os"
	"strconv"

	"github.com/TrixiS/pot/internal/db/dbconn"
	"github.com/TrixiS/pot/internal/db/models"
	"github.com/TrixiS/pot/internal/kc"
	"github.com/asdine/storm/v3"
	"github.com/keybase/go-keychain"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var termModes = ssh.TerminalModes{
	ssh.ECHO:          1,
	ssh.ICANON:        1,
	ssh.OPOST:         1,
	ssh.TTY_OP_ISPEED: 14400,
	ssh.TTY_OP_OSPEED: 14400,
}

func NewCommand() *cobra.Command {
	connectCmd := &cobra.Command{
		Use:     "connect",
		Aliases: []string{"conn", "c"},
		Short:   "Make an SSH connection",
		Args:    cobra.ExactArgs(1),
		RunE:    runConnect,
	}

	return connectCmd
}

func runConnect(cmd *cobra.Command, args []string) error {
	db := dbconn.New()
	conn, err := GetConnectionByIDString(db, args[0])
	db.Close()

	if err != nil {
		return err
	}

	sshConn, err := DialSSH(conn)

	if err != nil {
		return err
	}

	defer sshConn.Close()

	session, err := sshConn.NewSession()

	if err != nil {
		return err
	}

	defer session.Close()

	fd := int(os.Stdin.Fd())
	termState, err := term.MakeRaw(fd)

	if err != nil {
		return err
	}

	defer term.Restore(fd, termState)

	termW, termH, err := term.GetSize(fd)

	if err != nil {
		return err
	}

	if err := session.RequestPty("xterm-256color", termH, termW, termModes); err != nil {
		return err
	}

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.Shell(); err != nil {
		return err
	}

	return session.Wait()
}

func GetConnectionByIDString(db *storm.DB, idString string) (*models.Connection, error) {
	conn := models.Connection{}

	if id, err := strconv.Atoi(idString); err == nil {
		return &conn, db.One("ID", id, &conn)
	}

	return &conn, db.One("Name", idString, &conn)
}

func DialSSH(conn *models.Connection) (*ssh.Client, error) {
	passwordBytes, err := keychain.GetGenericPassword(
		kc.ServiceName,
		conn.Host,
		conn.User,
		kc.AccessGroup,
	)

	if err != nil {
		return nil, err
	}

	config := ssh.ClientConfig{
		User: conn.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(string(passwordBytes)),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return ssh.Dial("tcp", fmt.Sprintf("%s:%d", conn.Host, conn.Port), &config)
}
