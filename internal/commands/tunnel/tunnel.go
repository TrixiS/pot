package tunnel

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"

	"github.com/TrixiS/pot/internal/commands/connect"
	"github.com/TrixiS/pot/internal/db/dbconn"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Make SSH tunnel",
		Args:  cobra.ExactArgs(3),
		RunE:  runTunnel,
	}

	return cmd
}

func runTunnel(cmd *cobra.Command, args []string) error {
	db := dbconn.New()
	conn, err := connect.GetConnectionByIDString(db, args[0])
	db.Close()

	if err != nil {
		return err
	}

	remotePort, err := strconv.Atoi(args[1])

	if err != nil {
		return fmt.Errorf("remote port parsing error: %w", err)
	}

	localPort, err := strconv.Atoi(args[2])

	if err != nil {
		return fmt.Errorf("local port parsing error: %w", err)
	}

	sshConn, err := connect.DialSSH(conn)

	if err != nil {
		return err
	}

	lis, err := sshConn.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", remotePort))

	if err != nil {
		return err
	}

	defer lis.Close()

	localAddress := fmt.Sprintf("127.0.0.1:%d", localPort)

	slog.Info("listening connections", "remotePort", remotePort, "localPort", localPort)

	for {
		remoteConn, err := lis.Accept()

		if err != nil {
			slog.Error("remote", "err", err)
			continue
		}

		localConn, err := net.Dial("tcp", localAddress)

		if err != nil {
			slog.Error("local", "err", err)
			remoteConn.Close()
			continue
		}

		go func() {
			done := make(chan struct{}, 2)

			go func() {
				io.Copy(remoteConn, localConn)
				done <- struct{}{}
			}()

			go func() {
				io.Copy(localConn, remoteConn)
				done <- struct{}{}
			}()

			<-done
			remoteConn.Close()
			localConn.Close()
		}()
	}
}
