/*  response.go
*
* @Author:             Nanang Suryadi
* @Date:               November 21, 2019
* @Last Modified by:   @suryakencana007
* @Last Modified time: 21/11/19 22:02
 */

package suki

import (
        "context"
        "net/http"
)

type ctxKeyResponse struct {
        Name string
}

func (r *ctxKeyResponse) String() string {
        return "context value " + r.Name
}

var CtxResponse = ctxKeyResponse{Name: "context response"}

type Pagination struct {
        Page  int `json:"page"`
        Size  int `json:"size"`
        Total int `json:"total"`
}

type Meta struct {
        Code    string `json:"code,omitempty"`
        Type    string `json:"error_type,omitempty"`
        Message string `json:"error_message,omitempty"`
}

type response struct {
        Meta       interface{} `json:"meta,omitempty"`
        Data       interface{} `json:"data,omitempty"`
        Pagination interface{} `json:"pagination,omitempty"`
}

func Response() *response {
        null := make(map[string]interface{})
        return &response{
                Meta:       null,
                Data:       null,
                Pagination: null,
        }
}

func (r *response) Errors(err ...Meta) {
        r.Meta = err
}

func (r *response) Success(code string) {
        r.Meta = Meta{Code: code}
}

func (r *response) Body(body interface{}) {
        r.Data = body
}

func (r *response) Page(p Pagination) {
        r.Pagination = p
}

func Status(r *http.Request, status int) {
        *r = *r.WithContext(context.WithValue(r.Context(), CtxResponse, status))
}
