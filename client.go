package restful

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/cenk/backoff"
)

//DefaultClient will return a restful client that uses the Default http Client, does not backoff and doesn't report stats
func DefaultClient(URL string) (*Client, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:        http.DefaultClient,
		CreateBackOff: func() backoff.BackOff { return &backoff.StopBackOff{} },
		BackOffNotify: nil,
		BaseURL:       u,
		Stats:         NilStats,
	}, nil
}

//Client does backed off calls to a json producing http service
type Client struct {
	Client        *http.Client
	CreateBackOff func() backoff.BackOff
	BackOffNotify backoff.Notify
	BaseURL       *url.URL
	Stats         Stats
}

//DoJSON will attempt to do the provided request with the configured backoff.  The response will be unmarshalled into response argument
func (r *Client) DoJSON(request *Request, data interface{}, response interface{}) error {
	if r.Stats == nil {
		r.Stats = NilStats
	}

	statName := func(name string) string { return fmt.Sprintf("%s.%s", request.Name, name) }

	var err error
	var body []byte
	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}

	//copy URL before modifying it
	u, err := url.Parse(r.BaseURL.String())
	if err != nil {
		return err
	}

	u.Path = request.Path
	u.RawQuery = request.Query

	b := r.CreateBackOff()
	b.Reset()
	var next time.Duration
	for {
		var reader io.Reader
		if data != nil {
			reader = bytes.NewBuffer(body)
		}
		req, requestError := http.NewRequest(request.Method, u.String(), reader)

		if request.Accept != "" {
			req.Header.Add("Accept", request.Accept)
		}

		if request.ContentType != "" {
			req.Header.Add("Content-Type", request.ContentType)
		}

		//request error indicates a bad URL or something, don't bother with trying/retrying
		if requestError != nil {
			r.Stats.Incr(statName("request_format_error"))
			return requestError
		}

		start := time.Now()
		status, body, err := r.do(req)
		end := time.Now()

		if err == nil {
			r.Stats.TimingPeriod(statName("get_time"), start, end)
			r.Stats.Incr(statName(fmt.Sprintf("response.%v", status)))
			unexpected := validateExpectedResponse(request.ExpectedStatus, status, body)
			if unexpected == nil {
				//You are likely pointing at the wrong thing if you got the expected response
				//but couldn't parse it.  Go ahead and don't retry
				if response != nil {
					if e := json.Unmarshal(body, response); e != nil {
						r.Stats.Incr(statName("request_error"))
						return fmt.Errorf("Cannot unmarshal response body: %v: %s", e, body)
					}
				}

				r.Stats.Incr(statName("request"))
				return nil
			}

			//Don't retry 400s
			if unexpected.IsClientRequestError() {
				r.Stats.Incr(statName("request_error"))
				return unexpected
			}

			err = unexpected
		}

		if next = b.NextBackOff(); next == backoff.Stop {
			return err
		}

		if r.BackOffNotify != nil {
			r.BackOffNotify(err, next)
		}

		r.Stats.Incr(statName("backoff"))
		time.Sleep(next)
	}
}

func (r *Client) do(req *http.Request) (status int, body []byte, err error) {

	var res *http.Response
	res, err = r.Client.Do(req)
	if err == nil {
		body, err = ioutil.ReadAll(res.Body)
		res.Body.Close()

		status = res.StatusCode
	}

	return
}
