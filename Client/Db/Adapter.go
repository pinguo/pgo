package Db

import (
    "context"
    "database/sql"
    "strings"
    "time"

    "github.com/pinguo/pgo"
    "github.com/pinguo/pgo/Util"
)

// Adapter of Db Client, add context support.
// usage: db := this.GetObject(Db.AdapterClass).(*Db.Adapter)
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
    // reuse previous db instance for read
    if !master && a.db != nil {
        return a.db
    }

    a.db = a.client.GetDb(master)
    return a.db
}

// Begin start a transaction with default timeout context and optional opts,
// if opts is nil, default driver option will be used.
func (a *Adapter) Begin(opts ...*sql.TxOptions) bool {
    opts = append(opts, nil)
    ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
    return a.BeginContext(ctx, opts[0])
}

// BeginContext start a transaction with specified context and optional opts,
// if opts is nil, default driver option will be used.
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

// QueryOne perform one row query using a default timeout context,
// and always returns a non-nil value, Errors are deferred until
// Row's Scan method is called.
func (a *Adapter) QueryOne(query string, args ...interface{}) *Row {
    ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
    return a.QueryOneContext(ctx, query, args...)
}

// QueryOneContext perform one row query using a specified context,
// and always returns a non-nil value, Errors are deferred until
// Row's Scan method is called.
func (a *Adapter) QueryOneContext(ctx context.Context, query string, args ...interface{}) *Row {
    start := time.Now()
    defer func() {
        elapse := time.Since(start)
        a.ProfileAdd("Db.QueryOne", elapse, query, args...)

        if elapse >= a.client.slowLogTime && a.client.slowLogTime > 0 {
            a.GetContext().Warn("Db.QueryOne slow, elapse:%s, query:%s, args:%v", elapse, query, args)
        }
    }()

    var row *sql.Row
    if a.tx != nil {
        row = a.tx.QueryRowContext(ctx, query, args...)
    } else {
        row = a.GetDb(false).QueryRowContext(ctx, query, args...)
    }

    // wrap row for profile purpose
    rowWrapper := rowPool.Get().(*Row)
    rowWrapper.SetContext(a.GetContext())
    rowWrapper.row = row
    rowWrapper.query = query
    rowWrapper.args = args

    return rowWrapper
}

// Query perform query using a default timeout context.
func (a *Adapter) Query(query string, args ...interface{}) *sql.Rows {
    ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
    return a.QueryContext(ctx, query, args...)
}

// QueryContext perform query using a specified context.
func (a *Adapter) QueryContext(ctx context.Context, query string, args ...interface{}) *sql.Rows {
    start := time.Now()
    defer func() {
        elapse := time.Since(start)
        a.ProfileAdd("Db.Query", elapse, query, args...)

        if elapse >= a.client.slowLogTime && a.client.slowLogTime > 0 {
            a.GetContext().Warn("Db.Query slow, elapse:%s, query:%s, args:%v", elapse, query, args)
        }
    }()

    var rows *sql.Rows
    var err error

    if a.tx != nil {
        rows, err = a.tx.QueryContext(ctx, query, args...)
    } else {
        rows, err = a.GetDb(false).QueryContext(ctx, query, args...)
    }

    if err != nil {
        a.GetContext().Error("Db.Query error, %s, query:%s, args:%v", err.Error(), query, args)
        return nil
    }

    return rows
}

// Exec perform exec using a default timeout context.
func (a *Adapter) Exec(query string, args ...interface{}) sql.Result {
    ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
    return a.ExecContext(ctx, query, args...)
}

// ExecContext perform exec using a specified context.
func (a *Adapter) ExecContext(ctx context.Context, query string, args ...interface{}) sql.Result {
    start := time.Now()
    defer func() {
        elapse := time.Since(start)
        a.ProfileAdd("Db.Exec", elapse, query, args...)

        if elapse >= a.client.slowLogTime && a.client.slowLogTime > 0 {
            a.GetContext().Warn("Db.Exec slow, elapse:%s, query:%s, args:%v", elapse, query, args)
        }
    }()

    var res sql.Result
    var err error

    if a.tx != nil {
        res, err = a.tx.ExecContext(ctx, query, args...)
    } else {
        res, err = a.GetDb(true).ExecContext(ctx, query, args...)
    }

    if err != nil {
        a.GetContext().Error("Db.Exec error, %s, query:%s, args:%v", err.Error(), query, args)
        return nil
    }

    return res
}

// Prepare creates a prepared statement for later queries or executions,
// the Close method must be called by caller.
func (a *Adapter) Prepare(query string) *Stmt {
    ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
    return a.PrepareContext(ctx, query)
}

// PrepareContext creates a prepared statement for later queries or executions,
// the Close method must be called by caller.
func (a *Adapter) PrepareContext(ctx context.Context, query string) *Stmt {
    var stmt *sql.Stmt
    var err error

    if a.tx != nil {
        stmt, err = a.tx.PrepareContext(ctx, query)
    } else {
        master, pos := true, strings.IndexByte(query, ' ')
        if pos != -1 && strings.ToUpper(query[:pos]) == "SELECT" {
            master = false
        }

        stmt, err = a.GetDb(master).PrepareContext(ctx, query)
    }

    if err != nil {
        a.GetContext().Error("Db.Prepare error, %s, query:%s", err.Error(), query)
        return nil
    }

    // wrap stmt for profile purpose
    stmtWrapper := stmtPool.Get().(*Stmt)
    stmtWrapper.SetContext(a.GetContext())
    stmtWrapper.stmt = stmt
    stmtWrapper.client = a.client
    stmtWrapper.query = query

    return stmtWrapper
}

func (a *Adapter) ProfileAdd(key string, elapse time.Duration, query string, args ...interface{}) {
    newKey := key
    if a.client.sqlLog == true {
        if len(args) > 0 {
            for _, v := range args {
                query = strings.Replace(query, "?", Util.ToString(v), 1)
            }
        }

        newKey = key + "(" + query + ")"
    }

    a.GetContext().ProfileAdd(newKey, elapse)
}
