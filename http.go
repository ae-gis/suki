/*  http.go
*
* @Author:             Nanang Suryadi
* @Date:               November 21, 2019
* @Last Modified by:   @suryakencana007
* @Last Modified time: 21/11/19 12:28
 */

package suki

import (
        "context"
        "fmt"
        "net"
        "net/http"
        "net/url"
        "os"
        "os/signal"
        "strings"
        "sync"
        "syscall"
        "time"

        "github.com/go-chi/chi"
        "github.com/spf13/cobra"
        "google.golang.org/grpc"
)

func Velkommen() string {
        return `
========================================================================================
   _     _     _     _     _  
  / \   / \   / \   / \   / \ 
 ( s ) ( u ) ( k ) ( i ) ( ~ )
  \_/   \_/   \_/   \_/   \_/ 
========================================================================================
- port    : %d
-----------------------------------------------------------------------------------------
`
}

type CmdHttp interface {
        handlerFunc(handler http.Handler) error
        command(cmd *cobra.Command, args []string) error
        GetCmd() *cobra.Command
        GRPCHandler(handler *grpc.Server)
}

type cmdHttp struct {
        stop <-chan bool // stop chan signal for server

        Port         int
        ReadTimeout  int
        WriteTimeout int
        Filename     string
        Cmd          *cobra.Command
        handler      http.Handler
        grpcHandler  *grpc.Server
        srv          *Server
}

// ServerBaseContext wraps an http.Handler to set the request context to the
// `baseCtx`.
func (c *cmdHttp) serverRoute() http.Handler {
        fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                if r.ProtoMajor == 2 && strings.HasPrefix(
                        r.Header.Get("Content-Type"), "application/grpc") &&
                        c.grpcHandler != nil {
                        c.grpcHandler.ServeHTTP(w, r)
                }
                c.handler.ServeHTTP(w, r)
        })
        return fn
}

func (c *cmdHttp) GRPCHandler(handler *grpc.Server) {
        c.grpcHandler = handler
}

func (c *cmdHttp) handlerFunc(handler http.Handler) error {
        ctx := context.Background()
        go func() {
                defer c.srv.Stop()
                <-ctx.Done()
                Info("I have to go...")
                Info("Stopping server gracefully")
        }()
        addrURL := url.URL{Scheme: "http", Host: fmt.Sprintf(":%v", c.Port)}
        Info(fmt.Sprintf("started server %s", addrURL.String()))
        c.srv = StartWebServer(
                addrURL,
                c.ReadTimeout,
                c.WriteTimeout,
                chi.ServerBaseContext(ctx, handler),
        )
        sc := make(chan os.Signal, 10)
        signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
        select {
        case s := <-sc:
                Info(fmt.Sprintf("shutting down server with signal %q", s.String()))
        case <-c.stop:
                Info("shutting down server with stop channel")
        case <-c.srv.StopNotify():
                Info("shutting down server with stop signal")
        }
        return nil
}

func (c *cmdHttp) command(cmd *cobra.Command, args []string) error {
        if c.handler == nil {
                Panic("handler function is nil")
        }

        // Description µ micro service
        fmt.Println(
                fmt.Sprintf(
                        Velkommen(),
                        c.Port,
                ))
        return c.handlerFunc(c.serverRoute())
}

func (c *cmdHttp) GetCmd() *cobra.Command {
        return c.Cmd
}

func NewCmdHttp(handler http.Handler, port, readTimeout, writeTimeout int) CmdHttp {
        return NewCmdHttpSignaled(handler, port, readTimeout, writeTimeout, nil)
}

func NewCmdHttpSignaled(
        handler http.Handler,
        port,
        readTimeout,
        writeTimeout int,
        stop <-chan bool,
) CmdHttp {
        c := &cmdHttp{
                stop:         stop,
                Port:         port,
                ReadTimeout:  readTimeout,
                WriteTimeout: writeTimeout,
                handler:      handler,
        }
        c.Cmd = &cobra.Command{
                Use:   "http",
                Short: "Used to run the http service",
                RunE:  c.command,
        }
        return c
}

// Server warps http.Server.
type Server struct {
        mu         sync.RWMutex
        addrURL    url.URL
        httpServer *http.Server

        stopc chan struct{}
        donec chan struct{}
}

// StopNotify returns receive-only stop channel to notify the server has stopped.
func (srv *Server) StopNotify() <-chan struct{} {
        return srv.stopc
}

// Stop stops the server. Useful for testing.
func (srv *Server) Stop() {
        Warn(fmt.Sprintf("stopping server %s", srv.addrURL.String()))
        srv.mu.Lock()
        if srv.httpServer == nil {
                srv.mu.Unlock()
                return
        }
        graceTimeOut := time.Duration(50)
        ctx, cancel := context.WithTimeout(context.Background(), graceTimeOut)
        defer cancel()
        if err := srv.httpServer.Shutdown(ctx); err != nil {
                Debug("Wait is over due to error")
                if err := srv.httpServer.Close(); err != nil {
                        Debug(err.Error())
                }
                Debug(err.Error())
        }
        close(srv.stopc)
        <-srv.donec
        srv.mu.Unlock()
        Warn(fmt.Sprintf("stopped server %s", srv.addrURL.String()))
}

// StartWebServer starts a web server
func StartWebServer(addr url.URL, readTimeout, writeTimeout int, handler http.Handler) *Server {
        stopc := make(chan struct{})
        srv := &Server{
                addrURL: addr,
                httpServer: &http.Server{
                        Addr:         addr.Host,
                        Handler:      handler,
                        ReadTimeout:  time.Duration(readTimeout) * time.Second,
                        WriteTimeout: time.Duration(writeTimeout) * time.Second,
                },
                stopc: stopc,
                donec: make(chan struct{}),
        }
        listener, err := net.Listen("tcp", addr.Host)
        if err != nil {
                Error(err.Error())
        }
        go func() {
                defer func() {
                        if err := recover(); err != nil {
                                Warn(
                                        "shutting down server with err ",
                                        Field("error", fmt.Sprintf(`(%v)`, err)),
                                )
                                os.Exit(0)
                        }
                        close(srv.donec)
                }()
                if err := srv.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
                        Fatal(
                                "shutting down server with err ",
                                Field("error", err),
                        )
                }
        }()
        return srv
}

const (
        StatusSuccess               = http.StatusOK
        StatusErrorForm             = http.StatusBadRequest
        StatusErrorUnknown          = http.StatusBadGateway
        StatusInternalError         = http.StatusInternalServerError
        StatusUnauthorized          = http.StatusUnauthorized
        StatusCreated               = http.StatusCreated
        StatusAccepted              = http.StatusAccepted
        StatusForbidden             = http.StatusForbidden
        StatusInvalidAuthentication = http.StatusProxyAuthRequired
)

var statusMap = map[int][]string{
        StatusSuccess:               {"STATUS_OK", "Success"},
        StatusErrorForm:             {"STATUS_BAD_REQUEST", "Invalid data request"},
        StatusErrorUnknown:          {"STATUS_BAG_GATEWAY", "Oops something went wrong"},
        StatusInternalError:         {"INTERNAL_SERVER_ERROR", "Oops something went wrong"},
        StatusUnauthorized:          {"STATUS_UNAUTHORIZED", "Not authorized to access the service"},
        StatusCreated:               {"STATUS_CREATED", "Resource has been created"},
        StatusAccepted:              {"STATUS_ACCEPTED", "Resource has been accepted"},
        StatusForbidden:             {"STATUS_FORBIDDEN", "Forbidden access the resource "},
        StatusInvalidAuthentication: {"STATUS_INVALID_AUTHENTICATION", "The resource owner or authorization server denied the request"},
}

func StatusCode(code int) string {
        return statusMap[code][0]
}

func StatusText(code int) string {
        return statusMap[code][1]
}
