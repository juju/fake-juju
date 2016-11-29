package service

import (
	"bytes"

	gc "gopkg.in/check.v1"

	"github.com/juju/loggo"
)

// Yes, a fake of the fake!
type DummyService struct {
	options *FakeJujuOptions
	hasRun  bool
}

func (s *DummyService) TestStart(c *gc.C) {
	s.hasRun = true
}

type FakeJujuRunnerSuite struct {
}

// The FakeJujuRunner.Run method executes the "service", by
// running the test suite it's implemented as. Logging is
// configured too.
func (s *FakeJujuRunnerSuite) TestRun(c *gc.C) {
	output := &bytes.Buffer{}
	options := &FakeJujuOptions{
		output: output,
		level:  loggo.DEBUG,
		mongo:  false,
	}
	service := &DummyService{
		options: options,
	}
	fake := &FakeJujuRunner{
		service: gc.Suite(service),
		options: options,
	}
	result := fake.Run()

	c.Assert(result.Succeeded, gc.Equals, 1)
	c.Assert(result.RunError, gc.IsNil)
	c.Assert(service.hasRun, gc.Equals, true)
	c.Assert(
		bytes.Contains(output.Bytes(), []byte("Starting service")), gc.Equals, true)
}

var _ = gc.Suite(&FakeJujuRunnerSuite{})
