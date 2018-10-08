package Redis

import (
    "bufio"
    "bytes"
    "fmt"
    "io"
    "net"
    "strconv"
    "time"

    "github.com/pinguo/pgo"
)

func newConn(addr string, nc net.Conn, pool *Pool) *Conn {
    r := bufio.NewReader(nc)
    w := bufio.NewWriter(nc)

    return &Conn{
        addr: addr,
        nc:   nc,
        rw:   bufio.NewReadWriter(r, w),
        pool: pool,
    }
}

type Conn struct {
    prev       *Conn
    next       *Conn
    lastActive time.Time

    addr string
    nc   net.Conn
    rw   *bufio.ReadWriter
    pool *Pool
    down bool
}

func (c *Conn) Close(force bool) {
    if force || c.down || !c.pool.putFreeConn(c) {
        c.nc.Close()
    } else {
        c.lastActive = time.Now()
    }
}

func (c *Conn) CheckActive() bool {
    if time.Since(c.lastActive) < c.pool.maxIdleTime {
        return true
    }

    c.ExtendDeadLine()
    payload, ok := c.Do("PING").([]byte)
    return ok && bytes.Equal(payload, replyPong)
}

func (c *Conn) ExtendDeadLine(deadLine ...time.Duration) bool {
    deadLine = append(deadLine, c.pool.netTimeout)
    return c.nc.SetDeadline(time.Now().Add(deadLine[0])) == nil
}

func (c *Conn) Do(cmd string, args ...interface{}) interface{} {
    c.WriteCmd(cmd, args...)
    return c.ReadReply()
}

func (c *Conn) WriteCmd(cmd string, args ...interface{}) {
    fmt.Fprintf(c.rw, "*%d\r\n$%d\r\n%s\r\n", len(args)+1, len(cmd), cmd)
    for _, arg := range args {
        argBytes := pgo.Encode(arg)
        fmt.Fprintf(c.rw, "$%d\r\n", len(argBytes))
        c.rw.Write(argBytes)
        c.rw.Write(lineEnding)
    }

    if e := c.rw.Flush(); e != nil {
        c.parseError(errSendFailed+e.Error(), true)
    }
}

// read reply from server,
// return []byte, int, nil or slice of these types
func (c *Conn) ReadReply() interface{} {
    line, e := c.rw.ReadSlice('\n')
    if e != nil {
        c.parseError(errReadFailed+e.Error(), true)
    }

    if !bytes.HasSuffix(line, lineEnding) {
        c.parseError(errCorrupted+"unexpected line ending", true)
    }

    payload := line[1 : len(line)-2]

    switch line[0] {
    case '+':
        if bytes.Equal(payload, replyOK) {
            return replyOK
        } else if bytes.Equal(payload, replyPong) {
            return replyPong
        } else {
            data := make([]byte, len(payload))
            copy(data, payload)
            return data
        }

    case '-':
        c.parseError(errBase+string(payload), false)

    case ':':
        if n, e := strconv.Atoi(string(payload)); e != nil {
            c.parseError(errCorrupted+e.Error(), true)
        } else {
            return n
        }

    case '$':
        if size, e := strconv.Atoi(string(payload)); e != nil {
            c.parseError(errCorrupted+e.Error(), true)
        } else if size >= 0 {
            data := make([]byte, size+2)
            if _, e := io.ReadFull(c.rw, data); e != nil {
                c.parseError(errCorrupted+e.Error(), true)
            }
            return data[:size]
        }

    case '*':
        if argc, e := strconv.Atoi(string(payload)); e != nil {
            c.parseError(errCorrupted+e.Error(), true)
        } else if argc >= 0 {
            argv := make([]interface{}, argc)
            for i := range argv {
                argv[i] = c.ReadReply()
            }
            return argv
        }

    default:
        c.parseError(errInvalidResp+string(line[:1]), true)
    }
    return nil
}

func (c *Conn) parseError(err string, fatal bool) {
    if fatal {
        c.down = true
    }

    panic(err)
}
