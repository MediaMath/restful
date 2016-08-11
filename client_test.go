package restful

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var test = &foo{98}

func TestGetResponse(t *testing.T) {
	count := int32(0)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)

		resp := &foo{18}
		b, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}))
	defer ts.Close()

	fooResponse := &foo{}
	status, _, err := New(tstClient()).DoJSON(tstRequest(t, ts.URL), fooResponse)
	c := atomic.LoadInt32(&count)

	require.Nil(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, 18, fooResponse.Foo)
	assert.Equal(t, int32(1), c)
}

func TestDontBackoff400Responses(t *testing.T) {
	count := int32(0)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		http.Error(w, "foo", http.StatusTeapot)
	}))
	defer ts.Close()

	status, _, err := WithBackoff(New(tstClient()), newCountBackOff(500), nil).DoJSON(tstRequest(t, ts.URL), &foo{})
	c := atomic.LoadInt32(&count)

	assert.NotNil(t, err)
	assert.Equal(t, int32(1), c)
	assert.Equal(t, http.StatusTeapot, status)
}

func TestDontBackoff200ResponsesIfErrors(t *testing.T) {
	count := int32(0)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		//is a 200 but we cant parse it
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	status, _, err := WithBackoff(New(tstClient()), newCountBackOff(500), nil).DoJSON(tstRequest(t, ts.URL), &foo{})
	c := atomic.LoadInt32(&count)

	assert.NotNil(t, err)
	assert.Equal(t, int32(1), c)
	assert.Equal(t, http.StatusOK, status)
}

func TestBackoff500Responses(t *testing.T) {
	count := int32(0)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		if atomic.LoadInt32(&count) > 4 {
			t.Fatal("Retried too much")
		}

		http.Error(w, "foo", http.StatusInternalServerError)
	}))
	defer ts.Close()

	notified := int32(0)
	notify := func(e error, t time.Duration) { atomic.AddInt32(&notified, 1) }

	status, _, err := WithBackoff(New(tstClient()), newCountBackOff(4), notify).DoJSON(tstRequest(t, ts.URL), &foo{})
	c := atomic.LoadInt32(&count)
	n := atomic.LoadInt32(&notified)

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Equal(t, int32(4), c)
	assert.Equal(t, int32(3), n)
}

func newCountBackOff(max int) CreateBackOff {
	return func() BackOff { return &countBackoff{max: max} }
}

type countBackoff struct {
	count int
	max   int
}

func (b *countBackoff) Reset() { b.count = 0 }
func (b *countBackoff) Stop() (bool, time.Duration) {
	b.count++
	return b.count == b.max, time.Millisecond
}

func tstClient() *http.Client {
	return &http.Client{Timeout: time.Duration(time.Second)}
}

func tstRequest(t testing.TB, url string) *http.Request {
	req, requestError := http.NewRequest("GET", url, nil)
	if requestError != nil {
		t.Fatalf("Request error: %v", requestError)
	}

	return req
}

type foo struct {
	Foo int `json:"foo"`
}
