/*  recovery.go
*
* @Author:             Nanang Suryadi
* @Date:               November 26, 2019
* @Last Modified by:   @suryakencana007
* @Last Modified time: 26/11/19 19:08
 */

package ruuto

import (
        "net/http"

        "github.com/ae-gis/suki"
)

func Recovery() func(next http.Handler) http.Handler {
        return func(next http.Handler) http.Handler {
                return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                        defer func() {
                                if err := recover(); err != nil {
                                        suki.Status(r, suki.StatusInternalError)
                                        suki.With(
                                                suki.Field("method", r.Method),
                                                suki.Field("path", r.URL.Path),
                                        ).Error("Internal server error handled")
                                        switch internalErr := err.(type) {
                                        case error:
                                                suki.WriteJSON(w, r, internalErr.Error())
                                        default:
                                                suki.WriteJSON(w, r, err)
                                        }
                                }
                        }()
                        next.ServeHTTP(w, r)
                })
        }
}
