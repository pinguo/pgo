package Db

import (
    "context"
    "database/sql"
    "time"

    "github.com/pinguo/pgo"
)

// Adapter of Db Client, add context support.
// usage: db := this.GetObject("@pgo/Client/Db/Adapter").(*Adapter)
type Adapter struct {
    pgo.Object
    client *Client
    db     *sql.DB
    tx     *sql.Tx
}

func (a *Adapter) Construct(componentId ...string) {
    id := defaultComponentId
    if len(componentId) > 0 {
        id = componentId[0]
    }

    a.client = pgo.App.Get(id).(*Client)
}

func (a *Adapter) GetClient() *Client {
    return a.client
}

func (a *Adapter) GetDb(master bool) *sql.DB {
    // reuse previous db instance for slave
    if !master && a.db != nil {
        return a.db
    }

    a.db = a.client.GetDb(master)
    return a.db
}

// Begin start a transaction with default timeout context and optional opts,
// if opts is nil, defaults will be used.
func (a *Adapter) Begin(opts ...*sql.TxOptions) bool {
    opts = append(opts, nil)
    ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
    return a.BeginContext(ctx, opts[0])
}

// BeginContext start a transaction with specified context and optional opts,
// if opts is nil, defaults will be used.
func (a *Adapter) BeginContext(ctx context.Context, opts *sql.TxOptions) bool {
    if tx, e := a.GetDb(true).BeginTx(ctx, opts); e != nil {
        a.GetContext().Error("Db.Begin error, " + e.Error())
        return false
    } else {
        a.tx = tx
        return true
    }
}

// Commit commit transaction that previously started.
func (a *Adapter) Commit() bool {
    if a.tx == nil {
        a.GetContext().Error("Db.Commit not in transaction")
        return false
    } else {
        if e := a.tx.Commit(); e != nil {
            a.GetContext().Error("Db.Commit error, " + e.Error())
            return false
        }
        return true
    }
}

// Rollback roll back transaction that previously started.
func (a *Adapter) Rollback() bool {
    if a.tx == nil {
        a.GetContext().Error("Db.Rollback not in transaction")
        return false
    } else {
        if e := a.tx.Rollback(); e != nil {
            a.GetContext().Error("Db.Rollback error, " + e.Error())
            return false
        }
        return true
    }
}

// InTransaction check if adapter is in transaction.
func (a *Adapter) InTransaction() bool {
    return a.tx != nil
}

// Query perform query using a default timeout context.
func (a *Adapter) Query(query string, args ...interface{}) *sql.Rows {
    ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
    return a.QueryContext(ctx, query, args...)
}

// QueryContext perform query using a specified context.
func (a *Adapter) QueryContext(ctx context.Context, query string, args ...interface{}) *sql.Rows {
    pgoCtx, start := a.GetContext(), time.Now()
    defer func() {
        elapse := time.Since(start)
        pgoCtx.ProfileAdd("Db.Query", elapse)

        if elapse >= a.client.slowLogTime && a.client.slowLogTime > 0 {
            pgoCtx.Warn("Db.Query slow, elapse:%s, query:%s, args:%v", elapse, query, args)
        }
    }()

    var rows *sql.Rows
    var err error

    if a.tx != nil {
        rows, err = a.tx.QueryContext(ctx, query, args...)
    } else {
        rows, err = a.GetDb(false).QueryContext(ctx, query, args...)
    }

    if err == nil {
        return rows
    }

    pgoCtx.Error("Db.Query failed, error:%s, query:%s, args:%v", err.Error(), query, args)
    return nil
}

// Exec perform exec using a default timeout context.
func (a *Adapter) Exec(query string, args ...interface{}) sql.Result {
    ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
    return a.ExecContext(ctx, query, args...)
}

// ExecContext perform exec using a specified context.
func (a *Adapter) ExecContext(ctx context.Context, query string, args ...interface{}) sql.Result {
    pgoCtx, start := a.GetContext(), time.Now()
    defer func() {
        elapse := time.Since(start)
        pgoCtx.ProfileAdd("Db.Exec", elapse)

        if elapse >= a.client.slowLogTime && a.client.slowLogTime > 0 {
            pgoCtx.Warn("Db.Exec slow, elapse:%s, query:%s, args:%v", elapse, query, args)
        }
    }()

    var res sql.Result
    var err error

    if a.tx != nil {
        res, err = a.tx.ExecContext(ctx, query, args...)
    } else {
        res, err = a.GetDb(true).ExecContext(ctx, query, args...)
    }

    if err == nil {
        return res
    }

    pgoCtx.Error("Db.Exec failed, err:%s, query:%s, args:%v", err.Error(), query, args)
    return nil
}
