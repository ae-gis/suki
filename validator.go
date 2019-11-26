/*  validator.go
*
* @Author:             Nanang Suryadi
* @Date:               November 22, 2019
* @Last Modified by:   @suryakencana007
* @Last Modified time: 22/11/19 14:02
 */

package suki

import (
        "fmt"
        "reflect"
        "strings"
        "time"

        "github.com/go-playground/validator/v10"
)

func Validate(s interface{}) (errors []Meta) {
        validate := validator.New()
        _ = validate.RegisterValidation("date", DateValidation)
        _ = validate.RegisterValidation("datetime", DatetimeValidation)
        _ = validate.RegisterValidation("daterange", DateRangeValidation)
        validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
                name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
                if name == "-" {
                        return ""
                }
                return name
        })

        if err := validate.Struct(s); err != nil {
                for _, err := range err.(validator.ValidationErrors) {
                        errors = append(errors, Meta{
                                Code:    err.Field(),
                                Type:    err.Type().String(),
                                Message: fmt.Sprintf("Invalid Type %v for input %s", err.Value(), err.Field()),
                        })
                }
                return errors
        }
        return nil
}


func DateValidation(fl validator.FieldLevel) bool {
        if _, err := time.Parse("2006-01-02", fl.Field().String()); err != nil {
                return false
        }
        return true
}

func DatetimeValidation(fl validator.FieldLevel) bool {
        if _, err := time.Parse(time.RFC3339, fl.Field().String()); err != nil {
                return false
        }
        return true
}

func DateRangeValidation(fl validator.FieldLevel) bool {

        var date = fl.Field().String()
        var minDate = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
        var maxDate = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)

        datetime, err := time.Parse("2006-01-02", date)
        if err != nil {
                return false
        }
        if datetime.Before(minDate) || datetime.After(maxDate) {
                return false
        }
        return true
}

func ParseDate(dtStr string) time.Time {
        date, err := time.Parse("2006-01-02", dtStr)
        if err != nil {
                return time.Time{}
        }
        return date
}

func ParseDatetime(dtStr string) time.Time {
        dateTime, err := time.Parse(time.RFC3339, dtStr)
        if err != nil {
                return time.Time{}
        }
        return dateTime
}
