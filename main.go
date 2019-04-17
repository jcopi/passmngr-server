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

	fasthttp.ListenAndServe(":8080", NewPrimaryHandler(aliases, postHandlers, "./static/"))
}
