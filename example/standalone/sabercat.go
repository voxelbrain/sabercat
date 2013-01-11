package main

import (
	"github.com/voxelbrain/goptions"
	"github.com/voxelbrain/sabercat"
	"labix.org/v2/mgo"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

var options = struct {
	MongoURL         *url.URL      `goptions:"-m, --mongodb, description='MongoDB URL to connect to (example: mongodb://localhost/db)', obligatory"`
	CollectionPrefix string        `goptions:"-c, --collection, description='Prefix of GridFS collection name'"`
	CacheTime        time.Duration `goptions:"-t, --cache-time, description='Duration to keep files in cache'"`
	GridfsPrefix     string        `goptions:"-g, --gridfs-prefix, description='Prefix applied before requesting GridFS'"`
	HttpPrefix       string        `goptions:"-p, --http-prefix, description='Prefix under which GridFS is served'"`
	StripSlash       bool          `goptions:"-s, --strip-slash, description='Strip leading slash before requesting GridFS'"`
	Address          *net.TCPAddr  `goptions:"-a, --address, description='Address to bind webserver to'"`
	Help             goptions.Help `goptions:"-h, --help, description='Show this help'"`
}{
	CollectionPrefix: "fs",
	HttpPrefix:       "/",
	Address: &net.TCPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: 8080,
	},
}

func init() {
	goptions.ParseAndFail(&options)
}

func AddPrefix(prefix string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = prefix + r.URL.Path
		handler.ServeHTTP(w, r)
	})
}

func main() {
	session, err := mgo.Dial(options.MongoURL.String())
	if err != nil {
		log.Fatalf("Could not connect to mongodb: %s", err)
	}

	gfs := session.DB("").GridFS(options.CollectionPrefix)

	log.Printf("Starting server...")
	http.Handle("/",
		Cache(options.CacheTime,
			http.StripPrefix(options.HttpPrefix,
				AddPrefix(options.GridfsPrefix,
					http.FileServer(sabercat.GridDir{
						GridFS:     gfs,
						StripSlash: options.StripSlash,
					})))))
	log.Fatalf("ListenAndServe: %s", http.ListenAndServe(options.Address.String(), nil))
}
