/*  response_test.go
*
* @Author:             Nanang Suryadi
* @Date:               November 21, 2019
* @Last Modified by:   @suryakencana007
* @Last Modified time: 21/11/19 22:38
 */

package suki

import (
        "encoding/json"
        "io/ioutil"
        "net/http"
        "net/http/httptest"
        "strings"
        "testing"

        "github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
        data := map[string]interface{}{
                "message": "transaksi telah sukses",
        }
        result := Response()
        result.Body(data)
        assert.Equal(t, result.Data, data)
}

func TestResponseErrors(t *testing.T) {
        errs := make([]Meta, 0)
        errs = append(errs, Meta{
                Code:    StatusCode(StatusInternalError),
                Type:    "1000",
                Message: "constraint unique key duplicate",
        })

        result := Response()
        result.Errors(errs...)
        assert.Equal(t, result.Meta, errs)
        assert.Equal(t, "INTERNAL_SERVER_ERROR", result.Meta.([]Meta)[0].Code)
        assert.Equal(t, "constraint unique key duplicate", result.Meta.([]Meta)[0].Message)
}

func TestResponseErrorsJSON(t *testing.T) {
        errs := make([]Meta, 0)
        errs = append(errs, Meta{
                Code:    StatusCode(StatusInternalError),
                Type:    "1000",
                Message: "constraint unique key duplicate",
        })
        result := Response()
        result.Errors(errs...)

        r, err := http.NewRequest(http.MethodGet, "/", nil)

        assert.NoError(t, err)
        w := httptest.NewRecorder()
        w.WriteHeader(StatusInternalError) // set header code

        WriteJSON(w, r, result) // Write http Body to JSON

        if got, want := w.Code, StatusInternalError; got != want {
                t.Fatalf("status code got: %d, want %d", got, want)
        }

        expected, err := json.Marshal(result)
        if err != nil {
                t.Fatal(err)
        }

        actual, err := ioutil.ReadAll(w.Body)
        if err != nil {
                t.Fatal(err)
        }

        assert.Equal(t, result.Meta, errs)
        assert.Equal(t, "INTERNAL_SERVER_ERROR", result.Meta.([]Meta)[0].Code)
        assert.Equal(t, "constraint unique key duplicate", result.Meta.([]Meta)[0].Message)
        assert.Equal(t, string(expected), strings.TrimSuffix(string(actual), "\n"))
}

func TestResponseCSV(t *testing.T) {
        rows := make([][]string, 0)
        rows = append(rows, []string{"SO Number", "Nama Warung", "Area", "Fleet Number", "Jarak Warehouse", "Urutan"})
        rows = append(rows, []string{"SO45678", "WPD00011", "Jakarta Selatan", "1", "45.00", "1"})
        rows = append(rows, []string{"SO45645", "WPD001123", "Jakarta Selatan", "1", "43.00", "2"})
        rows = append(rows, []string{"SO45645", "WPD003343", "Jakarta Selatan", "1", "43.00", "3"})

        r, err := http.NewRequest(http.MethodGet, "/csv", nil)

        assert.NoError(t, err)
        w := httptest.NewRecorder()
        w.WriteHeader(StatusSuccess) // set header code

        WriteCSV(w, r, rows, "result-route-fleets") // Write http Body to JSON

        if got, want := w.Code, StatusSuccess; got != want {
                t.Fatalf("status code got: %d, want %d", got, want)
        }

        actual, err := ioutil.ReadAll(w.Body)
        if err != nil {
                t.Fatal(err)
        }

        assert.Equal(t, `SO Number,Nama Warung,Area,Fleet Number,Jarak Warehouse,Urutan
SO45678,WPD00011,Jakarta Selatan,1,45.00,1
SO45645,WPD001123,Jakarta Selatan,1,43.00,2
SO45645,WPD003343,Jakarta Selatan,1,43.00,3
`, string(actual))
        assert.Contains(t, w.Header().Get("Content-Type"), "text/csv")

}
