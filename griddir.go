/*
Package sabercat implements net/http.FileSystem to serve
contents directly from MongoDB's GridFS.

To serve all files from a database containing a GridFS
called `fs` via http:

	session, err := mgo.Dial("mongodb://localhost/database")
	if err != nil {
		log.Fatalf("Could not connect to mongodb: %s", err)
	}
	gfs := session.DB("").GridFS("fs")
	http.Handle("/", http.FileServer(sabercat.GridDir{gfs})

Directory listing has not been implemented.
*/
package sabercat

import (
	"labix.org/v2/mgo"
	"log"
	"net/http"
)

const (
	VERSION = "1.4.2"
)

type GridDir struct {
	GridFS *mgo.GridFS
	// If true, the leading slash will be stripped
	StripSlash bool
}

func (gd GridDir) Open(filename string) (http.File, error) {
	if gd.StripSlash && filename[0] == '/' {
		filename = filename[1:]
	}
	f, err := gd.GridFS.Open(filename)
	if err != nil && err != mgo.ErrNotFound {
		// Check if connection is alive
		pingErr := gd.GridFS.Files.Database.Session.Ping()
		if pingErr != nil {
			log.Printf("Refreshing connection...")
			gd.GridFS.Files.Database.Session.Refresh()
			f, err = gd.GridFS.Open(filename)
		} else {
			log.Printf("Unknown error: %s", err)
		}
	}
	return &gridFile{f}, err
}
