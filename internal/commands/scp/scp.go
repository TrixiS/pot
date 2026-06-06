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
		Long: `pot scp <connection_id> 'local:hello.txt -> remote:/home/uploaded_hello.txt'
pot scp <connection_id> 'remote:/home/uploaded.txt -> local:downloaded.txt'`,
		Args: cobra.MinimumNArgs(2),
		RunE: runSCP,
	}

	return cmd
}

func runSCP(cmd *cobra.Command, args []string) error {
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

	for i := 1; i < len(args); i++ {
		transferString := args[i]
		transfer, err := parseFileTransfer(transferString)

		if err != nil {
			fmt.Println(transferString, err)
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

var (
	errNoArrow       = errors.New("'->' separator not found")
	errInvalidPrefix = errors.New("must use 'local:' and 'remote:' prefixes on opposite sides")
)

func parseFileTransfer(input string) (transfer FileTransfer, _err error) {
	const (
		localPrefix  = "local:"
		remotePrefix = "remote:"
		arrow        = "->"
	)

	idx := strings.Index(input, arrow)

	if idx == -1 {
		return transfer, errNoArrow
	}

	s1, e1 := trimIndices(input, 0, idx)
	s2, e2 := trimIndices(input, idx+len(arrow), len(input))

	part1 := input[s1:e1]
	part2 := input[s2:e2]

	if strings.HasPrefix(part1, localPrefix) && strings.HasPrefix(part2, remotePrefix) {
		transfer.localPath = part1[len(localPrefix):]
		transfer.remotePath = part2[len(remotePrefix):]
		transfer.direction = DirectionUpload
	} else if strings.HasPrefix(part1, remotePrefix) && strings.HasPrefix(part2, localPrefix) {
		transfer.localPath = part2[len(localPrefix):]
		transfer.remotePath = part1[len(remotePrefix):]
		transfer.direction = DirectionDownload
	} else {
		return transfer, errInvalidPrefix
	}

	transfer.localPath = strings.TrimSpace(transfer.localPath)
	transfer.remotePath = strings.TrimSpace(transfer.remotePath)

	return transfer, nil
}

func trimIndices(s string, start int, end int) (int, int) {
	for start < end && s[start] <= ' ' {
		start++
	}

	for end > start && s[end-1] <= ' ' {
		end--
	}

	return start, end
}
