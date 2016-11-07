package retry

import "testing"

func TestFib(t *testing.T) {
	for i, want := range []int64{0, 1, 1, 2, 3, 5, 8, 13, 21} {
		if got := fib(uint(i)); want != got {
			t.Fatalf("fib: for index %v, want: %v, got: %v", i, want, got)
		}
	}
}
