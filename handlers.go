package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	period byte = byte('.')
)

func CommonHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	w.Header().Set("Content-language", "en")
}

func Socket(ws *websocket.Conn) {
	defer ws.Close()

	// Websocket echo server
	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			break
		}
		err = ws.WriteMessage(mt, message)
		if err != nil {
			break
		}

		// fmt.Printf("Received Message [%v] '%v'\n", mt, string(message))
	}
}

func SocketUpgrader(upgrader websocket.Upgrader, w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		InternalError(w, r)
	}
	Socket(ws)
}

func NewSocketUpgrader(upgrader websocket.Upgrader) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		SocketUpgrader(upgrader, w, r)
	}
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Resource Not Found"))
}

func InternalError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal Error Occured"))
}

func InvalidMethod(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Invalid HTTP Method"))
}

func SecurityError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Security Error Occured"))
}

func RedirectToHttps(w http.ResponseWriter, r *http.Request) {
	CommonHeaders(w, r)
	fmt.Println(r.Header)
	if r.URL.Host == "" {
		r.URL.Host = r.Host
	}
	r.URL.Scheme = "https"
	http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
}
