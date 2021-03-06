# [restful](https://github.com/MediaMath/restful) &middot; [![CircleCI Status](https://circleci.com/gh/MediaMath/restful.svg?style=shield)](https://circleci.com/gh/MediaMath/restful) [![GitHub license](https://img.shields.io/badge/license-BSD3-blue.svg)](https://github.com/MediaMath/restful/blob/master/LICENSE) [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/MediaMath/restful/blob/master/CONTRIBUTING.md)

## restful - a very simple wrapper around restful clients in golang

Provides basic backoff and stats reporting facilities.

```go

import "github.com/MediaMath/restful

func main() {

    type foo struct {
        Foo int `json:"foo"`
    }

    client := restful.WithExpectedResult(http.DefaultClient, http.StatusOK)

    fooResponse := &foo{}
    status, body, err := restful.DoJSON(client, http.NewRequest("GET", "http://example.com", nil), fooResponse)

    if err != nil || fooResponse.Foo != 98 {
	log.Fatal("Incorrect", status, body, err)
    }
}
```
