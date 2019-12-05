/*  store.go
*
* @Author:             Nanang Suryadi
* @Date:               November 24, 2019
* @Last Modified by:   @suryakencana007
* @Last Modified time: 24/11/19 09:00
 */

package sqlx

import (
        "context"
        "database/sql"
        "fmt"
        "strings"
        "time"

        "github.com/ae-gis/suki"
        _ "github.com/lib/pq"
)

const POSTGRES string = "postgres"

type TxArgs struct {
        Query string
        Args  []interface{}
}

type Factory interface {
        Close() error
        BeginCtx() (context.Context, context.CancelFunc)
        BeginTx(ctx context.Context) (*sql.Tx, context.CancelFunc)
        QueryCtx(ctx context.Context, fn func(rs *sql.Rows) error, query string, args ...interface{}) error
        QueryRowCtx(ctx context.Context, fn func(rs *sql.Row) error, query string, args ...interface{}) error
        ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
        TxExecContextWithID(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (ids interface{}, err error)
        TxExecContext(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (affected int64, err error)
        TxCommit(ctx context.Context, tx *sql.Tx) error
        WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error
        PrepareContext(ctx context.Context, query string) (stmt *sql.Stmt, err error)
}

type DB struct {
        *sql.DB
        RetryCount int
        Timeout    int
        Concurrent int
}

func (r *DB) Close() error {
        return r.DB.Close()
}

func (r *DB) BeginCtx() (context.Context, context.CancelFunc) {
        return context.WithTimeout(context.Background(), time.Duration(r.Timeout)*time.Second)
}

func (r *DB) BeginTx(ctx context.Context) (*sql.Tx, context.CancelFunc) {
        c, cancel := context.WithTimeout(ctx, time.Duration(r.Timeout)*time.Second)
        tx, err := r.DB.BeginTx(c, &sql.TxOptions{Isolation: sql.LevelSerializable})
        if err != nil {
                panic(err)
        }
        return tx, cancel
}

func (r *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
        return r.DB.ExecContext(ctx, query, args...)
}

func (r *DB) PrepareContext(ctx context.Context, query string) (stmt *sql.Stmt, err error) {
        return r.DB.PrepareContext(ctx, query)
}

func (r *DB) QueryRowCtx(ctx context.Context, fn func(rs *sql.Row) error, query string, args ...interface{}) error {
        if r.DB == nil {
                suki.Error("the database connection is nil",
                        suki.Field("query", query),
                        suki.Field("args", args))
                return fmt.Errorf("cannot access your db connection")
        }
        rs := r.DB.QueryRowContext(ctx, query, args...)
        if err := fn(rs); err != nil {
                if err == sql.ErrNoRows {
                        suki.Warn("result not found",
                                suki.Field("query", query),
                                suki.Field("args", args))
                        return nil
                }
                suki.Error("query row failed",
                        suki.Field("query", query),
                        suki.Field("args", args))
                return err
        }
        return nil
}

func (r *DB) QueryCtx(ctx context.Context, fn func(rs *sql.Rows) error, query string, args ...interface{}) error {
        if r.DB == nil {
                suki.Error("the database connection is nil",
                        suki.Field("query", query),
                        suki.Field("args", args))
                return fmt.Errorf("cannot access your db connection")
        }
        rs, err := r.DB.QueryContext(ctx, query, args...)
        if err != nil {
                suki.Warn("query failed",
                        suki.Field("query", query),
                        suki.Field("args", args))
                return err
        }
        defer func() {
                err = rs.Close()
        }()

        if err := fn(rs); err != nil {
                if err == sql.ErrNoRows {
                        suki.Warn("result not found",
                                suki.Field("query", query),
                                suki.Field("args", args))
                        return nil
                }
                suki.Error("query row failed",
                        suki.Field("query", query),
                        suki.Field("args", args))
                return err
        }
        return nil
}

func (r *DB) TxExecContextWithID(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (ids interface{}, err error) {
        if strings.Contains(query, "RETURNING id") {
                stmt, err := tx.PrepareContext(ctx, query)
                if err != nil {
                        suki.Error("ExecContextWithID:",
                                suki.Field("error", err.Error()),
                                suki.Field("query", query),
                                suki.Field("args", args),
                        )
                        return nil, err
                }
                if err := stmt.QueryRowContext(ctx, args...).Scan(&ids); err != nil {
                        suki.Error("ExecContextWithID:",
                                suki.Field("error", err.Error()),
                                suki.Field("query", query),
                                suki.Field("args", args),
                        )
                        return nil, err
                }
                err = stmt.Close()
                if err != nil {
                        suki.Error("ExecContextWithID:",
                                suki.Field("error", err.Error()),
                                suki.Field("query", query),
                                suki.Field("args", args),
                        )
                        return nil, err
                }
                return ids, nil
        }
        err = fmt.Errorf("query has no RETUNING id syntax")
        suki.Error("ExecContextWithID:",
                suki.Field("error", err.Error()),
                suki.Field("query", query),
                suki.Field("args", args),
        )
        return nil, err
}

func (r *DB) TxExecContext(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (affected int64, err error) {
        stmt, err := tx.PrepareContext(ctx, query)
        if err != nil {
                suki.Error("ExecContextWithID:",
                        suki.Field("error", err.Error()),
                        suki.Field("query", query),
                        suki.Field("args", args),
                )
                return affected, err
        }
        result, err := stmt.ExecContext(ctx, args...)
        if err != nil {
                suki.Error("ExecContextWithID:",
                        suki.Field("error", err.Error()),
                        suki.Field("query", query),
                        suki.Field("args", args),
                )
                return affected, err
        }
        affected, err = result.RowsAffected()
        if err != nil {
                suki.Error("TxExecContext: RowsAffected",
                        suki.Field("error", err.Error()),
                )
                return affected, err
        }
        err = stmt.Close()
        if err != nil {
                suki.Error("TxExecContext:",
                        suki.Field("error", err.Error()),
                        suki.Field("query", query),
                        suki.Field("args", args),
                )
                return affected, err
        }
        return affected, nil

}

func (r *DB) TxCommit(ctx context.Context, tx *sql.Tx) error {
        // commit db transaction
        if err := tx.Commit(); err != nil {
                if err = tx.Rollback(); err != nil {
                        suki.Error("TxCommit:",
                                suki.Field("error", err.Error()),
                        )
                        return err
                } // rollback if fail query statement
                return err
        }
        return nil
}

func (r *DB) WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
        tx, cancel := r.BeginTx(ctx)
        defer cancel()
        return fn(tx)
}

func New(driverName, connString string, retryCount, timeout, concurrent int) (*DB, error) {
        db, err := sql.Open(driverName, connString)
        if err != nil {
                suki.Error(err.Error())
                panic(fmt.Errorf("cannot access your db connection").Error())
        }
        return &DB{
                db,
                retryCount,
                timeout,
                concurrent,
        }, nil
}
