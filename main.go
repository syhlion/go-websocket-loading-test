package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8001", "we service address")

func main() {
	flag.Parse()
	go h.run()
	http.HandleFunc("/", serverHandler)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}
