// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the README file.
// Source code and contact info at http://github.com/streadway/handy

package encoding

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type plain string

func (h plain) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(h))
}

type json string

func (h json) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json;charset=utf8")
	w.Write([]byte(h))
}

func decode(t *testing.T, body io.Reader) string {
	plain := &bytes.Buffer{}
	gz, err := gzip.NewReader(body)
	if err != nil {
		t.Fatalf("expected a gzip stream, got: %q", err)
	}
	io.Copy(plain, gz)
	return plain.String()
}

func acceptGzip() *http.Request {
	return &http.Request{
		Header: http.Header{"Accept-Encoding": {"gzip"}},
	}
}

func TestGzip(t *testing.T) {
	const msg = "the meaning of life, the universe and everything"

	var (
		handler = Gzip(plain(msg))
		resp    = httptest.NewRecorder()
		req     = acceptGzip()
	)

	handler.ServeHTTP(resp, req)

	if hdr := resp.HeaderMap.Get("Content-Encoding"); hdr != "gzip" {
		t.Fatalf("expected content encoding to be gzip, got: %q", hdr)
	}

	if hdr := resp.HeaderMap.Get("Vary"); hdr != "Accept-Encoding" {
		t.Fatalf("expected to vary on accept encoding, got: %q", hdr)
	}

	if want, got := msg, decode(t, resp.Body); want != got {
		t.Fatalf("expected to decompress message, got: %q", got)
	}
}

func TestMatchingGzipTypes(t *testing.T) {
	const msg = `{"meaning": 42}`

	var (
		types   = []string{"application/json"}
		handler = GzipTypes(types, json(msg))
		resp    = httptest.NewRecorder()
		req     = acceptGzip()
	)

	handler.ServeHTTP(resp, req)

	if want, got := "gzip", resp.HeaderMap.Get("Content-Encoding"); want != got {
		t.Fatalf("expected content encoding %q, got: %q", want, got)
	}

	if want, got := msg, decode(t, resp.Body); want != got {
		t.Fatalf("expected decoded json stream %q, got: %q", want, got)
	}
}

func TestNonMatchingGzipTypes(t *testing.T) {
	const msg = `just some plain text`

	var (
		types   = []string{"application/json"}
		handler = GzipTypes(types, plain(msg))
		resp    = httptest.NewRecorder()
		req     = acceptGzip()
	)

	handler.ServeHTTP(resp, req)

	if want, got := "", resp.HeaderMap.Get("Content-Encoding"); want != got {
		t.Fatalf("expected no content encoding, got: %q", got)
	}
}

func TestGzipper_InvalidLevel(t *testing.T) {
	const msg = `Fear is the mind killer`
	Gzipper(42)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError) // test error propagation

			n, err := w.Write([]byte(msg))
			if want, got := n, 0; want != got {
				t.Errorf("Write(%q) => want %d written bytes, got %d", msg, want, got)
			}

			for _, hdr := range [...]string{"Content-Encoding", "Vary"} {
				if got := w.Header().Get(hdr); got != "" {
					t.Errorf("Write(%q) => want no %s header, got %q", msg, hdr, got)
				}
			}

			want, got := errors.New("gzip: invalid compression level: 42"), err
			if !reflect.DeepEqual(want, got) {
				t.Errorf("Write(%q) => want err %q, got %q", msg, want, got)
			}
		}),
	).ServeHTTP(httptest.NewRecorder(), acceptGzip())
}

func TestGzipper_WriteHeader(t *testing.T) {
	const msg = `Fear is the little-death that brings total obliteration.`
	srv := httptest.NewServer(Gzipper(gzip.DefaultCompression)(plain(msg)))
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL, nil)
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	hdrs := map[string]string{"Content-Encoding": "gzip", "Vary": "Accept-Encoding"}
	for hdr, want := range hdrs {
		if got := resp.Header.Get(hdr); got != want {
			t.Errorf("want %s to be %q, got %q", hdr, want, got)
		}
	}
}
