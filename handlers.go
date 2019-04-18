package main

import (
	"bytes"

	"bitbucket.org/creachadair/cityhash"
	"github.com/valyala/fasthttp"
)

const (
	period byte = byte('.')
)

// Route is the structure holding information and handlers for post routes
type Route struct {
	path     []byte
	pathHash uint32
	fn       func(*fasthttp.RequestCtx)
}

// NewRoute returns a new route structure
func NewRoute(path []byte, fn func(*fasthttp.RequestCtx)) Route {
	return Route{path, cityhash.Hash32(path), fn}
}

// AliasRoute is the structure holding information for a route alias
type AliasRoute struct {
	alias     []byte
	aliasHash uint32
	path      []byte
	pathHash  uint32
}

// NewAliasRoute creates a new alias route struct
func NewAliasRoute(alias []byte, aliased []byte) AliasRoute {
	return AliasRoute{alias, cityhash.Hash32(alias), aliased, cityhash.Hash32(aliased)}
}

// NewPrimaryHandler returns a new request handler with the appropriate function signature
func NewPrimaryHandler(aliases []AliasRoute, posts []Route, root string) func(*fasthttp.RequestCtx) {
	fs := &fasthttp.FS{
		Root:                 root,
		IndexNames:           []string{"index.html"},
		GenerateIndexPages:   false,
		Compress:             true,
		AcceptByteRange:      true,
		CompressedFileSuffix: string(".compress"),
		PathNotFound:         NotFound,
	}

	fsHandler := fs.NewRequestHandler()

	return func(ctx *fasthttp.RequestCtx) {
		PrimaryHandler(aliases, posts, fsHandler, ctx)
	}
}

// PrimaryHandler is the function that will handle every http request
func PrimaryHandler(aliases []AliasRoute, postRoutes []Route, fsHandler func(*fasthttp.RequestCtx), ctx *fasthttp.RequestCtx) {
	CommonHeaders(ctx)

	path := ctx.Path()
	pathHash := cityhash.Hash32(path)

	for _, r := range aliases {
		if r.aliasHash == pathHash && bytes.Equal(r.alias, path) {
			path = r.path
			pathHash = r.pathHash
			ctx.URI().SetPathBytes(path)
			break
		}
	}

	if ctx.IsPost() {
		for _, r := range postRoutes {
			if r.pathHash == pathHash && bytes.Equal(r.path, path) {
				r.fn(ctx)
				break
			}
		}
	} else if ctx.IsGet() {
		fsHandler(ctx)
	} else {
		InvalidMethod(ctx)
	}
}

func CommonHeaders(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	ctx.Response.Header.Set("Content-language", "en")
}

// HelloWorld is a hello world request handler
func HelloWorld(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/plain")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody([]byte("hello world."))
}

// NotFound is the resource not found 404 request handler
func NotFound(ctx *fasthttp.RequestCtx) {
	ctx.Error("Resource Not Found", fasthttp.StatusNotFound)
}

func InvalidMethod(ctx *fasthttp.RequestCtx) {
	ctx.Error("Invalid HTTP Method", fasthttp.StatusBadRequest)
}

func SecurityError(ctx *fasthttp.RequestCtx) {
	ctx.Error("Security Error Occured", fasthttp.StatusInternalServerError)
}

func RedirectToHttps(ctx *fasthttp.RequestCtx) {
	CommonHeaders(ctx)
	uri := ctx.URI()
	uri.SetScheme("https")
	ctx.Redirect(string(uri.FullURI()), fasthttp.StatusMovedPermanently)
}
