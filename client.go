package restful

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

//Restful takes an http request and an optional response json struct
type Restful interface {
	DoJSON(request *http.Request, response interface{}) (status int, body []byte, err error)
}

//New creates a basic Restful with the provided http client
func New(client *http.Client) Restful {
	return &base{client}
}

//WithExpectedResult will error if the response status is not the provided one
func WithExpectedResult(r Restful, expected int) Restful {
	return &expectedResponse{r, expected}
}

//WithStats collects stats while it does the JSON
func WithStats(r Restful, stats Stats, requestName string) Restful {
	return &statsCollected{r, stats, requestName}
}

//WithBackoff will use the provided backoff policy
func WithBackoff(r Restful, b CreateBackOff, n Notify) Restful {
	return &backoff{r, b, n}
}

type base struct {
	*http.Client
}

func (r *base) DoJSON(request *http.Request, response interface{}) (status int, body []byte, err error) {
	var res *http.Response

	res, err = r.Client.Do(request)
	if err == nil {
		body, err = ioutil.ReadAll(res.Body)
		res.Body.Close()
		status = res.StatusCode
	}

	if err == nil && response != nil {
		err = json.Unmarshal(body, response)
	}

	return
}

type expectedResponse struct {
	Restful
	expected int
}

func (r *expectedResponse) DoJSON(request *http.Request, response interface{}) (status int, body []byte, err error) {
	status, body, err = r.Restful.DoJSON(request, response)

	if status != r.expected {
		err = &UnexpectedResponseError{
			Expected: r.expected,
			Received: status,
			Body:     body,
		}
	}

	return
}

//Stats is an interface for reporting statistics
type Stats interface {
	Incr(string)
	TimingPeriod(string, time.Time, time.Time)
}

type statsCollected struct {
	Restful
	stats       Stats
	requestName string
}

func (r *statsCollected) DoJSON(request *http.Request, response interface{}) (status int, body []byte, err error) {
	start := time.Now()
	status, body, err = r.Restful.DoJSON(request, response)
	end := time.Now()

	r.stats.Incr(r.statName("request"))

	if err != nil {
		r.stats.Incr(r.statName("request_error"))
	}

	if status > 0 {
		r.stats.TimingPeriod(r.statName("get_time"), start, end)
		r.stats.Incr(r.statName(fmt.Sprintf("response.%v", status)))
	}

	return
}

func (r *statsCollected) statName(name string) string {
	return fmt.Sprintf("%s.%s", r.requestName, name)
}

type backoff struct {
	Restful
	CreateBackOff CreateBackOff
	BackOffNotify Notify
}

func (r *backoff) DoJSON(request *http.Request, response interface{}) (status int, body []byte, err error) {
	b := r.CreateBackOff()
	b.Reset()
	for {
		status, body, err = r.Restful.DoJSON(request, response)

		if err == nil {
			return
		}

		//only backoff on server errors
		if !isServerError(status) {
			return
		}

		stop, next := b.Stop()
		if stop {
			return
		}

		if r.BackOffNotify != nil {
			r.BackOffNotify(err, next)
		}

		time.Sleep(next)
	}
}

func isServerError(status int) bool {
	return status >= http.StatusInternalServerError
}
