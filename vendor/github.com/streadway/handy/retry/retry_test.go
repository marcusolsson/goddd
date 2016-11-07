package retry

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

type testRoundTrip struct {
	err   error
	resp  *http.Response
	count int
}

func (rt *testRoundTrip) RoundTrip(*http.Request) (*http.Response, error) {
	rt.count++
	return rt.resp, rt.err
}

// this actually tests the internal counter of the retry loop
func TestRetryAfterCount(t *testing.T) {
	const attempts = 2

	var (
		req, _ = http.NewRequest("GET", "http://example/test", nil)
		next   = &testRoundTrip{err: fmt.Errorf("next"), resp: nil}
		trans  = Transport{
			Retry: All(Errors(), Max(attempts)),
			Next:  next,
		}
	)

	resp, err := trans.RoundTrip(req)

	if have, got := next.err.Error(), err.Error(); have == got {
		t.Fatalf("expected to override error from next")
	}

	if want, got := attempts, next.count; want != got {
		t.Fatalf("expected to make %d attempts, got %d", want, got)
	}

	if resp != nil {
		t.Fatalf("expected response to be nil since error is not nil")
	}
}

// this also tests the internal counter of the retry loop
func TestNoRetry(t *testing.T) {
	const attempts = 2

	var (
		req, _ = http.NewRequest("GET", "http://example/test", nil)
		next   = &testRoundTrip{err: nil, resp: &http.Response{StatusCode: 200}}
		trans  = Transport{
			Retry: All(Errors(), Max(attempts)),
			Next:  next,
		}
	)

	resp, err := trans.RoundTrip(req)

	if err != nil {
		t.Fatalf("expected error to be nil but got: %s", err.Error())
	}

	if resp == nil {
		t.Fatalf("expected to obtain non-nil response")
	}
}

func TestRetryDelay(t *testing.T) {
	const attempts = 2

	var (
		req, _ = http.NewRequest("GET", "http://example/test", nil)
		next   = &testRoundTrip{err: fmt.Errorf("next")}
		trans  = Transport{
			Retry: All(Errors(), Max(attempts)),
			Next:  next,
			Delay: Constant(500 * time.Millisecond),
		}
	)

	_, err := trans.RoundTrip(req)

	if have, got := next.err.Error(), err.Error(); have == got {
		t.Fatalf("expected to override error from next")
	}

	if want, got := attempts, next.count; want != got {
		t.Fatalf("expected to make %d attempts, got %d", want, got)
	}
}
