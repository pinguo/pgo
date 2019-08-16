package RabbitMq

import (
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
    "sync"
    "time"

    "github.com/pinguo/pgo"
    "github.com/pinguo/pgo/Util"
    "github.com/streadway/amqp"
)

func newConnBox(id, addr, dsn string, maxChannelNum int, tlsDft ...string) *ConnBox {
    tlsCert, tlsCertKey, tlsRootCAs := "", "", ""
    if len(tlsDft) > 0 {
        tlsCert = tlsDft[0]
        tlsCertKey = tlsDft[1]
        tlsRootCAs = tlsDft[2]
    }

    connBox := &ConnBox{id: id, addr: addr, dsn: dsn, tlsCert: tlsCert, tlsCertKey: tlsCertKey, tlsRootCAs: tlsRootCAs, maxChannelNum: maxChannelNum}
    connBox.initConn()

    return connBox
}

type ConnBox struct {
    id              string
    addr            string
    useConnCount    int
    useChannelCount int
    channelList     chan *ChannelBox
    lock            sync.RWMutex
    newChannelLock  sync.RWMutex
    startTime       time.Time

    connection *amqp.Connection
    tlsCert    string
    tlsCertKey string
    tlsRootCAs string
    dsn        string

    maxChannelNum int

    notifyClose chan *amqp.Error

    disable bool
}

func (c *ConnBox) setEnable() {
    c.lock.Lock()
    defer c.lock.Unlock()
    c.disable = false
}

func (c *ConnBox) setDisable() {
    c.lock.Lock()
    defer c.lock.Unlock()
    c.disable = true
    c.close()
}

func (c *ConnBox) initConn() {
    func() {
        c.lock.Lock()
        defer c.lock.Unlock()
        var err error

        if c.tlsCert != "" && c.tlsCertKey != "" {
            cfg := new(tls.Config)
            if c.tlsRootCAs != "" {
                cfg.RootCAs = x509.NewCertPool()
                if ca, err := ioutil.ReadFile(c.tlsRootCAs); err == nil {
                    cfg.RootCAs.AppendCertsFromPEM(ca)
                }
            }

            if cert, err := tls.LoadX509KeyPair(c.tlsCert, c.tlsCertKey); err == nil {
                cfg.Certificates = append(cfg.Certificates, cert)
            }

            c.connection, err = amqp.DialTLS(c.dsn, cfg)
        } else {
            c.connection, err = amqp.Dial(c.dsn)
        }

        if err != nil {
            panic("Failed to connect to RabbitMQ:" + err.Error())
        }

        c.disable = false
        c.channelList = make(chan *ChannelBox, c.maxChannelNum)
        c.notifyClose = make(chan *amqp.Error)
        c.startTime = time.Now()
        c.connection.NotifyClose(c.notifyClose)
    }()

    go c.check(c.startTime)
}

func (c *ConnBox) check(startTime time.Time) {
    defer func() {
        if err := recover(); err != nil {
            pgo.GLogger().Error("Rabbit ConnBox.check err:" + Util.ToString(err))
        }
    }()

    for {
        if c.startTime != startTime {
            // 自毁
            return
        }

        select {
        case err, ok := <-c.notifyClose:
            if ok == false {
                return
            }

            if err != nil {
                func() {
                    defer func() {
                        if err := recover(); err != nil {
                            pgo.GLogger().Error("Rabbit ConnBox.check start initConn err:" + Util.ToString(err))
                        }
                    }()

                    c.setDisable()
                    c.initConn()
                }()
                return
            }

        default:
            time.Sleep(100 * time.Microsecond)

        }
    }
}

func (c *ConnBox) isClosed() bool {
    if c.disable || c.connection.IsClosed() {
        return true
    }
    return false
}

func (c *ConnBox) close() {
    if c.connection != nil && c.connection.IsClosed() == false {
        err := c.connection.Close()
        if err != nil {
            pgo.GLogger().Warn("Rabbit ConnBox.close err:" + err.Error())
        }
    }
}
