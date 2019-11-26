
## Suki

> Understanding the Japanese Word "Suki"
The common Japanese word suki, pronounced "suh-kee", means a liking of, or fondness for; it means you love something or have a taste for that thing

- Use a Log

```go
package main
import (
    "github.com/suryakencana007/suki"
)

func main(){
    
    suki.Error("Error",
        suki.Field("error", "error for log"),
    )
    suki.Debug("Debug",
        suki.Field("debug", "debug for log"),
        suki.Field("message", "message debug for log"),
    )
    suki.Info("Info",
        suki.Field("info", "info for log"),
        suki.Field("message", "message info for log"),
    )
    suki.Warn("Warn",
        suki.Field("warning", "warning for log"),
        suki.Field("message", "message warning for log"),
    )
        
    log := suki.With(suki.Field("Root", "Messaging"))
    log.Debug("Debug",
        suki.Field("debug", "debug for log"),
        suki.Field("message", "message debug for log"),
    )
    log.Info("Info",
        suki.Field("info", "info for log"),
        suki.Field("message", "message info for log"),
    )

}

``` 

- Use a Http Serve

```go
package main
import (
     "fmt"
     "html"
     "net/http"
     "github.com/suryakencana007/suki"
)

func main(){
        handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                   _, err := fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
                   if err != nil {
                           panic(err)
                   }
        })
        s := suki.NewCmdHttp(8009, 100, 10, handler).GetCmd()
        if err := s.Execute(); err != nil {
                panic(err)
        }
}

``` 

- Use a Breaker

```go
package main
import (
    "net/http"
    "github.com/hashicorp/go-cleanhttp"
    "github.com/suryakencana007/suki"
)

func main() {
    endpoint := `www.google.com`
    cb := suki.NewBreaker(
        "",
        100,
        10,
    )
    var res *http.Response
    err := cb.Execute(func() (err error) {
        client := cleanhttp.DefaultClient()
        req, _ := http.NewRequest(http.MethodGet,
            endpoint, nil)
        res, err = client.Do(req)
        return err
    })
    if err != nil {
        suki.Error("Error",
            suki.Field("error", err.Error()),
        )
        panic(err)
    }
}
```

