package restful

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

//DoJSON will parse the body of a request response as json
func DoJSON(restful Restful, request *http.Request, response interface{}) (status int, body []byte, err error) {
	var res *http.Response

	res, err = restful.Do(request)
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

//Restful takes an http request and an optional response json struct
type Restful interface {
	Do(request *http.Request) (response *http.Response, err error)
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

type expectedResponse struct {
	Restful
	expected int
}

func (r *expectedResponse) Do(request *http.Request) (response *http.Response, err error) {
	response, err = r.Restful.Do(request)

	if response != nil && response.StatusCode != r.expected {
		err = &UnexpectedResponseError{
			Expected: r.expected,
			Received: response.StatusCode,
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

func (r *statsCollected) Do(request *http.Request) (response *http.Response, err error) {
	start := time.Now()
	response, err = r.Restful.Do(request)
	end := time.Now()

	r.stats.Incr(r.statName("request"))

	if err != nil {
		r.stats.Incr(r.statName("request_error"))
	}

	if response != nil {
		r.stats.TimingPeriod(r.statName("get_time"), start, end)
		r.stats.Incr(r.statName(fmt.Sprintf("response.%v", response.StatusCode)))
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

func (r *backoff) Do(request *http.Request) (response *http.Response, err error) {
	b := r.CreateBackOff()
	b.Reset()
	for {
		response, err = r.Restful.Do(request)

		if err == nil {
			return
		}

		//only backoff on server errors
		if response != nil && !isServerError(response.StatusCode) {
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
