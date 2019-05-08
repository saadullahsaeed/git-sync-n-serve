package syncserve

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

func NewNeuteredFileSystem(fs http.FileSystem, logger *log.Entry) FileSystem {
	ns := FileSystem{fs: fs, Logger: logger}
	// we don't want to server .dot directories
	ns.dotRegex = regexp.MustCompile(`^\.(.*)`)
	return ns
}

type FileSystem struct {
	fs       http.FileSystem
	dotRegex *regexp.Regexp
	Logger   *log.Entry
}

// Open ...
func (nfs FileSystem) Open(path string) (http.File, error) {
	if nfs.dotRegex != nil {
		if nfs.dotRegex.MatchString(path[1:]) {
			return nil, errors.New("not found")
		}
	}

	f, err := nfs.fs.Open(path)
	if err != nil {
		nfs.Logger.Error(err)
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := fmt.Sprintf("%s/%s", strings.TrimSuffix(path, "/"), "index.html")
		_, err := nfs.fs.Open(index)
		if err != nil {
			nfs.Logger.Error(err)
			return nil, err
		}
	}
	return f, nil
}
