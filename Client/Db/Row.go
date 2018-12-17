package Db

import (
    "database/sql"

    "github.com/pinguo/pgo"
)

// Row wrapper for sql.Row
type Row struct {
    pgo.Object
    row   *sql.Row
    query string
    args  []interface{}
}

func (r *Row) close() {
    r.SetContext(nil)
    r.row = nil
    r.query = ""
    r.args = nil
    rowPool.Put(r)
}

// Scan copies the columns in the current row into the values pointed at by dest.
func (r *Row) Scan(dest ...interface{}) error {
    err := r.row.Scan(dest...)
    if err != nil && err != sql.ErrNoRows {
        r.GetContext().Error("Db.QueryOne error, %s, query:%s, args:%v", err.Error(), r.query, r.args)
    }

    r.close()
    return err
}
