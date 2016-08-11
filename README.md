## restful - a very simple wrapper around restful clients in golang

Provides basic backoff and stats reporting facilities.

```go

import "github.com/MediaMath/restful

func main() {

    type foo struct {
        Foo int `json:"foo"`
    }

    client, _ := restful.DefaultClient("http://example.com")

    fooResponse := &foo{}
    //Post a get request to the base url/foos that with accept headers accept1 and accept2 and expects a json
    //response {"foo":98}
    err := client.DoJSON(restful.NewGetRequest("unit-test", "foos", "accept1", "accept2"), nil, fooResponse)

    if err != nil || fooResponse.Foo != 98 {
	log.Fatal("Incorrect")
    }
}
```
