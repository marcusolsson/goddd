package location

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) TestEquality(c *C) {
	c.Check(Stockholm.UNLocode, Equals, Stockholm.UNLocode)
	c.Check(Stockholm.UNLocode, Not(Equals), Hongkong.UNLocode)
}
