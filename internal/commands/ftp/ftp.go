package ftp

import (
	"io"
	"os"
	"path"
	"sync"

	"github.com/TrixiS/pot/internal/commands/connect"
	"github.com/TrixiS/pot/internal/db/dbconn"
	"github.com/gdamore/tcell/v2"
	"github.com/pkg/sftp"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

const (
	MoveRune    = 'm'
	DeleteRune  = 'd'
	RefreshRune = 'r'
	BackRune    = 'b'
)

func NewCommand() *cobra.Command {
	ftpCmd := &cobra.Command{
		Use:   "ftp",
		Short: "SFTP file manager",
		Args:  cobra.ExactArgs(1),
		RunE:  runFtp,
	}

	return ftpCmd
}

type FSView struct {
	listView  *tview.List
	pathInput *tview.InputField
	fs        FS
	cwd       string
	entries   []DirEntry
}

func (f *FSView) enterDir(dirpath string) {
	f.cwd = dirpath
	f.entries = nil

	for entry := range f.fs.Walk(f.cwd, false) {
		f.entries = append(f.entries, entry)
	}
}

func (f FSView) update() {
	idx := f.listView.GetCurrentItem()
	f.listView.Clear()

	for _, entry := range f.entries {
		f.listView.AddItem(entry.Info.Name(), "", 0, nil)
	}

	if idx < f.listView.GetItemCount() {
		f.listView.SetCurrentItem(idx)
	}

	f.pathInput.SetText(f.cwd)
}

func newFSView(fs FS) *FSView {
	cwd, err := fs.Getwd()

	if err != nil {
		panic(err)
	}

	listView := tview.NewList().ShowSecondaryText(false)

	pathInput := tview.NewInputField()
	pathInput.SetFieldBackgroundColor(tcell.ColorNone)

	view := &FSView{listView: listView, pathInput: pathInput, fs: fs, cwd: cwd}

	pathInput.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}

		dirpath := pathInput.GetText()
		info, err := view.fs.Stat(dirpath)

		if err != nil || !info.IsDir() {
			return
		}

		view.enterDir(dirpath)
		view.update()
	})

	listView.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		entry := view.entries[i]

		if entry.Info.IsDir() {
			view.enterDir(path.Join(view.cwd, entry.Info.Name()))
			view.update()
		}
	})

	view.enterDir(cwd)
	view.update()
	return view
}

// TODO: split status bar with progress bar and error message text view

func makeFileViewInputCapture(
	currentFSView *FSView,
	otherFSView *FSView,
) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case BackRune:
			currentFSView.enterDir(path.Join(currentFSView.cwd, ".."))
			currentFSView.update()
		case MoveRune:
			entry := currentFSView.entries[currentFSView.listView.GetCurrentItem()]

			if entry.Info.IsDir() {
				basepath := entry.Path[len(currentFSView.cwd):]
				dirpath := path.Join(otherFSView.cwd, basepath)

				if err := otherFSView.fs.Mkdir(dirpath); err != nil {
					panic(err)
				}

				wg := sync.WaitGroup{}

				for child := range currentFSView.fs.Walk(entry.Path, true) {
					basepath := child.Path[len(currentFSView.cwd):]
					filepath := path.Join(otherFSView.cwd, basepath)

					if child.Info.IsDir() {
						if err := otherFSView.fs.Mkdir(filepath); err != nil {
							panic(err)
						}

						continue
					}

					wg.Add(1)

					go func() {
						if err := copyFile(currentFSView.fs, otherFSView.fs, child.Path, filepath); err != nil {
							panic(err)
						}

						wg.Done()
					}()
				}

				wg.Wait()
			} else {
				basepath := entry.Path[len(currentFSView.cwd):]
				filepath := path.Join(otherFSView.cwd, basepath)

				if err := copyFile(currentFSView.fs, otherFSView.fs, entry.Path, filepath); err != nil {
					panic(err)
				}
			}

			otherFSView.enterDir(otherFSView.cwd)
			otherFSView.update()
		case DeleteRune:
			itemIdx := currentFSView.listView.GetCurrentItem()
			entry := currentFSView.entries[itemIdx]
			currentFSView.fs.Remove(path.Join(currentFSView.cwd, entry.Info.Name()))
			currentFSView.enterDir(currentFSView.cwd)
			currentFSView.update()
		case RefreshRune:
			currentFSView.enterDir(currentFSView.cwd)
			currentFSView.update()
		default:
			return event
		}

		return nil
	}
}

func runFtp(cmd *cobra.Command, args []string) error {
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

	remoteFSView := newFSView(SftpFS{client: sftpClient})
	localFSView := newFSView(LocalFS{})

	remoteFSView.listView.SetInputCapture(
		makeFileViewInputCapture(remoteFSView, localFSView),
	)

	localFSView.listView.SetInputCapture(
		makeFileViewInputCapture(localFSView, remoteFSView),
	)

	grid := tview.NewGrid().
		SetRows(1, 2).
		SetBorders(true).
		AddItem(remoteFSView.pathInput, 1, 0, 1, 1, 0, 0, false).
		AddItem(localFSView.pathInput, 1, 1, 1, 1, 0, 0, false).
		AddItem(remoteFSView.listView, 2, 0, 1, 1, 0, 0, true).
		AddItem(localFSView.listView, 2, 1, 1, 1, 0, 0, false)

	app := tview.NewApplication().SetRoot(grid, true).SetFocus(grid).EnableMouse(true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRight:
			app.SetFocus(localFSView.listView)
		case tcell.KeyLeft:
			app.SetFocus(remoteFSView.listView)
		}

		return event
	})

	return app.Run()
}

func copyFile(from FS, to FS, fromFilepath string, toFilepath string) error {
	fromFile, err := from.Open(fromFilepath, os.O_RDONLY)

	if err != nil {
		return err
	}

	defer fromFile.Close()

	toFile, err := to.Open(toFilepath, os.O_CREATE|os.O_WRONLY)

	if err != nil {
		return err
	}

	_, err = io.Copy(toFile, fromFile)
	toFile.Close()
	return err
}
