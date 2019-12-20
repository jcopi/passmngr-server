package main

import (
	"sync/atomic"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

const (
	period byte = byte('.')
)

// ApplyMiddleWare returns an http Handler that calls the middleware the the handler on the request
func ApplyMiddleWare(handler http.Handler, middle func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		middle(w, r)
		handler.ServeHTTP(w, r)
	})
}

// CountRequests increments the integer at the provided address, if there is no present DNT header
func CountRequests(count *uint64, w http.ResponseWriter, r *http.Request) {
	// If a request is made with a DNT header it should not be counted
	dntHeaders := r.Header[http.CanonicalHeaderKey("DNT")]
	for _, header := range dntHeaders {
		if strings.ContainsRune(header, '1') {
			return	
		}
	}
	
	atomic.AddUint64(count, uint64(1))
}

func NewCountRequests(count *uint64) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		CountRequests(count, w, r)	
	}
}

// CommonHeaders writes standard headers to the provided http.ResponseWriter
func CommonHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	w.Header().Set("Content-Security-Policy", "frame-ancestors 'none'; default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; object-src 'none'; font-src 'self'; form-action 'self'; connect-src wss://*.passmngr.com/socket wss://*.passmngr.io/socket wss://passmngr.com/socket wss://passmngr.io/socket")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-language", "en")
}

// Socket handles websocket connections.
// When this function returns, the websocket connection that it is handling is closed
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

// SocketUpgrader upgrades an http request to a websocket connection
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

func MatrixWellKnownServer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"m.server\": \"passmngr.modular.im:443\"}"))
}

func MatrixWellKnownClient(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow_Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"m.homeserver\": {\"base_url\": \"https://passmngr.modular.im\"},\"m.identity_server\": {\"base_url\": \"https://vector.im\"}}"))
}

// NotFound is a handler function for http not found errors (404)
func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Resource Not Found"))
}

// InternalError is a handler function for http internal errors (500)
func InternalError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal Error Occured"))
}

// InvalidMethod is a handler function for http invalid method errors (400)
func InvalidMethod(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Invalid HTTP Method"))
}

// SecurityError is a handler function for security errors.
// To avoid providing an unintentional oracle this method should be generic.
func SecurityError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Security Error Occured"))
}

// RedirectToHTTPS rediects http requests to the equivalent https resource
func RedirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	CommonHeaders(w, r)
	fmt.Println(r.Header)
	if r.URL.Host == "" {
		r.URL.Host = r.Host
	}
	r.URL.Scheme = "https"
	http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
}
