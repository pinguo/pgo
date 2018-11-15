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
        return &gzipWriter{nil, gzip.NewWriter(ioutil.Discard)}
    }
}

func (g *Gzip) HandleRequest(ctx *Context) {
    ae := ctx.GetHeader("Accept-Encoding", "")
    if !strings.Contains(ae, "gzip") {
        return
    }

    ext := filepath.Ext(ctx.GetPath())
    switch strings.ToLower(ext) {
    case ".png", ".gif", ".jpeg", ".jpg":
        return
    }

    gw := g.pool.Get().(*gzipWriter)
    defer g.pool.Put(gw)

    gw.ResponseWriter = ctx.GetOutput()
    gw.writer.Reset(ctx.GetOutput())
    defer gw.writer.Close()

    ctx.SetOutput(gw)
    ctx.SetHeader("Content-Encoding", "gzip")
    ctx.Next()
}

type gzipWriter struct {
    http.ResponseWriter
    writer *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (n int, e error) {
    return g.writer.Write(data)
}

func (g *gzipWriter) WriteString(data string) (n int, e error) {
    return io.WriteString(g.writer, data)
}
