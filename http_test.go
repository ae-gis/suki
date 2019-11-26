/*  http_test.go
*
* @Author:             Nanang Suryadi
* @Date:               November 21, 2019
* @Last Modified by:   @suryakencana007
* @Last Modified time: 21/11/19 17:41
 */

package suki

import (
        "fmt"
        "html"
        "net/http"
        "os"
        "sync"
        "testing"

        "github.com/go-chi/chi"
        "github.com/spf13/cobra"
        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
)

const (
        Port         int = 8081
        ReadTimeout  int = 5
        WriteTimeout int = 100
)

func TestNewHttp(t *testing.T) {
        var wg sync.WaitGroup
        stop := make(chan bool)
        wg.Add(1)
        defer wg.Wait()

        httpSignal := NewCmdHttpSignaled(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                _, err := fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
                assert.NoError(t, err)
        }), Port+1, ReadTimeout, WriteTimeout, stop)

        go func() {
                defer wg.Done()
                err := httpSignal.GetCmd().Execute()
                assert.NoError(t, err)
        }()

        stop <- true
}

func TestHttp(t *testing.T) {
        var wg sync.WaitGroup
        assert := require.New(t)
        stop := make(chan bool)
        wg.Add(1)

        cmd := NewCmdHttpSignaled(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                _, err := fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
                assert.NoError(err)
        }), Port+2, ReadTimeout, WriteTimeout, stop).GetCmd()

        go func() {
                defer wg.Done()
                _, err := cmd.ExecuteC()
                assert.NoError(err)
        }()

        stop <- true
        wg.Wait()
}

func TestNewHttpCmdWithFilename(t *testing.T) {
        var wg sync.WaitGroup
        wg.Add(1)
        defer wg.Wait()

        stop := make(chan bool)

        cmd := NewCmdHttpSignaled(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                _, err := fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
                assert.NoError(t, err)
        }), Port+3, ReadTimeout, WriteTimeout, stop).GetCmd()
        go func() {
                defer wg.Done()
                _, err := cmd.ExecuteC()
                assert.NoError(t, err)
        }()

        stop <- true
        os.Args = []string{""}
}

func TestListenAndServe(t *testing.T) {
        var (
                err error
                wg  sync.WaitGroup
        )
        stop := make(chan bool)

        cc := &cmdHttp{stop: stop, Port: Port+4}
        cc.Cmd = &cobra.Command{
                Use:   "http",
                Short: "Used to run the http service",
                RunE: func(cmd *cobra.Command, args []string) (err error) {
                        mux := chi.NewMux()
                        return cc.handlerFunc(mux)
                },
        }

        wg.Add(1)
        go func() {
                defer wg.Done()
                err = cc.Cmd.Execute()
        }()
        assert.NoError(t, err)
        stop <- true
        wg.Wait()
}

