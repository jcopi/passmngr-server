package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

func main() {
	os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")

	cache := new(map[string][]Session)

	upgrader := websocket.Upgrader{
		EnableCompression: true,
	}

	mux := http.NewServeMux()

	mux.Handle("/socket", ApplyMiddleWare(http.HandlerFunc(NewSocketUpgrader(upgrader)), CommonHeaders))
	mux.Handle("/", ApplyMiddleWare(http.FileServer(http.Dir("./static/")), CommonHeaders))

	umux := http.NewServeMux()

	umux.Handle("/", ApplyMiddleWare(http.HandlerFunc(RedirectToHttps), CommonHeaders))

	allowableCipherSuites := []uint16{
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	}
	// PRODUCTION
	server := http.Server{
		Addr:    ":443",
		Handler: mux,
		TLSConfig: &tls.Config{
			MinVersion:   tls.VersionTLS12,
			CipherSuites: allowableCipherSuites,
		},
	}

	go http.ListenAndServe(":80", umux)

	// Production Certificates
	log.Fatal(server.ListenAndServeTLS("/etc/letsencrypt/live/www.passmngr.io/fullchain.pem", "/etc/letsencrypt/live/www.passmngr.io/privkey.pem"))

	// Debug Certificates
	//log.Fatal(server.ListenAndServeTLS("localhost.crt", "localhost.key"))
}
