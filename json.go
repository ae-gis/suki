/*  json.go
*
* @Author:             Nanang Suryadi
* @Date:               November 21, 2019
* @Last Modified by:   @suryakencana007
* @Last Modified time: 21/11/19 22:20
 */

package suki

import (
        "bytes"
        "encoding/json"
        "net/http"
)

// Write writes the data to http response writer
func WriteJSON(w http.ResponseWriter, r *http.Request, v interface{}) {
        buf := &bytes.Buffer{}
        enc := json.NewEncoder(buf)
        enc.SetEscapeHTML(true)
        if err := enc.Encode(v); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        if status, ok := r.Context().Value(CtxResponse).(int); ok {
                w.WriteHeader(status)
        }
        _, err := w.Write(buf.Bytes())
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
}
