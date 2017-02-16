package service_test

import (
	"bytes"
	"strings"

	gc "gopkg.in/check.v1"

	"github.com/juju/juju/api"
	"github.com/juju/loggo"
	"github.com/juju/testing"

	"../service"
)

type FakeJujuRunnerSuite struct {
	testing.MgoSuite
	output  *bytes.Buffer
	runner  *service.FakeJujuRunner
	options *service.FakeJujuOptions
}

func (s *FakeJujuRunnerSuite) SetUpTest(c *gc.C) {
	s.MgoSuite.SetUpTest(c)
	s.output = &bytes.Buffer{}
	s.options = &service.FakeJujuOptions{
		Output:            s.output,
		Series:            "xenial",
		Level:             loggo.DEBUG,
		Mongo:             testing.MgoServer.Port(),
		Port:              12345,
		UseRandomCert:     true,
		AutoStartMachines: true,
	}
	s.runner = service.NewFakeJujuRunner(s.options)
}

// The FakeJujuRunner.Run method sets up logging and starts the service main
// loop, which can be terminated with the Stop method.
func (s *FakeJujuRunnerSuite) TestRun(c *gc.C) {
	s.runner.Run()
	s.runner.Stop()
	result := s.runner.Wait()

	c.Assert(result.String(), gc.Equals, "OK: 1 passed")
	c.Assert(result.Succeeded, gc.Equals, 1)
	c.Assert(result.RunError, gc.IsNil)
	c.Assert(
		strings.Contains(s.output.String(), "Starting service"), gc.Equals, true)
}

// The "bootstrap" API endpoint setups up a new controller backed by the
// dummy provider.
func (s *FakeJujuRunnerSuite) TestBootstrapAPI(c *gc.C) {
	s.runner.Run()
	defer s.runner.Wait()
	defer s.runner.Stop()

	client := api.NewFakeJujuClientWithPort(12346)
	err := client.Bootstrap()
	if err != nil {
		c.Log(s.output.String())
		c.Error(err.Error())
	}
}

var _ = gc.Suite(&FakeJujuRunnerSuite{})
