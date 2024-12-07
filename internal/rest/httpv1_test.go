package rest_test

import (
	"io"
)

type testCase struct {
	name         string
	method       string
	path         string
	prepareMocks func()
	body         io.Reader
	expectedCode int
	expectedBody string
}
