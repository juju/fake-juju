package service_test

import (
	"bytes"

	gc "gopkg.in/check.v1"

	"github.com/juju/loggo"

	"../service"
)

type DummySuite struct {
	options *service.FakeJujuOptions
	hasRun  bool
}

func (s *DummySuite) TestStart(c *gc.C) {
	s.hasRun = true
}

type FakeJujuRunnerSuite struct {}

// The FakeJujuRunner.Run method executes the given gocheck suite. Logging is
// configured too.
func (s *FakeJujuRunnerSuite) TestRun(c *gc.C) {
	output := &bytes.Buffer{}
	options := &service.FakeJujuOptions{
		Output: output,
		Level:  loggo.DEBUG,
		Mongo: -1,  // We don't need any MongoDB for this test
	}
	suite := &DummySuite{
		options: options,
	}
	runner := service.NewFakeJujuRunner(gc.Suite(suite), options)
	result := runner.Run()

	c.Assert(result.Succeeded, gc.Equals, 1)
	c.Assert(result.RunError, gc.IsNil)
	c.Assert(suite.hasRun, gc.Equals, true)
	c.Assert(
		bytes.Contains(output.Bytes(), []byte("Starting service")), gc.Equals, true)
}

var _ = gc.Suite(&FakeJujuRunnerSuite{})
