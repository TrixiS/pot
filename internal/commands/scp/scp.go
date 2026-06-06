package scp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/TrixiS/pot/internal/commands/connect"
	"github.com/TrixiS/pot/internal/db/dbconn"
	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
)

// TODO: progress with status bars for every file
// TODO: flag to create dirs
// TODO: support directory upload/download
// TODO: support globs

type Direction int

const (
	DirectionUpload   Direction = 0
	DirectionDownload Direction = 1
)

type FileTransfer struct {
	localPath  string
	remotePath string
	direction  Direction
}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scp",
		Short: "SCP file transfer",
		Long: `
Uploading
pot scp <connection_id> local:hello.txt remote:/home/uploaded_hello.txt

Downloading
pot scp <connection_id> remote:/home/uploaded.txt local:downloaded.txt`,
		Args: cobra.MinimumNArgs(3),
		RunE: runSCP,
	}

	return cmd
}

func runSCP(cmd *cobra.Command, args []string) error {
	if (len(args)-1)%2 != 0 {
		return errors.New("expected an even number of filepath pairs")
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

	sftpClient, err := sftp.NewClient(sshConn)

	if err != nil {
		return err
	}

	defer sftpClient.Close()

	wg := &sync.WaitGroup{}

	for i := 1; i < len(args)-1; i += 2 {
		transfer, err := parseTransferPair(args[i], args[i+1])

		if err != nil {
			fmt.Println(err)
			continue
		}

		wg.Add(1)

		switch transfer.direction {
		case DirectionUpload:
			go func() {
				fmt.Println("uploading", transfer.localPath)
				err := uploadFile(sftpClient, transfer.localPath, transfer.remotePath)

				if err != nil {
					fmt.Println("failed to upload", transfer.localPath, err)
				}

				wg.Done()
			}()
		case DirectionDownload:
			go func() {
				fmt.Println("downloading", transfer.remotePath)
				err := downloadFile(sftpClient, transfer.remotePath, transfer.localPath)

				if err != nil {
					fmt.Println("failed to download", transfer.remotePath, err)
				}

				wg.Done()
			}()
		default:
			panic(transfer.direction)
		}
	}

	wg.Wait()
	return nil
}

func uploadFile(sftpClient *sftp.Client, localPath string, remotePath string) error {
	localFile, err := os.Open(localPath)

	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}

	defer localFile.Close()

	remoteFile, err := sftpClient.Create(remotePath)

	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}

	defer remoteFile.Close()

	_, err = io.Copy(remoteFile, localFile)

	if err != nil {
		return fmt.Errorf("failed to copy data to remote: %w", err)
	}

	return nil
}

func downloadFile(sftpClient *sftp.Client, remotePath string, localPath string) error {
	remoteFile, err := sftpClient.Open(remotePath)

	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}

	defer remoteFile.Close()

	localFile, err := os.Create(localPath)

	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}

	defer localFile.Close()

	_, err = io.Copy(localFile, remoteFile)

	if err != nil {
		return fmt.Errorf("failed to copy data from remote: %w", err)
	}

	return nil
}

func parseTransferPair(arg1 string, arg2 string) (t FileTransfer, _err error) {
	const (
		localPrefix  = "local:"
		remotePrefix = "remote:"
	)

	if strings.HasPrefix(arg1, remotePrefix) {
		if !strings.HasPrefix(arg2, localPrefix) {
			return t, fmt.Errorf(
				"expected local:filepath after remote for download, got '%s'", arg2,
			)
		}

		t.remotePath = arg1[len(remotePrefix):]
		t.localPath = arg2[len(localPrefix):]
		t.direction = DirectionDownload
		return
	}

	if strings.HasPrefix(arg1, localPrefix) {
		if !strings.HasPrefix(arg2, remotePrefix) {
			return t, fmt.Errorf("expected remote:filepath after local for upload, got '%s'", arg2)
		}

		t.localPath = arg1[len(localPrefix):]
		t.remotePath = arg2[len(remotePrefix):]
		t.direction = DirectionUpload
		return
	}

	return t, fmt.Errorf(
		"expected '%s' or '%s' prefix before filepath, got '%s'",
		localPrefix, remotePrefix, arg1,
	)
}
