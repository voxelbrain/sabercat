package sabercat

import (
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"path/filepath"
)

type GridDir struct {
	PathPrefix string
	*mgo.GridFS
}

func (gd *GridDir) Open(name string) (http.File, error) {
	filename := filepath.Join(gd.PathPrefix, name)
	f, err := gd.GridFS.Open(filename)
	if err != nil {
		return nil, err
	}

	gf := &GridFile{f}
	return gf, nil
}
