package restful

import "fmt"

//UnexpectedResponseError indicates that the request happened correctly but the response was unexpected
type UnexpectedResponseError struct {
	Expected int
	Received int
}

func (u *UnexpectedResponseError) Error() string {
	return fmt.Sprintf("%v:%v", u.Expected, u.Received)
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
