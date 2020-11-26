package server

import (
	"net/http"
	"os"
)

// A custom FileSystem implementation to plug into http.FileServer which
// doesn't expose directory listings.
type FileSystem struct {
	fs http.FileSystem
}

func (fs FileSystem) Open(path string) (http.File, error) {
	f, err := fs.fs.Open(path)
	if err != nil {
		return nil, err
	}
	s, err := f.Stat()
	if s.IsDir() {
		return nil, os.ErrNotExist
	}
	return f, nil
}
