package main

//go:generate mkdir -p build
//go:generate sh -c "cat assets/angular.min.js assets/angular-route.min.js assets/app.js > build/app.js"
//go:generate cp assets/index.html build/index.html
//go:generate go-bindata build/

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"code.google.com/p/go.net/websocket"
)

func main() {
	port := flag.Int("p", 8000, "port to listen on")
	dir := flag.String("d", ".", "directory to serve")
	flag.Parse()

	s := NewServer(http.Dir(*dir))

	http.Handle("/", s)
	http.Handle("/_srv/api", websocket.Handler(s.wsHandler))

	fmt.Fprintf(os.Stderr, "Listening on 0.0.0.0:%d\n", *port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
