package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
)

var addr = flag.String("addr", ":8001", "ws address")
var logDir = flag.String("logDir", "", "log dir")

func main() {
	flag.Parse()

	f, err := os.OpenFile(*logDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0775)
	if err == nil {
		log.SetOutput(f)
	}
	log.SetPrefix(strconv.FormatInt(makeTimestamp(), 10) + ":")

	go h.run()
	http.HandleFunc("/", serverHandler)
	log.Println("Server Start", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))

}
