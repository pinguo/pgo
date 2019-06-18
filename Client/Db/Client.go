package Db

import (
    "database/sql"
    "fmt"
    "time"

    "github.com/pinguo/pgo/Util"
)

// DB client component, wrapper for database/sql,
// support read-write splitting, configuration:
// db:
//     class:  "@pgo/Client/Db/Client"
//     driver: "mysql"
//     dsn:    "user:pass@tcp(127.0.0.1:3306)/db?charset=utf8&timeout=0.5s"
//     slaves: ["slave1 dsn", "slave2 dsn"]
//     maxIdleConn: 5
//     maxConnTime: "1h"
//     slowLogTime: "100ms"
type Client struct {
    driver string   // driver name
    dsn    string   // master dsn
    slaves []string // slaves dsn

    maxIdleConn int
    maxConnTime time.Duration
    slowLogTime time.Duration

    masterDb *sql.DB   // master db instance
    slaveDbs []*sql.DB // slave db instances
}

func (c *Client) Construct() {
    c.maxIdleConn = 5
    c.maxConnTime = time.Hour
    c.slowLogTime = 100 * time.Millisecond
}

func (c *Client) Init() {
    if c.driver == "" || c.dsn == "" {
        panic("Db: driver and dsn are required")
    }

    if Util.SliceSearchString(sql.Drivers(), c.driver) == -1 {
        panic(fmt.Sprintf("Db: driver %s is not registered", c.driver))
    }

    // create master db instance
    if db, e := sql.Open(c.driver, c.dsn); e != nil {
        panic(fmt.Sprintf("Db: open %s error, %s", c.dsn, e.Error()))
    } else {
        db.SetConnMaxLifetime(c.maxConnTime)
        db.SetMaxIdleConns(c.maxIdleConn)
        c.masterDb = db
    }

    // create slave db instances
    for _, dsn := range c.slaves {
        if db, e := sql.Open(c.driver, dsn); e != nil {
            panic(fmt.Sprintf("Db: open %s error, %s", dsn, e.Error()))
        } else {
            db.SetConnMaxLifetime(c.maxConnTime)
            db.SetMaxIdleConns(c.maxIdleConn)
            c.slaveDbs = append(c.slaveDbs, db)
        }
    }
}

// SetDriver set driver db use, eg. "mysql"
func (c *Client) SetDriver(driver string) {
    c.driver = driver
}

// SetDsn set master dsn, the dsn is driver specified,
// eg. dsn format for github.com/go-sql-driver/mysql is
// [username[:password]@][protocol[(address)]]/dbname[?param=value]
func (c *Client) SetDsn(dsn string) {
    c.dsn = dsn
}

// SetSlaves set dsn for slaves
func (c *Client) SetSlaves(v []interface{}) {
    for _, vv := range v {
        c.slaves = append(c.slaves, vv.(string))
    }
}

// SetMaxIdleConn set max idle conn, default is 5
func (c *Client) SetMaxIdleConn(maxIdleConn int) {
    c.maxIdleConn = maxIdleConn
}

// SetMaxConnTime set conn life time, default is 1h
func (c *Client) SetMaxConnTime(v string) {
    if maxConnTime, err := time.ParseDuration(v); err != nil {
        panic("Db.SetMaxConnTime error, " + err.Error())
    } else {
        c.maxConnTime = maxConnTime
    }
}

// SetSlowTime set slow log time, default is 100ms
func (c *Client) SetSlowLogTime(v string) {
    if slowLogTime, err := time.ParseDuration(v); err != nil {
        panic("Db.SetSlowLogTime error, " + err.Error())
    } else {
        c.slowLogTime = slowLogTime
    }
}

// GetDb get a master or slave db instance
func (c *Client) GetDb(master bool) *sql.DB {
    if num := len(c.slaveDbs); !master && num > 0 {
        idx := 0
        if num > 1 {
            idx = (time.Now().Nanosecond() / 1000) % num
        }

        return c.slaveDbs[idx]
    }

    return c.masterDb
}
