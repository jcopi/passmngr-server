package main

import (
	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

func main() {
	aliases := []AliasDefinition{
		{HashableByteSlice("/index"), HashableByteSlice("/index.html")},
		{HashableByteSlice("index"), HashableByteSlice("index.html")},
	}
	middlewares := []MiddlewareDefinition{}

	upgrader := websocket.FastHTTPUpgrader{
		EnableCompression: true,
	}

	posts := []RouteDefinition{
		{HashableByteSlice("/helloword"), HelloWorld},
		{HashableByteSlice("/socket"), NewSocketUpgrader(upgrader)},
	}
	gets := []RouteDefinition{
		{HashableByteSlice("/socket"), NewSocketUpgrader(upgrader)},
	}

	go fasthttp.ListenAndServe(":80", RedirectToHttps)
	fasthttp.ListenAndServeTLS(":443", "/etc/letsencrypt/live/www.passmngr.io/fullchain.pem", "/etc/letsencrypt/live/www.passmngr.io/privkey.pem", NewPrimaryHandler(aliases, middlewares, posts, gets, "./static/"))

	// DEBUG
	// fasthttp.ListenAndServe(":80", NewPrimaryHandler(aliases, middlewares, posts, gets, "./static/"))
}
