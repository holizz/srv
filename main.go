package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("p", 8000, "port to listen on")
	dir := flag.String("d", ".", "directory to serve")
	flag.Parse()

	http.Handle("/", NewSrvServer(http.Dir(*dir)))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

type File struct {
	Path string `json:"path"`
}

type SrvServer struct {
	dir        http.Dir
	fileServer http.Handler
	html       string
	js         string
}

func NewSrvServer(dir http.Dir) SrvServer {
	html, err := ioutil.ReadFile("assets/index.html")
	if err != nil {
		panic(err)
	}

	js, err := ioutil.ReadFile("assets/app.js")
	if err != nil {
		panic(err)
	}

	s := SrvServer{
		dir:        dir,
		fileServer: http.FileServer(dir),
		html:       string(html),
		js:         string(js),
	}

	return s
}

func (s SrvServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/_srv/api" {
		paths := r.URL.Query()["path"]
		if len(paths) > 0 {
			file, err := s.dir.Open(paths[0])
			if err != nil {
				panic(err)
			}
			defer file.Close()

			info, err := file.Stat()
			if err != nil {
				panic(err)
			}

			if !info.IsDir() {
				panic(fmt.Errorf("oh no"))
			}

			files, err := file.Readdir(999) //TODO: 999 is too small
			if err != nil {
				panic(err)
			}

			outputFiles := []File{}
			for _, f := range files {
				outputFiles = append(outputFiles, File{
					Path: f.Name(),
				})
			}

			bytes, err := json.Marshal(outputFiles)
			if err != nil {
				panic(err)
			}

			w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
			io.WriteString(w, string(bytes))
			return
		}
	}

	if r.URL.Path == "/_srv/app.js" {
		w.Header()["Content-Type"] = []string{"application/javascript; charset=utf-8"}
		io.WriteString(w, s.js)
		return
	}

	file, err := s.dir.Open(r.URL.Path)
	if err != nil {
		s.fileServer.ServeHTTP(w, r)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		s.fileServer.ServeHTTP(w, r)
		return
	}

	if info.IsDir() {
		w.Header()["Content-Type"] = []string{"text/html; charset=utf-8"}
		io.WriteString(w, s.html)
		return
	} else {
		s.fileServer.ServeHTTP(w, r)
		return
	}
}
