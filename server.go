package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"golang.org/x/net/websocket"

	"github.com/dustin/go-humanize"
	"gopkg.in/fsnotify.v1"
)

type Server struct {
	dir        http.Dir
	dirString  string
	fileServer http.Handler
	html       string
	js         string
}

func NewServer(dir string) Server {
	html, err := Asset("build/index.html")
	if err != nil {
		panic(err)
	}

	js, err := Asset("build/app.js")
	if err != nil {
		panic(err)
	}

	httpDir := http.Dir(dir)

	s := Server{
		dir:        httpDir,
		dirString:  dir,
		fileServer: http.FileServer(httpDir),
		html:       string(html),
		js:         string(js),
	}

	return s
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		// Redirect /dir to /dir/
		if !strings.HasSuffix(r.URL.Path, "/") {
			w.Header()["Location"] = []string{
				r.URL.Path + "/",
			}
			w.WriteHeader(http.StatusFound)
			return
		}

		w.Header()["Content-Type"] = []string{
			"text/html; charset=utf-8",
		}
		io.WriteString(w, s.html)
		return
	} else {
		s.fileServer.ServeHTTP(w, r)
		return
	}
}

func (s Server) wsHandler(ws *websocket.Conn) {
	// Wait for the path
	var _path string
	fmt.Fscanln(ws, &_path)

	path, err := url.QueryUnescape(_path)
	if err != nil {
		panic(err)
	}

	// Send dir
	err = s.writeDirectory(ws, path)
	if err != nil {
		return
	}

	// Send dir whenever a file is modified
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	err = watcher.Add(s.dirString + path)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-watcher.Events:
			err := s.writeDirectory(ws, path)
			if err != nil {
				return
			}
		case err := <-watcher.Errors:
			panic(err)
		}
	}
}

func (s Server) writeDirectory(w io.Writer, path string) error {
	file, err := s.dir.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("oh no")
	}

	files, err := readAllFiles(file)
	if err != nil {
		return err
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

	err = json.NewEncoder(w).Encode(outputFiles)
	if err != nil {
		return err
	}

	return nil
}

func readAllFiles(file http.File) ([]os.FileInfo, error) {
	files := []os.FileInfo{}
	for {
		more, err := file.Readdir(100)
		if err == io.EOF {
			return files, nil
		} else if err != nil {
			return nil, err
		}
		files = append(files, more...)
	}
}
