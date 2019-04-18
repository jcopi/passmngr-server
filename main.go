package main

import (
	"bitbucket.org/creachadair/cityhash"
	"github.com/valyala/fasthttp"
)

func main() {
	postHandlers := []Route{
		{[]byte("helloworld"), cityhash.Hash32([]byte("helloworld")), HelloWorld},
	}

	aliases := []AliasRoute{
		NewAliasRoute([]byte("/index"), []byte("/index.html")),
	}

	go fasthttp.ListenAndServe(":80", RedirectToHttps)
	fasthttp.ListenAndServeTLS(":443", "/etc/letsencrypt/live/www.passmngr.io/fullchain.pem", "/etc/letsencrypt/live/www.passmngr.io/privkey.pem", NewPrimaryHandler(aliases, postHandlers, "./static/"))
}
