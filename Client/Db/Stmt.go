package Db

import (
    "context"
    "database/sql"
    "time"

    "github.com/pinguo/pgo"
)

// Stmt wrap sql.Stmt, add context support
type Stmt struct {
    pgo.Object
    stmt   *sql.Stmt
    client *Client
    query  string
}

// Close close sql.Stmt and return instance to pool
func (s *Stmt) Close() {
    s.SetContext(nil)
    s.stmt.Close()
    s.stmt = nil
    s.query = ""
    stmtPool.Put(s)
}

// QueryOne perform one row query using a default timeout context,
// and always returns a non-nil value, Errors are deferred until
// Row's Scan method is called.
func (s *Stmt) QueryOne(args ...interface{}) *Row {
    ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
    return s.QueryOneContext(ctx, args...)
}

// QueryOneContext perform one row query using a specified context,
// and always returns a non-nil value, Errors are deferred until
// Row's Scan method is called.
func (s *Stmt) QueryOneContext(ctx context.Context, args ...interface{}) *Row {
    start := time.Now()
    defer func() {
        elapse := time.Since(start)
        s.GetContext().ProfileAdd("Db.StmtQueryOne", elapse)

        if elapse >= s.client.slowLogTime && s.client.slowLogTime > 0 {
            s.GetContext().Warn("Db.StmtQueryOne slow, elapse:%s, query:%s, args:%v", elapse, s.query, args)
        }
    }()

    row := s.stmt.QueryRowContext(ctx, args...)

    // wrap row for profile purpose
    rowWrapper := rowPool.Get().(*Row)
    rowWrapper.SetContext(s.GetContext())
    rowWrapper.row = row
    rowWrapper.query = s.query
    rowWrapper.args = args

    return rowWrapper
}

// Query perform query using a default timeout context.
func (s *Stmt) Query(args ...interface{}) *sql.Rows {
    ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
    return s.QueryContext(ctx, args...)
}

// QueryContext perform query using a specified context.
func (s *Stmt) QueryContext(ctx context.Context, args ...interface{}) *sql.Rows {
    start := time.Now()
    defer func() {
        elapse := time.Since(start)
        s.GetContext().ProfileAdd("Db.StmtQuery", elapse)

        if elapse >= s.client.slowLogTime && s.client.slowLogTime > 0 {
            s.GetContext().Warn("Db.StmtQuery slow, elapse:%s, query:%s, args:%v", elapse, s.query, args)
        }
    }()

    rows, err := s.stmt.QueryContext(ctx, args...)
    if err != nil {
        s.GetContext().Error("Db.StmtQuery error, %s, query:%s, args:%v", err.Error(), s.query, args)
        return nil
    }

    return rows
}

// Exec perform exec using a default timeout context.
func (s *Stmt) Exec(args ...interface{}) sql.Result {
    ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
    return s.ExecContext(ctx, args...)
}

// ExecContext perform exec using a specified context.
func (s *Stmt) ExecContext(ctx context.Context, args ...interface{}) sql.Result {
    start := time.Now()
    defer func() {
        elapse := time.Since(start)
        s.GetContext().ProfileAdd("Db.StmtExec", elapse)

        if elapse >= s.client.slowLogTime && s.client.slowLogTime > 0 {
            s.GetContext().Warn("Db.StmtExec slow, elapse:%s, query:%s, args:%v", elapse, s.query, args)
        }
    }()

    res, err := s.stmt.ExecContext(ctx, args...)
    if err != nil {
        s.GetContext().Error("Db.StmtExec error, %s, query:%s, args:%v", err.Error(), s.query, args)
        return nil
    }

    return res
}
