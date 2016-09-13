package restful

//Copyright 2016 MediaMath <http://www.mediamath.com>.  All rights reserved.
//Use of this source code is governed by a BSD-style
//license that can be found in the LICENSE file.

import (
	"fmt"
	"log"
	"testing"
)

func TestIsUnexpectedOnNonUnexpected(t *testing.T) {
	e := IsUnexpectedResponseError(fmt.Errorf("Nope"))
	if e != nil {
		t.Fatal(e)
	}
}

func TestIsUnexpectedOnNil(t *testing.T) {
	e := IsUnexpectedResponseError(nil)
	if e != nil {
		t.Fatal(e)
	}
}

func TestIsUnexpectedOnUnexpected(t *testing.T) {
	e := IsUnexpectedResponseError(&UnexpectedResponseError{})
	if e == nil {
		t.Fatal("Wasn't error")
	}
}

func TestUnexpectedIsError(t *testing.T) {
	a := func(e error) { log.Printf(e.Error()) }
	a(&UnexpectedResponseError{})
}
