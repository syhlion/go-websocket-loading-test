package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

var addr = flag.String("addr", ":8001", "ws address")
var logDir = flag.String("logDir", "", "log dir")

func main() {
	flag.Parse()

	f, err := os.OpenFile(*logDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0775)
	if err == nil {
		log.SetOutput(f)
	}

	go h.run()
	http.HandleFunc("/", serverHandler)
	log.Fatal(http.ListenAndServe(*addr, nil))

}
