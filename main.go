package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func main() {
	upgrader := websocket.Upgrader{
		EnableCompression: true,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/socket", NewSocketUpgrader(upgrader))
	mux.Handle("/", http.FileServer(http.Dir("./static/")))

	umux := http.NewServeMux()

	umux.HandleFunc("/", RedirectToHttps)

	// DEBUG
	// go http.ListenAndServe(":80", umux)
	// log.Fatal(http.ListenAndServeTLS(":443", "./localhost.crt", "./localhost.key", mux))

	// PRODUCTION
	go http.ListenAndServe(":80", umux)
	log.Fatal(http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/www.passmngr.io/fullchain.pem", "/etc/letsencrypt/live/www.passmngr.io/privkey.pem", mux))
}
