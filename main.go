package main

import (
	"io"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"github.com/robryk/tarstore/tarfs"
	"github.com/robryk/tarstore/ddar"
)

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: %s archive_name
where archive_name is a ddar archive name (not necessarily existent)
`, os.Args[0])
	os.Exit(1)
}

var ddarArchive ddar.Archive

func handleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	idx := strings.Index(path, "/")
	var name string
	if idx == -1 {
		name = path
	} else {
		name = path[:idx]
	}
	switch r.Method {
	case "HEAD":
		fallthrough
	case "GET":
		fs := tarfs.New(func() (io.ReadCloser, error) { return ddarArchive.Get(name) })
		http.StripPrefix(fmt.Sprintf("/%s/", name), http.FileServer(fs)).ServeHTTP(w, r)
	case "PUT":
		if idx != -1 {
			http.NotFound(w, r)
			return
		}
		err := ddarArchive.Add(name, r.Body)
		if err != nil {
			log.Printf("error adding %s to archive: %v", name, err)
			http.Error(w, "error storing tarball", http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		usage()
	}
	ddarArchive = ddar.Archive(os.Args[1])
	http.HandleFunc("/", handleRequest)
	panic(http.ListenAndServe(":8888", nil))
}
