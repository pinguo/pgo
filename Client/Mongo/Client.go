package Mongo

import (
    "fmt"
    "net/url"
    "strings"
    "time"

    "github.com/globalsign/mgo"
)

// Mongo Client component, configuration:
// "mongo": {
//     "class": "@pgo/Client/Mongo/Client",
//     "dsn": "mongodb://host1:port1/[db[?options]]",
//     "connectTimeout": "1s",
//     "readTimeout": "10s",
//     "writeTimeout": "10s"
// }
//
// see Dial() for query options, default:
// replicaSet=
// connect=replicaSet
// maxPoolSize=100
// minPoolSize=1
// maxIdleTimeMS=300000
// ssl=false
// w=1
// j=false
// wtimeoutMS=10000
// readPreference=secondaryPreferred
type Client struct {
    session        *mgo.Session
    dsn            string
    connectTimeout time.Duration
    readTimeout    time.Duration
    writeTimeout   time.Duration
}

func (c *Client) Construct() {
    c.dsn = defaultDsn
    c.connectTimeout = defaultConnectTimeout
    c.readTimeout = defaultReadTimeout
    c.writeTimeout = defaultWriteTimeout
}

func (c *Client) Init() {
    server, query := c.dsn, defaultOptions
    if pos := strings.IndexByte(c.dsn, '?'); pos > 0 {
        dsnOpts, _ := url.ParseQuery(c.dsn[pos+1:])
        options, _ := url.ParseQuery(defaultOptions)

        for k, v := range dsnOpts {
            if len(v) > 0 && len(v[0]) > 0 {
                options.Set(k, v[0])
            }
        }
        server = c.dsn[:pos]
        query = options.Encode()
    }

    c.dsn = server + "?" + query
    dialInfo, e := mgo.ParseURL(c.dsn)
    if e != nil {
        panic(fmt.Sprintf(errInvalidDsn, c.dsn, e.Error()))
    }

    dialInfo.Timeout = c.connectTimeout
    dialInfo.ReadTimeout = c.readTimeout
    dialInfo.WriteTimeout = c.writeTimeout

    if c.session, e = mgo.DialWithInfo(dialInfo); e != nil {
        panic(fmt.Sprintf(errDialFailed, c.dsn, e.Error()))
    }

    c.session.SetMode(mgo.Monotonic, true)
}

func (c *Client) SetDsn(dsn string) {
    c.dsn = dsn
}

func (c *Client) SetConnectTimeout(v string) {
    if connectTimeout, e := time.ParseDuration(v); e != nil {
        panic(fmt.Sprintf(errSetProp, "connectTimeout", e.Error()))
    } else {
        c.connectTimeout = connectTimeout
    }
}

func (c *Client) SetReadTimeout(v string) {
    if readTimeout, e := time.ParseDuration(v); e != nil {
        panic(fmt.Sprintf(errSetProp, "readTimeout", e.Error()))
    } else {
        c.readTimeout = readTimeout
    }
}

func (c *Client) SetWriteTimeout(v string) {
    if writeTimeout, e := time.ParseDuration(v); e != nil {
        panic(fmt.Sprintf(errSetProp, "writeTimeout", e.Error()))
    } else {
        c.writeTimeout = writeTimeout
    }
}

func (c *Client) GetSession() *mgo.Session {
    return c.session.Copy()
}
