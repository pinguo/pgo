package pgo

import (
    "net/http"
    "os"
    "path/filepath"

    "github.com/pinguo/pgo/Util"
)

// File static file plugin
type File struct {
    excludeExtension []string
}

func (f *File) SetExcludeExtension(v []interface{}) {
    for _, vv := range v {
        f.excludeExtension = append(f.excludeExtension, vv.(string))
    }
}

func (f *File) HandleRequest(ctx *Context) {
    // if extension is empty or excluded, pass
    if ext := filepath.Ext(ctx.GetPath()); ext == "" {
        return
    } else if len(f.excludeExtension) != 0 {
        if Util.SliceSearchString(f.excludeExtension, ext) != -1 {
            return
        }
    }

    // abort request
    defer ctx.Abort()

    // GET or HEAD method is required
    method := ctx.GetMethod()
    if method != http.MethodGet && method != http.MethodHead {
        http.Error(ctx.GetOutput(), "", http.StatusMethodNotAllowed)
        return
    }

    // file in @public directory is required
    path := filepath.Join(App.GetPublicPath(), Util.CleanPath(ctx.GetPath()))
    h, e := os.Open(path)
    if e != nil {
        http.Error(ctx.GetOutput(), "", http.StatusNotFound)
        return
    }

    defer h.Close()
    info, _ := h.Stat()
    http.ServeContent(ctx.GetOutput(), ctx.GetInput(), path, info.ModTime(), h)
}
