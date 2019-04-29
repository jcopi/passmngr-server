package main

import (
	"fmt"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

const (
	period byte = byte('.')
)

type Route func(*fasthttp.RequestCtx)
type Middleware func(*fasthttp.RequestCtx)
type Alias struct {
	path     HashableByteSlice
	pathHash uint32
}

// PrimaryHandler is the function that will handle every http request
func PrimaryHandler(aliases Hashmap, middlewares Hashmap, postRoutes Hashmap, getRoutes Hashmap, fsHandler func(*fasthttp.RequestCtx), ctx *fasthttp.RequestCtx) {
	path := HashableByteSlice(ctx.Path())
	pathHash := path.Hash()

	CommonHeaders(ctx)

	if alias := aliases.GetPreHash(path, pathHash); alias != nil {
		path = alias.(Alias).path
		pathHash = alias.(Alias).pathHash
		ctx.URI().SetPathBytes(path)
	}

	if middleware := middlewares.GetPreHash(path, pathHash); middleware != nil {
		middleware.(Middleware)(ctx)
	}

	if ctx.IsPost() {
		if postRoute := postRoutes.GetPreHash(path, pathHash); postRoute != nil {
			postRoute.(Route)(ctx)
		} else {
			NotFound(ctx)
		}
	} else if ctx.IsGet() {
		if getRoute := getRoutes.GetPreHash(path, pathHash); getRoute != nil {
			getRoute.(Route)(ctx)
		} else {
			fsHandler(ctx)
		}
	} else {
		InvalidMethod(ctx)
	}
}

type RouteDefinition struct {
	path  HashableByteSlice
	route Route
}
type MiddlewareDefinition struct {
	path       HashableByteSlice
	middleware Middleware
}
type AliasDefinition struct {
	path      HashableByteSlice
	aliasPath HashableByteSlice
}

// NewPrimaryHandler returns a new request handler with the appropriate function signature
func NewPrimaryHandler(aliases []AliasDefinition, middlewares []MiddlewareDefinition, posts []RouteDefinition, gets []RouteDefinition, root string) func(*fasthttp.RequestCtx) {
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

	aliasMap := NewHashmap()
	middlewareMap := NewHashmap()
	postMap := NewHashmap()
	getMap := NewHashmap()

	for _, a := range aliases {
		aliasMap.Set(a.path, Alias{a.aliasPath, a.aliasPath.Hash()})
	}

	for _, m := range middlewares {
		middlewareMap.Set(m.path, m.middleware)
	}

	for _, p := range posts {
		postMap.Set(p.path, p.route)
	}

	for _, g := range gets {
		getMap.Set(g.path, g.route)
	}

	return func(ctx *fasthttp.RequestCtx) {
		PrimaryHandler(aliasMap, middlewareMap, postMap, getMap, fsHandler, ctx)
	}
}

// PrimaryHandler is the function that will handle every http request

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

func SecureSocket(ws *websocket.Conn) {
	defer ws.Close()

	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			break
		}

		fmt.Printf("Received Message [%v] '%v'\n", mt, string(message))
	}
}

func SocketUpgrader(upgrader websocket.FastHTTPUpgrader, ctx *fasthttp.RequestCtx) {
	err := upgrader.Upgrade(ctx, SecureSocket)
	if err != nil {
		InternalError(ctx)
	}
}

func NewSocketUpgrader(upgrader websocket.FastHTTPUpgrader) func(*fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		SocketUpgrader(upgrader, ctx)
	}
}

// NotFound is the resource not found 404 request handler
func NotFound(ctx *fasthttp.RequestCtx) {
	ctx.Error("Resource Not Found", fasthttp.StatusNotFound)
}

func InternalError(ctx *fasthttp.RequestCtx) {
	ctx.Error("Internal Error Occured", fasthttp.StatusInternalServerError)
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
