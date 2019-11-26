/*  store_test.go
*
* @Author:             Nanang Suryadi
* @Date:               November 24, 2019
* @Last Modified by:   @suryakencana007
* @Last Modified time: 24/11/19 14:00
 */

package sqlx

import (
        "context"
        "database/sql"
        "fmt"
        "testing"
        "time"

        "gitlab.com/suryakencana007/suki"

        "github.com/lib/pq"
        "github.com/ory/dockertest"
        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/suite"
)

const (
        pgSchema = `CREATE TABLE IF NOT EXISTS users (
id integer NOT NULL, 
name varchar(255) NOT NULL
);`
)

type ConnPGSuite struct {
        suite.Suite
        DB       *DB
        pool     *dockertest.Pool
        Resource *dockertest.Resource
}

func (s *ConnPGSuite) GetResource() *dockertest.Resource {
        return s.Resource
}

func (s *ConnPGSuite) SetResource(resource *dockertest.Resource) {
        s.Resource = resource
}

func (s *ConnPGSuite) GetPool() *dockertest.Pool {
        return s.pool
}

func (s *ConnPGSuite) SetPool(pool *dockertest.Pool) {
        s.pool = pool
}

func (s *ConnPGSuite) GetDB() *DB {
        return s.DB
}

func (s *ConnPGSuite) SetDB(db *DB) {
        s.DB = db
}

func (s *ConnPGSuite) SetupTest() {
        var err error
        s.pool, err = dockertest.NewPool("")
        if err != nil {
                panic(fmt.Sprintf("could not connect to docker: %s\n", err))
        }
        err = NewPoolPG(s)
        if err != nil {
                panic(fmt.Sprintf("prepare pg with docker: %v\n", err))
        }
}

func (s *ConnPGSuite) TearDownTest() {
        if err := s.DB.Close(); err != nil {
                panic(fmt.Sprintf("could not db close: %v\n", err))
        }
        if err := s.pool.RemoveContainerByName("pg_test"); err != nil {
                panic(fmt.Sprintf("could not remove postgres container: %v\n", err))
        }
}

func (s *ConnPGSuite) TestMainCommitInFailedTransaction() {
        t := s.T()
        txn, cancel := s.DB.BeginTx(context.Background())
        defer cancel()
        rows, err := txn.Query("SELECT error")
        assert.Error(t, err)
        if err == nil {
                rows.Close()
                t.Fatal("expected failure")
        }
        err = txn.Commit()
        assert.Error(t, err)
        if err != pq.ErrInFailedTransaction {
                t.Fatalf("expected ErrInFailedTransaction; got %#v", err)
        }
}

func (s *ConnPGSuite) TestExecContext() {
        t := s.T()
        ctx, cancel := s.DB.BeginCtx()
        defer cancel()
        args := []interface{}{
                1003,
                "TEST WithTransaction Func",
        }
        query := `INSERT INTO users (id, name) VALUES ($1, $2)`
        result, err := s.DB.ExecContext(ctx, query, args...)
        assert.NoError(t, err)
        ids, err := result.RowsAffected()
        assert.NoError(t, err)
        assert.Equal(t, 1, int(ids))
}

func (s *ConnPGSuite) TestPrepareContext() {
        t := s.T()
        ctx, cancel := s.DB.BeginCtx()
        defer cancel()
        args := []interface{}{
                1003,
                "TEST WithTransaction Func",
        }
        query := `INSERT INTO users (id, name) VALUES ($1, $2)`
        stmt, err := s.DB.PrepareContext(ctx, query)
        assert.NoError(t, err)
        result, err := stmt.ExecContext(ctx, args...)
        assert.NoError(t, err)
        ids, err := result.RowsAffected()
        assert.NoError(t, err)
        assert.Equal(t, 1, int(ids))
}

func (s *ConnPGSuite) TestQueryCtxFailed() {
        t := s.T()
        ctx, cancel := s.DB.BeginCtx()
        defer cancel()
        err := s.DB.QueryCtx(ctx, func(rows *sql.Rows) error {
                return nil
        }, "SELECT error")
        assert.Error(t, err)
}

func (s *ConnPGSuite) TestGetUserID() {
        t := s.T()
        names := make([]string, 0)
        ctx, cancel := s.DB.BeginCtx()
        defer cancel()
        err := s.DB.QueryCtx(ctx, func(rows *sql.Rows) error {
                for rows.Next() {
                        var name string
                        if err := rows.Scan(&name); err != nil {
                                t.Fatal(err)
                                return err
                        }
                        names = append(names, name)
                }
                assert.NoError(t, rows.Err())
                assert.IsType(t, []string{}, names)
                return nil
        }, "SELECT id FROM users")
        assert.NoError(t, err)
}

func (s *ConnPGSuite) TestWithTransaction() {
        t := s.T()
        ctx := context.Background()
        args := []interface{}{
                1003,
                "TEST WithTransaction Func",
        }
        query := `INSERT INTO users (id, name) VALUES ($1, $2) RETURNING id`
        err := s.DB.WithTransaction(ctx, func(tx *sql.Tx) error {
                affected, err := s.DB.TxExecContext(ctx, tx, query, args...)
                suki.Info("WithTransaction",
                        suki.Field("Affected", affected),
                )
                if err != nil {
                        return err
                }
                ids, err := s.DB.TxExecContextWithID(ctx, tx, query, args...)
                suki.Info("WithTransaction",
                        suki.Field("LastInsertID", ids),
                )
                if err != nil {
                        return err
                }
                return s.DB.TxCommit(ctx, tx)
        })
        assert.NoError(t, err)
}

func (s *ConnPGSuite) TestWithTransactionFail() {
        t := s.T()
        ctx := context.Background()
        args := []interface{}{
                1001,
                "TEST WithTransaction Func",
        }
        query := `INSERT INTO users (id, name) VALUES (?, $2)`
        err := s.DB.WithTransaction(ctx, func(tx *sql.Tx) error {
                affected, err := s.DB.TxExecContext(ctx, tx, query, args...)
                suki.Info("WithTransaction",
                        suki.Field("Affected", affected),
                )
                if err != nil {
                        return err
                }
                return s.DB.TxCommit(ctx, tx)
        })
        assert.Error(t, err)
}

func (s *ConnPGSuite) TestContextTimeOutFail() {
        t := s.T()
        ctx := context.Background()
        args := make([]interface{}, 0)
        query := `SELECT pg_sleep(5)`
        err := s.DB.WithTransaction(ctx, func(tx *sql.Tx) error {
                affected, err := s.DB.TxExecContext(ctx, tx, query, args...)
                suki.Info("WithTransaction",
                        suki.Field("Affected", affected),
                )
                return err
        })
        assert.Error(t, err)
}

func (s *ConnPGSuite) TestMainDB() {
        assert.IsType(s.T(), &DB{}, s.GetDB())
        assert.Equal(s.T(), 100010, getServerVersion(s.T(), s.GetDB()))
}

func TestMainPGSuite(t *testing.T) {
        suite.Run(t, new(ConnPGSuite))
}

type ConnectionSuite interface {
        T() *testing.T
        GetResource() *dockertest.Resource
        SetResource(resource *dockertest.Resource)
        GetPool() *dockertest.Pool
        SetPool(pool *dockertest.Pool)
        GetDB() *DB
        SetDB(factory *DB)
}

func NewPoolPG(c ConnectionSuite) (err error) {
        t := c.T()
        resource, err := c.GetPool().RunWithOptions(
                &dockertest.RunOptions{
                        Name:       "pg_test",
                        Repository: "postgres",
                        Tag:        "10-alpine",
                        Env: []string{
                                "POSTGRES_PASSWORD=root",
                                "POSTGRES_USER=root",
                                "POSTGRES_DB=dev",
                        },
                })
        c.SetResource(resource)
        if err != nil {
                return fmt.Errorf("%v", err.Error())
        }
        err = c.GetResource().Expire(5)
        assert.NoError(t, err)
        purge := func() error {
                return c.GetPool().Purge(c.GetResource())
        }

        if err := c.GetPool().Retry(func() error {
                connInfo := fmt.Sprintf(`postgresql://%s:%s@%s:%s/%s?sslmode=disable`,
                        "root",
                        "root",
                        "localhost",
                        c.GetResource().GetPort("5432/tcp"),
                        "dev",
                )
                db, err := New(
                        POSTGRES,
                        connInfo,
                        3,
                        5,
                        500,
                )
                if err != nil {
                        panic(err.Error())
                }
                c.SetDB(db)
                return c.GetDB().Ping()
        }); err != nil {
                _ = purge()
                return fmt.Errorf("check connection %v", err.Error())
        }
        if _, err := c.GetDB().Exec(pgSchema); err != nil {
                _ = purge()
                return fmt.Errorf("failed to create schema %v", err.Error())
        }

        return nil
}

func getServerVersion(t *testing.T, db *DB) int {
        var (
                version int
        )
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        err := db.QueryRowCtx(ctx, func(rs *sql.Row) error {
                return rs.Scan(&version)
        }, `SHOW server_version_num;`)
        if err != nil {
                t.Log(err)
        }
        return version
}
