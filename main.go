package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"code.google.com/p/go.net/websocket"
)

func main() {
	port := flag.Int("p", 8000, "port to listen on")
	dir := flag.String("d", ".", "directory to serve")
	flag.Parse()

	s := NewServer(http.Dir(*dir))

	http.Handle("/", s)
	http.Handle("/_srv/api", websocket.Handler(s.wsHandler))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
