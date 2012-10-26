// gridfsd - GridFS HTTP Server
//
// Copyright (c) 2011 voxelbrain UG (haftungsbeschränkt)
// Authors:
//     Sebastian Friedel <sef@voxelbrain.com>
//     David Lehmann <dtl@voxelbrain.com>
//
// All rights reserved.
package main

import (
	"flag"
	"fmt"
	"http"
	"io"
	"launchpad.net/gobson/bson"
	"launchpad.net/mgo"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
	"sort"
	"template"
)

// flags for command line configuration
var (
	httpAddr        string
	mongoAddr       string
	mongoDatabase   string
	mongoCollection string
	serveIndex      bool
)

// initializing the command line flags
func init() {
	flag.StringVar(&httpAddr, "listen", "localhost:7000", "http address")
	flag.StringVar(&mongoAddr, "dbaddr", "localhost", "database address")
	flag.StringVar(&mongoDatabase, "dbname", "", "database name")
	flag.StringVar(&mongoCollection, "dbfs", "fs", "database collection")
	flag.BoolVar(&serveIndex, "index", false, "serve directory indexes")
	flag.Parse()

	// database name is mandatory
	if mongoDatabase == "" {
		flag.Usage()
		os.Exit(1)
	}
}

// A GridFSServer implements the http.Handler interface and serves files
// from GridFS that start with a specific prefix. It handles GET requests only.
// Remember to call GridFSServer.Close to close the mgo.Mongo session.
type GridFSServer struct {
	prefix  string
	session *mgo.Session
	gfs     *mgo.GridFS
}

// Factory for GridFSServer
func NewGridFSServer(prefix, addr, database, collection string) *GridFSServer {
	session, err := mgo.Mongo(addr)
	if err != nil {
		panic(err)
	}

	db := session.DB(database)
	gfs := db.GridFS(collection)

	server := &GridFSServer{prefix, session, gfs}

	return server
}

// Closes a mgo.Mongo session.
func (srv *GridFSServer) Close() {
	srv.session.Close()
}

// Dispatch requests
func (srv *GridFSServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	method := req.Method
	path := req.URL.Path

	// handle errors by returning a HTTP 500 status code to the client
	defer func() {
		if e := recover(); e != nil {
			trace := debug.Stack()
			msg := fmt.Sprintf("Internal Server Error\n\n%s\n%s", trace, e)
			http.Error(w, msg, http.StatusInternalServerError)
		}
	}()

	if method != "GET" {
		http.Error(w, "Method Not Allowed "+method, http.StatusMethodNotAllowed)
		return
	}

	if strings.HasSuffix(path, "/") && serveIndex {
		srv.ServeIndex(w, req)
		return
	}

	srv.ServeFile(w, req)
}

// Serve existing files from GridFS or responde with HTTP 404 Not Found if
// no file with the requested path can be found.
func (srv *GridFSServer) ServeFile(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path[len(srv.prefix):]
	file, err := srv.gfs.Open(path)

	if err == mgo.NotFound {
		http.NotFound(w, req)
		return
	}

	if err != nil {
		panic(err)
	}

	defer file.Close()

	io.Copy(w, file)
}

// the directory listing template
var parsedIndexTemplate = template.Must(template.New("index").Parse(indexTemplate))

// needed as container for GridFS queries that are used in directory listings
type gfsFile struct {
	Filename string `bson:",omitempty"`
}

// listing context for html template rendering
type directoryIndex struct {
	Path        string
	Files       []string
	Directories sort.StringSlice
}

// Serve a simple directory listing if the `index`-flag is `true`
func (srv *GridFSServer) ServeIndex(w http.ResponseWriter, req *http.Request) {
	var (
		result gfsFile
		q      bson.M
	)

	path := req.URL.Path[len(srv.prefix):]
	ctx := &directoryIndex{req.URL.Path, make([]string, 0), make([]string, 0)}

	if path != "" {
		pattern := "^" + regexp.QuoteMeta(path)
		q = bson.M{"filename": bson.RegEx{pattern, ""}}
	}

	cur := srv.gfs.Files.Find(q).Sort(bson.M{"filename": 1})

	if path != "" {
		if count, err := cur.Count(); err != nil {
			panic(err)
		} else if count == 0 {
			http.NotFound(w, req)
			return
		}
	}

	iter := cur.Iter()

	for iter.Next(&result) {
		resultPath := result.Filename[len(path):]
		parts := strings.Split(resultPath, "/")

		if len(parts) > 1 {
			i := ctx.Directories.Search(parts[0])

			if !(i < len(ctx.Directories) && ctx.Directories[i] == parts[0]) {
				ctx.Directories = append(ctx.Directories, parts[0])
			}
			ctx.Directories.Sort()

		} else {
			ctx.Files = append(ctx.Files, resultPath)
		}
	}

	if iter.Err() != nil {
		panic(iter.Err())
	}

	if err := parsedIndexTemplate.Execute(w, ctx); err != nil {
		panic(err)
	}
}

func main() {
	prefix := "/"
	server := NewGridFSServer(prefix, mongoAddr, mongoDatabase, mongoCollection)
	defer server.Close()

	http.Handle(prefix, server)

	err := http.ListenAndServe(httpAddr, nil)
	if err != nil {
		panic(err)
	}
}
