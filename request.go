package restful

import "strings"

//NewGetRequest creates a Request representing a Get
func NewGetRequest(name string, path string, accept ...string) *Request {
	return &Request{
		Name:           name,
		Path:           path,
		Method:         "GET",
		ExpectedStatus: 200,
		Accept:         strings.Join(accept, ", "),
	}

}

//NewPostRequest creates a Request representing a Post
func NewPostRequest(name string, path string, contentType string) *Request {
	return &Request{
		Name:           name,
		Method:         "POST",
		Path:           path,
		ExpectedStatus: 200,
		ContentType:    contentType,
	}
}

//NewDeleteRequest creates a Request representing a Delete
func NewDeleteRequest(name string, path string, accept ...string) *Request {
	return &Request{
		Name:           name,
		Path:           path,
		Method:         "DELETE",
		ExpectedStatus: 204,
		Accept:         strings.Join(accept, ", "),
	}
}

//Request is the meta data about the request to send
type Request struct {
	Name           string
	Method         string
	Path           string
	ExpectedStatus int
	Accept         string
	ContentType    string
	Query          string
}
