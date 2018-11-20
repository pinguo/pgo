package pgo

import (
    "compress/gzip"
    "io"
    "io/ioutil"
    "net/http"
    "path/filepath"
    "strings"
    "sync"
)

// Gzip gzip compression plugin
type Gzip struct {
    pool sync.Pool
}

func (g *Gzip) Construct() {
    g.pool.New = func() interface{} {
        return &gzipWriter{writer: gzip.NewWriter(ioutil.Discard)}
    }
}

func (g *Gzip) HandleRequest(ctx *Context) {
    ae := ctx.GetHeader("Accept-Encoding", "")
    if !strings.Contains(ae, "gzip") {
        return
    }

    ext := filepath.Ext(ctx.GetPath())
    switch strings.ToLower(ext) {
    case ".png", ".gif", ".jpeg", ".jpg", ".ico":
        return
    }

    gw := g.pool.Get().(*gzipWriter)
    gw.reset(ctx)

    defer func() {
        gw.finish()
        g.pool.Put(gw)
    }()

    ctx.Next()
}

type gzipWriter struct {
    http.ResponseWriter
    writer *gzip.Writer
    ctx    *Context
    size   int
}

func (g *gzipWriter) reset(ctx *Context) {
    g.ResponseWriter = ctx.GetOutput()
    g.ctx = ctx
    g.size = -1
    ctx.SetOutput(g)
}

func (g *gzipWriter) finish() {
    if g.size > 0 {
        g.writer.Close()
    }
}

func (g *gzipWriter) start() {
    if g.size == -1 {
        g.size = 0
        g.writer.Reset(g.ResponseWriter)
        g.ctx.SetHeader("Content-Encoding", "gzip")
    }
}

func (g *gzipWriter) Flush() {
    if g.size > 0 {
        g.writer.Flush()
    }

    if flusher, ok := g.ResponseWriter.(http.Flusher); ok {
        flusher.Flush()
    }
}

func (g *gzipWriter) Write(data []byte) (n int, e error) {
    if len(data) == 0 {
        return 0, nil
    }

    g.start()

    n, e = g.writer.Write(data)
    g.size += n
    return
}

func (g *gzipWriter) WriteString(data string) (n int, e error) {
    if len(data) == 0 {
        return 0, nil
    }

    g.start()

    n, e = io.WriteString(g.writer, data)
    g.size += n
    return
}
