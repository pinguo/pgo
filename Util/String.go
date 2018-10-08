package Util

import (
    "bytes"
    "crypto/md5"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io"
    "unicode"
)

func IsAllDigit(s string) bool {
    for _, v := range s {
        if !unicode.IsDigit(v) {
            return false
        }
    }

    if len(s) == 0 {
        return false
    }

    return true
}

func IsAllLetter(s string) bool {
    for _, v := range s {
        if !unicode.IsLetter(v) {
            return false
        }
    }

    if len(s) == 0 {
        return false
    }

    return true
}

func IsAllLower(s string) bool {
    for _, v := range s {
        if !unicode.IsLower(v) {
            return false
        }
    }

    if len(s) == 0 {
        return false
    }

    return true
}

func IsAllUpper(s string) bool {
    for _, v := range s {
        if !unicode.IsUpper(v) {
            return false
        }
    }

    if len(s) == 0 {
        return false
    }

    return true
}

func Md5Bytes(v interface{}) []byte {
    ctx := md5.New()
    switch vv := v.(type) {
    case string:
        io.WriteString(ctx, vv)
    case []byte:
        ctx.Write(vv)
    default:
        if j, e := json.Marshal(v); e == nil {
            ctx.Write(j)
        } else {
            var buf bytes.Buffer
            fmt.Fprint(&buf, v)
            ctx.Write(buf.Bytes())
        }
    }

    return ctx.Sum(nil)
}

func Md5String(v interface{}) string {
    return hex.EncodeToString(Md5Bytes(v))
}
