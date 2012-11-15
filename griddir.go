/*
Package sabercat implements net/http.FileSystem to serve
contents directly from MongoDB's GridFS.

To serve all files from a database containing a GridFS
called `fs` via http:

	session, err := mgo.Dial("mongodb://localhost/database")
	if err != nil {
		log.Fatalf("Could not connect to mongodb: %s", err)
	}
	db := session.DB("") // Use DB name from the URL
	http.Handle("/", http.FileServer(&sabercat.GridDir{
		PathPrefix: "",
		GridFS:     db.GridFS("fs"),
	}

Directory listing has not been implemented.
*/
package sabercat

import (
	"labix.org/v2/mgo"
	"net/http"
	"path/filepath"
)

const (
	VERSION = "1.1.1"
)

type GridDir struct {
	PathPrefix string
	GridFS     *mgo.GridFS
}

func (gd *GridDir) Open(name string) (http.File, error) {
	filename := filepath.Join(gd.PathPrefix, name)
	f, err := gd.GridFS.Open(filename)
	return &gridFile{f}, err
}
