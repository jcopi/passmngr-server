package main

import (
	"bitbucket.org/creachadair/cityhash"
	"github.com/valyala/fasthttp"
)

func main() {
	hm := NewHashmap()
	k := HashableByteSlice{0, 1, 2, 3, 4}
	v := 72
	hm.Set(k, v)

	postHandlers := []Route{
		{[]byte("helloworld"), cityhash.Hash32([]byte("helloworld")), HelloWorld},
	}

	aliases := []AliasRoute{
		NewAliasRoute([]byte("/index"), []byte("/index.html")),
	}

	go fasthttp.ListenAndServe(":80", RedirectToHttps)
	fasthttp.ListenAndServeTLS(":443", "/etc/letsencrypt/live/www.passmngr.io/fullchain.pem", "/etc/letsencrypt/live/www.passmngr.io/privkey.pem", NewPrimaryHandler(aliases, postHandlers, "./static/"))
}
