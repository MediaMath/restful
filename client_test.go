package restful

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cenk/backoff"
)

var test = &foo{98}

func TestGetResponse(t *testing.T) {
	count := int32(0)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		accept := r.Header.Get("Accept")
		if accept != "accept1, accept2" {
			http.Error(w, fmt.Sprintf("accept: %v", accept), http.StatusBadRequest)
		}

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
	err := testClient(t, ts.URL).DoJSON(NewGetRequest("unit-test", "", "accept1", "accept2"), test, fooResponse)
	if err != nil {
		t.Fatal(err)
	}

	if fooResponse.Foo != 18 {
		t.Fatalf("ERR: %v", fooResponse)
	}

	if count > 1 {
		t.Fatal(count)
	}
}

func TestDontBackoff400Responses(t *testing.T) {
	count := int32(0)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		if atomic.LoadInt32(&count) > 1 {
			t.Fatal("Retried too much")
		}

		http.Error(w, "foo", http.StatusTeapot)
	}))
	defer ts.Close()

	err := testClient(t, ts.URL).DoJSON(NewGetRequest("unittest", ""), test, &foo{})
	if err == nil {
		t.Fatal("No error")
	}

	c := atomic.LoadInt32(&count)
	if c != 1 {
		t.Fatalf("wrong backoff: %v", c)
	}
}

func TestBackoffNon400Responses(t *testing.T) {
	count := int32(0)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		if atomic.LoadInt32(&count) > 4 {
			t.Fatal("Retried too much")
		}

		http.Error(w, "foo", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := testClient(t, ts.URL)

	notified := int32(0)
	client.BackOffNotify = func(e error, t time.Duration) { atomic.AddInt32(&notified, 1) }

	err := client.DoJSON(NewGetRequest("unittest", ""), test, &foo{})
	if err == nil {
		t.Errorf("Didn't error")
	}

	c := atomic.LoadInt32(&count)
	if c != 4 {
		t.Errorf("wrong backoff: %v", c)
	}

	n := atomic.LoadInt32(&notified)
	if n != 3 {
		t.Errorf("wrong notify: %v", n)
	}
}

func TestExpectedResponseCodeUnparseableBody(t *testing.T) {
	count := int32(0)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	err := testClient(t, ts.URL).DoJSON(NewPostRequest("unittest", "", ""), test, &foo{})

	if err == nil {
		t.Fatal("no error")
	}

	c := atomic.LoadInt32(&count)
	if c > 1 {
		t.Fatalf("don't backoff unparseable: %v", c)
	}
}

func TestRequestError(t *testing.T) {
	_, err := DefaultClient("h%ttp%")
	if err == nil {
		t.Fatal("No error")
	}
}

func newCountBackOff(max int) backoff.BackOff {
	return &countBackoff{max: max}
}

type countBackoff struct {
	count int
	max   int
}

func (b *countBackoff) Reset() { b.count = 0 }
func (b *countBackoff) NextBackOff() time.Duration {
	if b.count == b.max-1 {
		return backoff.Stop
	}

	b.count++
	return time.Duration(1) * time.Millisecond
}

func testClient(t *testing.T, url string) *Client {
	client, err := DefaultClient(url)
	if err != nil {
		t.Fatal(err)
	}

	client.CreateBackOff = func() backoff.BackOff { return newCountBackOff(4) }
	return client
}

type foo struct {
	Foo int `json:"foo"`
}
