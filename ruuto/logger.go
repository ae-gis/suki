/*  logger.go
*
* @Author:             Nanang Suryadi
* @Date:               November 26, 2019
* @Last Modified by:   @suryakencana007
* @Last Modified time: 26/11/19 18:40
 */

package ruuto

import (
        "net/http"
        "time"

        "github.com/felixge/httpsnoop"
        "gitlab.com/suryakencana007/suki"
)

func Logger() func(next http.Handler) http.Handler {
        return func(next http.Handler) http.Handler {
                return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                        suki.With(
                                suki.Field("method", r.Method),
                                suki.Field("path", r.URL.Path),
                        ).Debug("Started Request")
                        m := httpsnoop.CaptureMetrics(next, w, r)
                        suki.With(
                                suki.Field("code", m.Code),
                                suki.Field("duration", int(m.Duration/time.Millisecond)),
                                suki.Field("duration-fmt", m.Duration.String()),
                                suki.Field("method", r.Method),
                                suki.Field("host", r.Host),
                                suki.Field("request", r.RequestURI),
                                suki.Field("remote-addr", r.RemoteAddr),
                                suki.Field("referer", r.Referer()),
                                suki.Field("user-agent", r.UserAgent()),
                        ).Info("Completed handling request")
                })
        }
}
