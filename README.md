## restful - a very simple wrapper around restful clients in golang

Provides basic backoff and stats reporting facilities.

```go

import "github.com/MediaMath/restful

func main() {

    type foo struct {
        Foo int `json:"foo"`
    }

    client := restful.New(http.DefaultClient)

    fooResponse := &foo{}
    status, body, err := client.DoJSON(http.NewRequest("GET", "http://example.com", nil), fooResponse)

    if err != nil || fooResponse.Foo != 98 {
	log.Fatal("Incorrect", status, body, err)
    }
}
```
