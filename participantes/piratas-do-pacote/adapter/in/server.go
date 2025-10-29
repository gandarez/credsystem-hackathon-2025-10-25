package in

import (
	"bytes"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/piratas-do-pacote/global"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/prefork"
)

var (
	pathFindService = []byte("/api/find-service")
	pathHealth      = []byte("/api/healthz")
)

func (h *HttpHandler) staticHandler(ctx *fasthttp.RequestCtx) {
	path := ctx.Path()

	if ctx.IsGet() {
		switch {

		case bytes.Equal(path, pathHealth):
			h.handleHeathZ(ctx)
			return
		}
	}

	if ctx.IsPost() {
		switch {
		case bytes.Equal(path, pathFindService):
			h.handleFindService(ctx)
			return
		}
	}

	ctx.SetStatusCode(fasthttp.StatusNotFound)
}

func (h *HttpHandler) ListenAndServe() {

	s := &fasthttp.Server{
		Handler:                       h.staticHandler,
		Name:                          defaultServerName,
		NoDefaultServerHeader:         true,
		NoDefaultContentType:          true,
		NoDefaultDate:                 true,
		DisableHeaderNamesNormalizing: true,
		ReduceMemoryUsage:             false,
		ReadTimeout:                   defaultReadTimeout,
		WriteTimeout:                  defaultWriteTimeout,
		ReadBufferSize:                defaultReadBufferSize,
		WriteBufferSize:               defaultWriteBufferSize,
		MaxRequestBodySize:            defaultMaxRequestBodySize,
		LogAllErrors:                  false,
	}

	usePrefork := os.Getenv("PREFORK") == "1"

	addr := ":" + global.GetEnvDefault("PORT", "8080")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		var err error
		if usePrefork {
			p := prefork.New(s)
			log.Printf("[startup] prefork ON listening at %s", addr)
			err = p.ListenAndServe(addr)
		} else {
			log.Printf("[startup] listening at %s", addr)
			err = s.ListenAndServe(addr)
		}
		if err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-stop
	log.Printf("[shutdown] starting graceful shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.ShutdownWithContext(ctx); err != nil {
		log.Printf("[shutdown] forced: %v", err)
	}
	log.Printf("[shutdown] done")
	return
}
