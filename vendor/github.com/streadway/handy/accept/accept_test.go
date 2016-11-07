// Copyright (c) 2015, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the README file.
// Source code and contact info at http://github.com/streadway/handy

package accept

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware(t *testing.T) {
	tests := []struct {
		accept     string
		middleware func(http.Handler) http.Handler
		code       int
	}{
		{
			accept:     "",
			middleware: Middleware(),
			code:       http.StatusOK,
		},
		{
			accept:     "text/plain",
			middleware: Middleware("text/plain"),
			code:       http.StatusOK,
		},
		{
			accept:     "text/plain",
			middleware: Middleware("application/json"),
			code:       http.StatusNotAcceptable,
		},
		{
			accept:     "text/event-stream",
			middleware: EventStream,
			code:       http.StatusOK,
		},
		{
			accept:     "text/html",
			middleware: HTML,
			code:       http.StatusOK,
		},
		{
			accept:     "application/json",
			middleware: JSON,
			code:       http.StatusOK,
		},
		{
			accept:     "text/plain",
			middleware: Plain,
			code:       http.StatusOK,
		},
		{
			accept:     "application/xml",
			middleware: XML,
			code:       http.StatusOK,
		},
	}

	for _, tt := range tests {
		w, r := httptest.NewRecorder(), newRequest(tt.accept)
		tt.middleware(okHandler).ServeHTTP(w, r)

		if want, got := tt.code, w.Code; want != got {
			t.Fatalf("%s want status %d, got %d", tt.accept, want, got)
		}
	}
}

func TestAcceptable(t *testing.T) {
	tests := []struct {
		accept string
		types  []string
		want   bool
	}{
		{
			accept: "",
			want:   true,
		},
		{
			accept: "*/*",
			want:   true,
		},
		{
			accept: "text/plain",
			want:   false,
		},
		{
			accept: "text/plain",
			types:  []string{"text/plain"},
			want:   true,
		},
		{
			accept: "text/html, text/plain",
			types:  []string{"text/plain"},
			want:   true,
		},
		{
			accept: "text/html, text/plain",
			types:  []string{"text/*", "text/html"},
			want:   true,
		},
		{
			accept: "text/html, text/plain, */*",
			types:  []string{"application/json"},
			want:   true,
		},
	}

	for _, tt := range tests {
		if want, got := tt.want, acceptable(tt.accept, tt.types); want != got {
			if tt.want {
				t.Errorf("want accept '%s' to pass (%s)", tt.accept, tt.types)
			} else {
				t.Errorf("want accept '%s' to not pass (%s)", tt.accept, tt.types)
			}
		}
	}
}

var okHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {}

func newRequest(accept string) *http.Request {
	return &http.Request{Header: map[string][]string{
		"Accept": []string{accept},
	}}
}
