package ftp

import (
	"io"
	"os"
	"path"

	"github.com/pkg/sftp"
)

type FileInfo interface {
	Name() string
	IsDir() bool
}

type DirEntry struct {
	Path string
	Info FileInfo
}

type FS interface {
	Getwd() (string, error)
	Walk(string, bool) func(func(DirEntry) bool)
	Open(string, int) (io.ReadWriteCloser, error)
	Mkdir(string) error
	Remove(string) error
}

type LocalFS struct {
}

func (fs LocalFS) Getwd() (string, error) {
	return os.Getwd()
}

func (fs LocalFS) Mkdir(p string) error {
	return os.Mkdir(p, 0o770)
}

func (fs LocalFS) Walk(p string, r bool) func(func(DirEntry) bool) {
	return func(yield func(DirEntry) bool) {
		files, err := os.ReadDir(p)

		if err != nil {
			return
		}

		for _, file := range files {
			walkFile := DirEntry{Path: path.Join(p, file.Name()), Info: file}

			if !yield(walkFile) {
				return
			}

			if !r || !file.IsDir() {
				continue
			}

			dirpath := path.Join(p, file.Name())

			for file := range fs.Walk(dirpath, r) {
				if !yield(file) {
					return
				}
			}
		}
	}
}

func (fs LocalFS) Open(filepath string, flag int) (io.ReadWriteCloser, error) {
	return os.OpenFile(filepath, flag, 0o660)
}

func (fs LocalFS) Remove(p string) error {
	return os.RemoveAll(p)
}

type SftpFS struct {
	client *sftp.Client
}

func (fs SftpFS) Getwd() (string, error) {
	return fs.client.Getwd()
}

func (fs SftpFS) Walk(p string, r bool) func(func(DirEntry) bool) {
	return func(yield func(DirEntry) bool) {
		files, err := fs.client.ReadDir(p)

		if err != nil {
			return
		}

		for _, file := range files {
			walkFile := DirEntry{Path: path.Join(p, file.Name()), Info: file}

			if !yield(walkFile) {
				return
			}

			if !r || !file.IsDir() {
				continue
			}

			dirpath := path.Join(p, file.Name())

			for file := range fs.Walk(dirpath, r) {
				if !yield(file) {
					return
				}
			}
		}
	}
}

func (fs SftpFS) Open(p string, flag int) (io.ReadWriteCloser, error) {
	return fs.client.OpenFile(p, flag)
}

func (fs SftpFS) Remove(p string) error {
	return fs.client.RemoveAll(p)
}

func (fs SftpFS) Mkdir(p string) error {
	return fs.client.Mkdir(p)
}
