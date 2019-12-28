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

func (r *response) Success(success Meta) {
        r.Meta = success
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

// APIStatusSuccess for standard request api status success
func (r *response) APIStatusSuccess() {
        r.Success(Meta{
                Code:    StatusCode(StatusSuccess),
                Type:    StatusCode(StatusSuccess),
                Message: StatusText(StatusSuccess),
        })
}

// APIStatusCreated
func (r *response) APIStatusCreated() {
        r.Success(Meta{
                Code:    StatusCode(StatusCreated),
                Type:    StatusCode(StatusCreated),
                Message: StatusText(StatusCreated),
        })
}

// APIStatusAccepted
func (r *response) APIStatusAccepted() {
        r.Success(Meta{
                Code:    StatusCode(StatusAccepted),
                Type:    StatusCode(StatusAccepted),
                Message: StatusText(StatusAccepted),
        })
}

// APIStatusErrorUnknown
func (r *response) APIStatusErrorUnknown(err error) {
        r.Errors(Meta{
                Code:    StatusCode(StatusErrorUnknown),
                Type:    StatusCode(StatusErrorUnknown),
                Message: err.Error(),
        })
}

// APIStatusInvalidAuthentication
func (r *response) APIStatusInvalidAuthentication(err error) {
        r.Errors(Meta{
                Code:    StatusCode(StatusInvalidAuthentication),
                Type:    StatusCode(StatusInvalidAuthentication),
                Message: err.Error(),
        })
}

// APIStatusUnauthorized
func (r *response) APIStatusUnauthorized(err error) {
        r.Errors(Meta{
                Code:    StatusCode(StatusUnauthorized),
                Type:    StatusCode(StatusUnauthorized),
                Message: err.Error(),
        })
}

// APIStatusForbidden
func (r *response) APIStatusForbidden(err error) {
        r.Errors(Meta{
                Code:    StatusCode(StatusForbidden),
                Type:    StatusCode(StatusForbidden),
                Message: err.Error(),
        })
}

// APIStatusBadRequest
func (r *response) APIStatusBadRequest(err error) {
        r.Errors(Meta{
                Code:    StatusCode(StatusErrorForm),
                Type:    StatusCode(StatusErrorForm),
                Message: err.Error(),
        })
}
