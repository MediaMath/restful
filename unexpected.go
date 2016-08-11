package restful

import (
	"fmt"
	"net/http"
)

//UnexpectedResponseError indicates that the request happened correctly but the response was unexpected
type UnexpectedResponseError struct {
	Expected int
	Received int
	Body     []byte
}

//IsClientRequestError is true if the received status is in the 400s
func (u *UnexpectedResponseError) IsClientRequestError() bool {
	return u.Received >= http.StatusBadRequest && u.Received < http.StatusInternalServerError
}

func (u *UnexpectedResponseError) Error() string {
	return fmt.Sprintf("%v:%v:%s", u.Expected, u.Received, u.Body)
}

func validateExpectedResponse(expected int, status int, body []byte) *UnexpectedResponseError {
	if expected != status {
		return &UnexpectedResponseError{
			Expected: expected,
			Received: status,
			Body:     body,
		}
	}

	return nil
}

//IsUnexpectedResponseError will return the error as a UnexpectedResponseError struct or nil
func IsUnexpectedResponseError(err error) *UnexpectedResponseError {
	if err == nil {
		return nil
	}

	e, is := err.(*UnexpectedResponseError)
	if is {
		return e
	}

	return nil
}
