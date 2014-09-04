package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"

	"code.google.com/p/go.net/websocket"

	"github.com/GeertJohan/go.rice"
	"gopkg.in/fsnotify.v1"
)

func main() {
	port := flag.Int("p", 8000, "port to listen on")
	dir := flag.String("d", ".", "directory to serve")
	flag.Parse()

	s := NewSrvServer(http.Dir(*dir))

	http.Handle("/", s)
	http.Handle("/_srv/api", websocket.Handler(s.wsHandler))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

type File struct {
	Path string `json:"path"`
}

type ByPath []File

func (b ByPath) Len() int           { return len(b) }
func (b ByPath) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByPath) Less(i, j int) bool { return b[i].Path < b[j].Path }

type SrvServer struct {
	dir        http.Dir
	fileServer http.Handler
	html       string
	js         string
}

func NewSrvServer(dir http.Dir) SrvServer {
	assets := rice.MustFindBox("assets")

	html, err := assets.String("index.html")
	if err != nil {
		panic(err)
	}

	js, err := assets.String("app.js")
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

func (s SrvServer) wsHandler(ws *websocket.Conn) {
	// Wait for the path
	var path string
	fmt.Fscan(ws, &s)

	// Send dir
	s.writeDirectory(ws, path)

	// Send dir whenever a file is modified
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	err = watcher.Add(path)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-watcher.Events:
			s.writeDirectory(ws, path)
		case err := <-watcher.Errors:
			panic(err)
		}
	}
}

func (s SrvServer) writeDirectory(w io.Writer, path string) {
	file, err := s.dir.Open(path)
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

	sort.Sort(ByPath(outputFiles))

	bytes, err := json.Marshal(outputFiles)
	if err != nil {
		panic(err)
	}

	_, err = io.WriteString(w, string(bytes))
	if err != nil {
		panic(err)
	}
}
