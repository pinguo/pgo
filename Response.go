package pgo

import (
    "io"
    "net/http"
)

// Response http.ResponseWriter wrapper
type Response struct {
    http.ResponseWriter
    status int
    size   int
}

func (r *Response) reset(w http.ResponseWriter) {
    r.ResponseWriter = w
    r.status = http.StatusOK
    r.size = -1
}

func (r *Response) finish() {
    if r.size == -1 {
        r.size = 0
        r.ResponseWriter.WriteHeader(r.status)
    }
}

func (r *Response) WriteHeader(status int) {
    if status > 0 && r.status != status && r.size == -1 {
        if len(http.StatusText(status)) > 0 {
            r.status = status
        }
    }
}

func (r *Response) Write(data []byte) (n int, e error) {
    r.finish()
    n, e = r.ResponseWriter.Write(data)
    r.size += n
    return
}

func (r *Response) WriteString(s string) (n int, e error) {
    r.finish()
    n, e = io.WriteString(r.ResponseWriter, s)
    r.size += n
    return
}

func (r *Response) Status() int {
    return r.status
}

func (r *Response) Size() int {
    return r.size
}
