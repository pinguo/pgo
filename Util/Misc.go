package Util

import (
    "bytes"
    "fmt"
    "hash/crc32"
    "math/rand"
    "net"
    "os"
    "regexp"
    "runtime"
    "strconv"
    "strings"
    "sync/atomic"
    "time"
)

var (
    seqId  uint32
    ipAddr []byte

    // env expand regexp: ${env}, ${env||default}
    envRe = regexp.MustCompile(`\$\{[^\}\|]+(\|\|[^\$\{\}]+?)?\}`)

    // lang format regexp: zh-cn, zh, en_US
    langRe = regexp.MustCompile(`(?i)([a-z]+)(?:[_-]([a-z]+))?`)

    // stack trace regexp: <table>/path/to/src/file.go:line<space>
    traceRe = regexp.MustCompile(`^\t(.*)/src/(.*:\d+)\s`)

    // version format regexp: v10.1.0
    verFmtRe = regexp.MustCompile(`(?i)^v?(\d+\.*)+`)

    // version element regexp: [0-9]+ or [a-z]+
    verEleRe = regexp.MustCompile(`(?i)\d+|[a-z]+`)
)

func init() {
    // generate a random sequence id
    random := rand.New(rand.NewSource(time.Now().UnixNano()))
    seqId = random.Uint32()

    // get ipv4 address
    if addrs, e := net.InterfaceAddrs(); e == nil {
        for _, v := range addrs {
            if ip, ok := v.(*net.IPNet); ok {
                ipv4 := ip.IP.To4()
                if !ip.IP.IsLoopback() && ipv4 != nil {
                    ipAddr = []byte(ipv4.String())
                }
            }
        }
    }
}

// GenUniqueId generate a 24 bytes unique id
func GenUniqueId() string {
    now := time.Now()
    seq := atomic.AddUint32(&seqId, 1)

    return fmt.Sprintf("%08x%06x%04x%06x",
        now.Unix()&0xFFFFFFFF,
        crc32.ChecksumIEEE(ipAddr)&0xFFFFFF,
        os.Getpid()&0xFFFF,
        seq&0xFFFFFF,
    )
}

// ExpandEnv expand env variables, format: ${env}, ${env||default}
func ExpandEnv(data []byte) []byte {
    rf := func(s []byte) []byte {
        tmp := bytes.Split(s[2:len(s)-1], []byte{'|', '|'})
        env := bytes.TrimSpace(tmp[0])

        if val, ok := os.LookupEnv(string(env)); ok {
            // return env value
            return []byte(val)
        } else if len(tmp) > 1 {
            // return default value
            return bytes.TrimSpace(tmp[1])
        }

        // return original
        return s
    }

    return envRe.ReplaceAllFunc(data, rf)
}

// FormatLanguage format lang to ll-CC format
func FormatLanguage(lang string) string {
    matches := langRe.FindStringSubmatch(lang)
    if len(matches) != 3 {
        return ""
    }

    matches[1] = strings.ToLower(matches[1])
    matches[2] = strings.ToUpper(matches[2])

    switch matches[2] {
    case "CHS", "HANS":
        matches[2] = "CN"
    case "CHT", "HANT":
        matches[2] = "TW"
    }

    if len(matches[2]) == 0 {
        return matches[1]
    }

    return matches[1] + "-" + matches[2]
}

// PanicTrace get panic trace
func PanicTrace(maxDepth int, multiLine bool) string {
    buf := make([]byte, 1024)
    if n := runtime.Stack(buf, false); n < len(buf) {
        buf = buf[:n]
    }

    stack := bytes.NewBuffer(buf)
    sources := make([]string, 0, maxDepth)
    meetPanic := false

    for {
        line, err := stack.ReadString('\n')
        if err != nil || len(sources) >= maxDepth {
            break
        }

        mat := traceRe.FindStringSubmatch(line)
        if mat == nil {
            continue
        }

        // skip until first panic
        if strings.HasPrefix(mat[2], "runtime/panic.go") {
            meetPanic = true
            continue
        }

        // skip system file
        if strings.HasPrefix(mat[1], runtime.GOROOT()) {
            continue
        }

        if meetPanic {
            sources = append(sources, mat[2])
        }
    }

    if multiLine {
        return strings.Join(sources, "\n")
    }

    return strings.Join(sources, ",")
}

// FormatVersion format version to have minimum depth,
// eg. FormatVersion("v10...2....2.1-alpha", 5) == "v10.2.2.1.0-alpha"
func FormatVersion(ver string, minDepth int) string {
    replaceFunc := func(s string) string {
        p, n := strings.Split(s, "."), 0
        for i, j := 0, len(p); i < j; i++ {
            if len(p[i]) > 0 {
                if n != i {
                    p[n] = p[i]
                }
                n++
            }
        }

        for p = p[:n]; n < minDepth; n++ {
            p = append(p, "0")
        }

        return strings.Join(p, ".")
    }

    return verFmtRe.ReplaceAllStringFunc(ver, replaceFunc)
}

// VersionCompare compare versions like version_compare of php,
// special version strings these are handled in the following order,
// (any string not found) < dev < alpha = a < beta = b < rc < #(empty) < ##(digit) < pl = p,
// result: -1(ver1 < ver2), 0(ver1 == ver2), 1(ver1 > ver2)
func VersionCompare(ver1, ver2 string) int {
    // trim leading v character
    ver1 = strings.TrimLeft(ver1, "vV")
    ver2 = strings.TrimLeft(ver2, "vV")

    v1 := verEleRe.FindAllStringSubmatch(ver1, -1)
    v2 := verEleRe.FindAllStringSubmatch(ver2, -1)

    vm := map[string]int{"dev": 1, "alpha": 2, "a": 2, "beta": 3, "b": 3, "rc": 4, "#": 5, "##": 6, "pl": 7, "p": 7}

    isDigit := func(b byte) bool {
        return '0' <= b && b <= '9'
    }

    compare := func(p1, p2 string) int {
        l1, l2 := 0, 0
        if isDigit(p1[0]) && isDigit(p2[0]) {
            l1, _ = strconv.Atoi(p1)
            l2, _ = strconv.Atoi(p2)
        } else if !isDigit(p1[0]) && !isDigit(p2[0]) {
            l1 = vm[strings.ToLower(p1)]
            l2 = vm[strings.ToLower(p2)]
        } else if isDigit(p1[0]) {
            l1 = vm["##"]
            l2 = vm[strings.ToLower(p2)]
        } else {
            l1 = vm[strings.ToLower(p1)]
            l2 = vm["##"]
        }

        if l1 > l2 {
            return 1
        } else if l1 < l2 {
            return -1
        } else {
            return 0
        }
    }

    c, n, l1, l2 := 0, 0, len(v1), len(v2)
    for ; n < l1 && n < l2 && c == 0; n++ {
        c = compare(v1[n][0], v2[n][0])
    }

    if c == 0 {
        if n != l1 {
            c = compare(v1[n][0], "#")
        } else if n != l2 {
            c = compare("#", v2[n][0])
        }
    }

    return c
}
