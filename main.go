// gridfsd - GridFS HTTP Server
//
// Copyright (c) 2011-2012 voxelbrain UG (haftungsbeschränkt)
// Authors:
//     Sebastian Friedel <sef@voxelbrain.com>
//     David Lehmann <dtl@voxelbrain.com>
//     Alexander Surma <asu@voxelbrain.com>
//
// All rights reserved.
package main

import (
	"github.com/voxelbrain/goptions"
	"labix.org/v2/mgo"
	"log"
	"net"
	"net/http"
	"net/url"
)

var options = struct {
	MongoURL         *url.URL     `goptions:"-m, --mongodb, description='MongoDB URL to connect to (example: mongodb://localhost/db)', obligatory"`
	CollectionPrefix string       `goptions:"-c, --collection, description='Prefix of GridFS collection name'"`
	PathPrefix       string       `goptions:"-p, --prefix, description='Prefix applied to all request paths'"`
	Address          *net.TCPAddr `goptions:"-a, --address, description='Address to bind webserver to'"`
	Verbosity        []bool       `goptions:"-v, --verbose, description='Increase verbosity'"`
	goptions.Help    `goptions:"-h, --help, description='Show this help'"`
}{
	CollectionPrefix: "fs",
	Address: &net.TCPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: 8080,
	},
}

func init() {
	goptions.ParseAndFail(&options)
}

func main() {
	session, err := mgo.Dial(options.MongoURL.String())
	if err != nil {
		log.Fatalf("Could not connect to mongodb: %s", err)
	}

	db := session.DB("") // Use DB name from the URL
	gd := &GridDir{
		PathPrefix: options.PathPrefix,
		GridFS:     db.GridFS(options.CollectionPrefix),
	}
	levelLogf(LOG_INFO, "Starting server...")
	http.Handle("/", http.FileServer(gd))

	log.Fatalf("ListenAndServe: %s", http.ListenAndServe(options.Address.String(), nil))
}
