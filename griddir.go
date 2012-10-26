package main

import (
	"labix.org/v2/mgo"
	"net/http"
	"path/filepath"
)

type GridDir struct {
	PathPrefix string
	*mgo.GridFS
}

func (gd *GridDir) Open(name string) (http.File, error) {
	levelLogf(LOG_INFO, "Request for %s", name)
	filename := filepath.Join(gd.PathPrefix, name)
	f, err := gd.GridFS.Open(filename)
	if err != nil {
		levelLogf(LOG_ERROR, "File %s not found: %s", name, err)
		return nil, err
	}

	gf := &GridFile{f}
	return gf, nil
}
