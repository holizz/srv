package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"

	"code.google.com/p/go.net/websocket"

	"github.com/GeertJohan/go.rice"
	"github.com/dustin/go-humanize"
	"gopkg.in/fsnotify.v1"
)

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
	var _path string
	fmt.Fscanln(ws, &_path)

	path, err := url.QueryUnescape(_path)
	if err != nil {
		panic(err)
	}

	// Send dir
	s.writeDirectory(ws, path)

	// Send dir whenever a file is modified
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	err = watcher.Add("." + path)
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
			Size: humanize.IBytes(uint64(f.Size())),
			Dir:  f.IsDir(),
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
