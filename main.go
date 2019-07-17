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

	mux.Handle("/socket", ApplyMiddleWare(http.HandlerFunc(NewSocketUpgrader(upgrader)), CommonHeaders))
	mux.Handle("/", ApplyMiddleWare(http.FileServer(http.Dir("./static/")), CommonHeaders))

	umux := http.NewServeMux()

	umux.Handle("/", ApplyMiddleWare(http.HandlerFunc(RedirectToHttps), CommonHeaders))

	// DEBUG
	//go http.ListenAndServe(":80", umux)
	//log.Fatal(http.ListenAndServeTLS(":443", "./localhost.crt", "./localhost.key", mux))

	// PRODUCTION
	go http.ListenAndServe(":80", umux)
	log.Fatal(http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/www.passmngr.io/fullchain.pem", "/etc/letsencrypt/live/www.passmngr.io/privkey.pem", mux))
}
