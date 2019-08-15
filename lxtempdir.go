// Package lxtempdir provides safe temporary directories for Linux.
package lxtempdir

import (
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"os"
)

// TempDir represents a temporary directory that you have access to. systemd-tmpfiles will
// *not* clean up the directory until the program dies or Close is called.
type TempDir struct {
	// The path to the temporary directory.
	Path string
	fd 	 int
}

// Create creates a new temporary directory. The arguments have identical meaning to
// ioutil.TempDir.
func Create(dir, prefix string) (*TempDir, error) {
	path, err := ioutil.TempDir(dir, prefix)
	if err != nil {
		return nil, err
	}

	fd := -1

	defer func() {
		if err != nil {
			if fd != -1 {
				unix.Close(fd)
			}

			os.Remove(path)
		}
	}()

	fd, err = unix.Open(path, unix.O_DIRECTORY, 0)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open temporary directory")
	}

	if err := unix.Flock(fd, unix.LOCK_SH); err != nil {
		return nil, errors.Wrap(err, "failed to lock temporary directory")
	}

	return &TempDir{
		Path: path,
		fd: fd,
	}, nil
}

// Close relinquishes access to tempdir. After this function is called, systemd-tmpfiles is
// free to clean up the temporary directory, even if your program is still running.
//
// This does not remove the directory and its contents; use os.RemoveAll first to do that.
func (tempdir *TempDir) Close() error {
	err := unix.Close(tempdir.fd)
	if err != nil {
		return errors.Wrap(err, "failed to close temporary directory")
	}
	return nil
}
