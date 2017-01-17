package service_test

import (
	"bytes"

	gc "gopkg.in/check.v1"

	"github.com/juju/loggo"

	"../service"
)

type FakeJujuRunnerSuite struct{}

// The FakeJujuRunner.Run method sets up logging and starts the service main
// loop, which can be terminated with the Stop method.
func (s *FakeJujuRunnerSuite) TestRun(c *gc.C) {
	output := &bytes.Buffer{}
	options := &service.FakeJujuOptions{
		Output: output,
		Level:  loggo.DEBUG,
		Mongo:  -1, // We don't need any MongoDB for this test
	}

	// We don't pass an actual FakeJujuSuite here, since we're only going
	// to start the main loop.
	runner := service.NewFakeJujuRunner(nil, options)
	runner.Run()
	runner.Stop()
	result := runner.Wait()

	c.Assert(result.String(), gc.Equals, "OK: 1 passed")
	c.Assert(result.Succeeded, gc.Equals, 1)
	c.Assert(result.RunError, gc.IsNil)
	c.Assert(
		bytes.Contains(output.Bytes(), []byte("Starting service")), gc.Equals, true)
}

var _ = gc.Suite(&FakeJujuRunnerSuite{})
