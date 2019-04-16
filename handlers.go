package main

import (
	"bytes"

	"bitbucket.org/creachadair/cityhash"
	"github.com/valyala/fasthttp"
)

type Route struct {
	path     []byte
	pathHash uint32
	fn       func(*fasthttp.RequestCtx)
}

func NewRoute(path []byte, fn func(*fasthttp.RequestCtx)) Route {
	return Route{path, cityhash.Hash32(path), fn}
}

type AliasRoute struct {
	alias     []byte
	aliasHash uint32
	path      []byte
	pathHash  uint32
}

func NewAliasRoute(alias []byte, aliased []byte) AliasRoute {
	return AliasRoute{alias, cityhash.Hash32(alias), aliased, cityhash.Hash32(aliased)}
}

// RouteClosure creates a closure to "bind" one argument to a function
func RouteClosure(f func([]Route, *fasthttp.RequestCtx), r []Route) func(*fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		f(r, ctx)
	}
}

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
func PrimaryHandler(aliases []AliasRoute, postRoutes []Route, fshandler func(*fasthttp.RequestCtx), ctx *fasthttp.RequestCtx) {
	path := ctx.Path()
	pathHash := cityhash.Hash32(path)

	for _, r := range aliases {
		if r.aliasHash == pathHash && bytes.Equal(r.alias, path) {
			path = r.path
			pathHash = r.pathHash
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
		NotFound(ctx)
	}

	/*if ctx.IsPost() {
		path := ctx.Path()
		pathHash := cityhash.Hash32(path)

		routeID := -1

		for i := 0; i <= len(postRoutes); i++ {
			if pathHash == postRoutes[i].pathHash && bytes.Equal(path, postRoutes[i].path) {
				routeID = i
				break
			}
		}

		if routeID >= 0 {
			postRoutes[routeID].fn(ctx)
			return
		}

		NotFound(ctx)
	} else if ctx.IsGet() {
		NotFound(ctx)
	}*/
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
