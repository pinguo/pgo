package pgo

import (
    "net/http"
    "os"
    "path/filepath"

    "github.com/pinguo/pgo/Util"
)

// File file plugin, this plugin only handle file in @public directory,
// request url with empty or excluded extension will not be handled.
type File struct {
    excludeExtensions []string
}

func (f *File) SetExcludeExtensions(v []interface{}) {
    for _, vv := range v {
        f.excludeExtensions = append(f.excludeExtensions, vv.(string))
    }
}

func (f *File) HandleRequest(ctx *Context) {
    // if extension is empty or excluded, pass
    if ext := filepath.Ext(ctx.GetPath()); ext == "" {
        return
    } else if len(f.excludeExtensions) != 0 {
        if Util.SliceSearchString(f.excludeExtensions, ext) != -1 {
            return
        }
    }

    // skip other plugins
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
