package Http

import (
    "net/http"
    "time"
)

// Option config option for http request
type Option struct {
    Header  http.Header
    Cookies []*http.Cookie
    Timeout time.Duration
}

// SetHeader set request header for the current request
func (o *Option) SetHeader(name, value string) *Option {
    if o.Header == nil {
        o.Header = make(http.Header)
    }

    o.Header.Set(name, value)
    return o
}

// SetCookie set request cookie for the current request
func (o *Option) SetCookie(name, value string) *Option {
    o.Cookies = append(o.Cookies, &http.Cookie{Name: name, Value: value})
    return o
}

// SetTimeout set request timeout for the current request
func (o *Option) SetTimeout(timeout time.Duration) *Option {
    o.Timeout = timeout
    return o
}
