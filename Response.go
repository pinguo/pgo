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

// WriteHeader cache status code until first write operation.
func (r *Response) WriteHeader(status int) {
    if status > 0 && r.status != status && r.size == -1 {
        if len(http.StatusText(status)) > 0 {
            r.status = status
        }
    }
}

// Write write data to underlying http.ResponseWriter
// and record num bytes that has written.
func (r *Response) Write(data []byte) (n int, e error) {
    r.finish()
    n, e = r.ResponseWriter.Write(data)
    r.size += n
    return
}

// WriteString write string data to underlying http.ResponseWriter
// and record num bytes that has written.
func (r *Response) WriteString(s string) (n int, e error) {
    r.finish()
    n, e = io.WriteString(r.ResponseWriter, s)
    r.size += n
    return
}

// ReadFrom is here to optimize copying from a regular file
// to a *net.TCPConn with sendfile.
func (r *Response) ReadFrom(src io.Reader) (n int64, e error) {
    r.finish()
    if rf, ok := r.ResponseWriter.(io.ReaderFrom); ok {
        n, e = rf.ReadFrom(src)
    } else {
        n, e = io.Copy(r.ResponseWriter, src)
    }
    r.size += int(n)
    return
}

// Status get response status.
func (r *Response) Status() int {
    return r.status
}

// Size get num bytes that has written.
func (r *Response) Size() int {
    return r.size
}
