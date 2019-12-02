/*  raw.go
*
* @Author:             Nanang Suryadi
* @Date:               November 24, 2019
* @Last Modified by:   @suryakencana007
* @Last Modified time: 24/11/19 22:00
 */

package sqlx

import (
        "bytes"
        "fmt"
        "reflect"
        "sort"
        "strconv"
        "strings"
        "time"
        "unicode"

        "gitlab.com/suryakencana007/suki"
)

type formatField func(key string, n int, opts tagOptions) string

type Model interface {
        TableName() string
}

type Query struct {
        raw   *SQL
        Query string
        Args  []interface{}
}

func Build() *Query {
        return &Query{}
}

func (s *Query) NewScope(model interface{}) *SQL {
        if s.raw == nil {
                s.raw = &SQL{query: s.clone(), Model: model, TagName: "sql"}
        } else {
                s.raw.Model = model
        }
        return s.raw
}

func (s *Query) SetTag(tag string) *Query {
        s.clone().raw.TagName = tag
        return s
}

// inner join ref_user refUser on refUser.id = refGame.user_id
func (s *Query) Join(model interface{}, fields ...string) *Query {
        return s.clone().raw.join(model, fields...).query
}

func (s *Query) Select(model interface{}) *Query {
        return s.NewScope(model).selectQuery().query
}

func (s *Query) Insert(model interface{}) *Query {
        return s.NewScope(model).insert().query
}

func (s *Query) Inserts(model interface{}) *Query {
        return s.NewScope(model).inserts().query
}

func (s *Query) Updates(model interface{}) *Query {
        return s.NewScope(model).updates().query
}

func (s *Query) Where(query interface{}, args ...interface{}) *Query {
        return s.clone().raw.Where(query, args...).query
}

func (s *Query) ToSQL() (string, []interface{}) {
        s.raw.Exec()
        if strings.EqualFold(s.Query, "insert") || strings.EqualFold(s.Query, "insert") {
                s.Query = strings.Join([]string{s.Query, "RETURNING id"}, " ")
        }
        return s.Query, s.Args
}

func (s *Query) clone() *Query {
        if s.raw == nil {
                s.raw = &SQL{query: s, TagName: "sql"}
        }
        return s
}

type SQL struct {
        Model           interface{}
        TagName         string
        query           *Query
        whereConditions []map[string]interface{}
}

func (r *SQL) Where(query interface{}, values ...interface{}) *SQL {
        r.whereConditions = append(r.whereConditions, map[string]interface{}{"query": query, "args": values})
        return r
}

func (r *SQL) Exec() {
        for _, w := range r.whereConditions {
                numargs := len(r.query.Args)
                query := w["query"].(string)
                lenQuestion := strings.Count(query, "?")
                for i := 1; i <= lenQuestion; i++ {
                        query = strings.Replace(query, "?", fmt.Sprintf("$%d", numargs+i), 1)
                }

                r.query.Query = strings.Join(
                        []string{
                                r.query.Query,
                                strings.Replace("WHERE ? ", "?", query, -1),
                        }, " ")

                args := w["args"].([]interface{})
                r.query.Args = append(r.query.Args, args...)
        }
}

func (r *SQL) join(model interface{}, fields ...string) *SQL {
        if len(fields) != 2 && !strings.Contains(r.query.Query, "select") {
                panic("select syntax or join field not found")
        }
        tableLeft := model.(Model).TableName()
        left := strings.Replace(fmt.Sprintf("?.%s", fields[0]), "?", suki.ToCamel(tableLeft), -1)
        right := strings.Replace(fmt.Sprintf("?.%s", fields[1]), "?", suki.ToCamel(r.Model.(Model).TableName()), -1)
        r.query.Query = fmt.Sprintf("%sINNER JOIN %s ON %s = %s ", r.query.Query, r.modelAlias(tableLeft), left, right)
        return r
}

func (r *SQL) modelAlias(tableName string) string {
        return strings.Replace(fmt.Sprintf("%s ?", tableName), "?", suki.ToCamel(tableName), -1)
}

func (r *SQL) selectQuery() *SQL {
        columns, err := r.fieldsToArgs(
                r.Model,
                func(key string, n int, opts tagOptions) string {
                        return fmt.Sprintf("%s.%s", suki.ToCamel(r.Model.(Model).TableName()), key)
                },
        )
        if err != nil {
                panic(err.Error())
        }
        // columns = append(columns, "create_date", "write_date")
        q := strings.Replace("SELECT %s FROM ? ", "?", r.modelAlias(r.Model.(Model).TableName()), -1)
        r.query.Query = fmt.Sprintf(
                q,
                strings.Join(columns, ", "),
        )
        return r
}

func (r *SQL) insert() *SQL {
        columns, err := r.fieldsToArgs(
                r.Model,
                func(key string, n int, opts tagOptions) string {
                        return key
                },
        )
        if err != nil {
                panic(err.Error())
        }
        columns = append(columns, "create_date", "write_date")
        r.query.Args = append(r.query.Args, time.Now().UTC(), time.Now().UTC())
        params := make([]string, 0)
        for i := 1; i <= len(r.query.Args); i++ {
                params = append(params, fmt.Sprintf(`$%d`, i))
        }
        q := strings.Replace("INSERT INTO ? (%s) VALUES (%v) ", "?", r.Model.(Model).TableName(), -1)
        r.query.Query = fmt.Sprintf(
                q,
                strings.Join(columns, ", "),
                strings.Join(params, ", "),
        )
        return r
}

func (r *SQL) inserts() *SQL {
        t := reflect.TypeOf(r.Model)
        val := reflect.ValueOf(r.Model)
        rows := make([][]string, 0)
        if t.Kind() == reflect.Slice {
                for i := 0; i < val.Len(); i++ {
                        model := val.Index(i).Interface()
                        f, err := r.fieldsToArgs(model, func(key string, n int, opts tagOptions) string {
                                return fmt.Sprintf(`$%d`, n)
                        })
                        if err != nil {
                                panic(err.Error())
                        }
                        f = append(f, fmt.Sprintf(`$%d`, len(r.query.Args)+1), fmt.Sprintf(`$%d`, len(r.query.Args)+2))
                        r.query.Args = append(r.query.Args, time.Now().UTC(), time.Now().UTC())

                        rows = append(rows, f)
                }
        }
        cols, err := TagsToField(r.TagName, val.Index(0).Interface())
        if err != nil {
                panic(err.Error())
        }
        keys := make([]string, 0)
        for k := range cols {
                keys = append(keys, k)
        }
        sort.Strings(keys)
        keys = append(keys, "create_date", "write_date")
        buff := bytes.NewBuffer([]byte{})
        for n, columns := range rows {
                buff.WriteString("(")
                for m, column := range columns {
                        buff.WriteString(column)
                        if m < len(columns)-1 {
                                buff.WriteString(", ")
                        }
                }
                buff.WriteString(")")
                if n < len(rows)-1 {
                        buff.WriteString(", ")
                }
        }
        q := strings.Replace("INSERT INTO ? (%s) VALUES %v ", "?", val.Index(0).Interface().(Model).TableName(), -1)
        r.query.Query = fmt.Sprintf(
                q,
                strings.Join(keys, ", "),
                buff.String(),
        )
        return r
}

func (r *SQL) updates() *SQL {
        r.query.Args = append(r.query.Args, time.Now().UTC())
        columns, err := r.fieldsToArgs(
                r.Model,
                func(key string, n int, opts tagOptions) string {
                        return fmt.Sprintf(`%s = $%d`, key, n)
                },
        )
        if err != nil {
                panic(err.Error())
        }
        columns = append(columns, "write_date = $1")
        q := strings.Replace("UPDATE ? SET %v ", "?", r.Model.(Model).TableName(), -1)
        r.query.Query = fmt.Sprintf(
                q,
                strings.Join(columns, ", "),
        )
        return r
}

func (r *SQL) fieldsToArgs(model interface{}, fn formatField) ([]string, error) {
        fields, err := TagsToField(r.TagName, model)
        if err != nil {
                return nil, err
        }
        keys := make([]string, 0)
        for k := range fields {
                keys = append(keys, k)
        }
        sort.Strings(keys)
        f := make([]string, 0)
        for _, k := range keys {
                field := fn(k, len(r.query.Args)+1, fields[k][1].(tagOptions))
                if len(field) < 1 {
                        continue
                }
                f = append(f, field)

                if integer, err := strconv.Atoi(fields[k][0].(string)); err == nil {
                        r.query.Args = append(r.query.Args, integer)
                        continue
                }

                if b, err := strconv.ParseBool(fields[k][0].(string)); err == nil {
                        r.query.Args = append(r.query.Args, b)
                        continue
                }

                r.query.Args = append(r.query.Args, fields[k][0])
        }
        return f, nil
}

func TagsToField(tag string, value interface{}) (result map[string][]interface{}, err error) {
        fn := func() (err error) {
                defer func() {
                        if e := recover(); e != nil {
                                err = e.(error)
                        }
                }()
                _ = reflect.ValueOf(value).Elem()
                return nil
        }
        if fn() != nil {
                obj := reflect.New(reflect.TypeOf(value))
                obj.Elem().Set(reflect.ValueOf(value))
                value = obj.Interface()
        }
        result = make(map[string][]interface{})
        t := reflect.ValueOf(value).Elem()
        for i := 0; i < t.NumField(); i++ {
                val := t.Field(i)
                field := t.Type().Field(i)
                tagVal := field.Tag.Get(tag)
                if isEmptyValue(val) || len(tagVal) < 1 {
                        continue
                }
                name, opts := parseTag(tagVal)
                if !isValidTag(name) {
                        name = ""
                }
                switch val.Interface().(type) {
                case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
                        result[name] = append(result[name], fmt.Sprintf("%v", val.Interface()))
                case *bool:
                        result[name] = append(result[name], fmt.Sprintf("%v", val.Elem().Interface()))
                case NullString:
                        if !val.Interface().(NullString).Valid {
                                continue
                        }
                        result[name] = append(result[name], val.Interface().(NullString).String)
                case NullInt64:
                        if !val.Interface().(NullInt64).Valid {
                                continue
                        }
                        result[name] = append(result[name], fmt.Sprintf("%v", val.Interface().(NullInt64).Int64))
                case NullTime:
                        if !val.Interface().(NullTime).Valid {
                                continue
                        }
                        result[name] = append(result[name], fmt.Sprintf("%v", val.Interface().(NullTime).Time))
                case NullFloat64:
                        if !val.Interface().(NullFloat64).Valid {
                                continue
                        }
                        result[name] = append(result[name], fmt.Sprintf("%v", val.Interface().(NullFloat64).Float64))
                case NullBool:
                        if !val.Interface().(NullBool).Valid {
                                continue
                        }
                        result[name] = append(result[name], fmt.Sprintf("%v", val.Interface().(NullBool).Bool))
                case time.Time:
                        result[name] = append(result[name], fmt.Sprintf("%v", val.Interface().(time.Time).UTC().Format(time.RFC3339)))
                default:
                        result[name] = append(result[name], fmt.Sprintf("%v", val.Interface()))
                }
                result[name] = append(result[name], opts)
        }
        return result, nil
}

func isEmptyValue(v reflect.Value) bool {
        switch v.Kind() {
        case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
                return v.Len() == 0
        case reflect.Bool:
                return !v.Bool()
        case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
                return v.Int() == 0
        case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
                return v.Uint() == 0
        case reflect.Float32, reflect.Float64:
                return v.Float() == 0
        case reflect.Interface, reflect.Ptr:
                return v.IsNil()
        }
        return false
}

func isValidTag(s string) bool {
        if s == "" {
                return false
        }
        for _, c := range s {
                switch {
                case strings.ContainsRune("!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
                        // Backslash and quote chars are reserved, but
                        // otherwise any punctuation chars are allowed
                        // in a tag name.
                case !unicode.IsLetter(c) && !unicode.IsDigit(c):
                        return false
                }
        }
        return true
}
