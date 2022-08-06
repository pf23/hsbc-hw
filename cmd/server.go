package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"

	"hsbc-hw/serving"
)

var (
	port = flag.Int("port", 8080, "The port to listen on")
)

func main() {
	flag.Parse()
	if *port < 0 || *port > 65535 {
		log.Fatalf("authenticate_server: invalid --port, must be in [0, 65535], found %d", *port)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("Hello"))
	})
	go func() {
		log.Printf("authenticate_server: start listen on :%d", *port)
		http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Printf("authenticate_server: gracefully shutdown")
	serving.Cleanup()
}
