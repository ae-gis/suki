/*  type.go
*
* @Author:             Nanang Suryadi
* @Date:               November 24, 2019
* @Last Modified by:   @suryakencana007
* @Last Modified time: 24/11/19 22:02
 */

package sqlx

import (
        "database/sql"
        "database/sql/driver"
        "encoding/json"
        "fmt"
        "time"
)

// NullInt64 is an alias for sql.NullInt64 data type
type NullInt64 struct {
        sql.NullInt64
}

// MarshalJSON for NullInt64
func (ni NullInt64) MarshalJSON() ([]byte, error) {
        if !ni.Valid {
                return []byte("0"), nil
        }
        return json.Marshal(ni.Int64)
}

func (ni NullInt64) Int() int {
        if !ni.Valid {
                return int(ni.Int64)
        }
        return 0
}

func Int64(value int64) NullInt64 {
        return NullInt64{struct {
                Int64 int64
                Valid bool
        }{Int64: value, Valid: true}}
}

// NullBool is an alias for sql.NullBool data type
type NullBool struct {
        sql.NullBool
}

// MarshalJSON for NullBool
func (nb NullBool) MarshalJSON() ([]byte, error) {
        if !nb.Valid {
                return []byte("null"), nil
        }
        return json.Marshal(nb.Bool)
}

func Bool(value bool) NullBool {
        return NullBool{struct {
                Bool  bool
                Valid bool
        }{Bool: value, Valid: true}}
}

// NullFloat64 is an alias for sql.NullFloat64 data type
type NullFloat64 struct {
        sql.NullFloat64
}

// MarshalJSON for NullFloat64
func (nf NullFloat64) MarshalJSON() ([]byte, error) {
        if !nf.Valid {
                return []byte("0"), nil
        }
        return json.Marshal(nf.Float64)
}

func Float64(value float64) NullFloat64 {
        return NullFloat64{struct {
                Float64 float64
                Valid   bool
        }{Float64: value, Valid: true}}
}

// NullString is an alias for sql.NullString data type
type NullString struct {
        sql.NullString
}

// MarshalJSON for NullString
func (ns NullString) MarshalJSON() ([]byte, error) {
        if !ns.Valid {
                return []byte("null"), nil
        }
        return json.Marshal(ns.String)
}

func String(value string) NullString {
        return NullString{struct {
                String string
                Valid  bool
        }{String: value, Valid: true}}
}

type NullTime struct {
        Time  time.Time
        Valid bool // Valid is true if Time is not NULL
}

// Scan implements the Scanner interface.
func (nt NullTime) Scan(value interface{}) error {
        nt.Time, nt.Valid = value.(time.Time)
        return nil
}

// Value implements the driver Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
        if !nt.Valid {
                return nil, nil
        }
        return nt.Time, nil
}

// MarshalJSON for NullTime
func (nt NullTime) MarshalJSON() ([]byte, error) {
        if !nt.Valid {
                return []byte("null"), nil
        }
        val := fmt.Sprintf("\"%s\"", nt.Time.Format(time.RFC3339))
        return []byte(val), nil
}

func Time(value time.Time) NullTime {
        return NullTime{
                Time:  value,
                Valid: true,
        }
}
