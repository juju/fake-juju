package service

import (
	"io"

	gc "gopkg.in/check.v1"
	corecharm "gopkg.in/juju/charmrepo.v2-unstable"

	"github.com/juju/juju/agent"
	"github.com/juju/juju/juju/testing"
	"github.com/juju/juju/state"
	"github.com/juju/juju/testing/factory"
	"github.com/juju/loggo"
)

// Runtime options for the fake-juju service
type FakeJujuOptions struct {
	output io.Writer   // Where to write logs
	level  loggo.Level // The log level
	series string      // Default Ubuntu series
	mongo  bool        // Whether to start mongo, it's off for unit tests
}

// Core fake-juju service.
//
// It's implemented as a gocheck test suite because that's the easiest way
// to re-use all the code that sets up the dummy provider. Ideally such
// logic should be factored out from testing-related tooling and be made
// standalone.
type FakeJujuService struct {
	testing.JujuConnSuite

	options *FakeJujuOptions
	state   *FakeJujuState
}

func (s *FakeJujuService) SetUpTest(c *gc.C) {
	s.JujuConnSuite.SetUpTest(c)

	// Note that LoggingSuite.SetUpTest (github.com/juju/testing/log.go),
	// called via s.JujuConnSuite.SetUpTest(), calls loggo.ResetLogging().
	// So we cannot set up logging before then, since any writer we
	// register will get removed.  Consequently we lose any logs that get
	// generated in the SetUpTest() call.
	setupLogging(s.options.output, s.options.level)

	s.PatchValue(&corecharm.CacheDir, c.MkDir())

	s.state = &FakeJujuState{
		state:   s.State,
		api:     s.APIState,
		options: s.options,
	}

	err := s.state.Initialize()
	c.Assert(err, gc.IsNil)

	controller := s.Factory.MakeMachine(c, &factory.MachineParams{
		InstanceId: s.state.NewInstanceId(),
		Nonce:      agent.BootstrapNonce,
		Jobs:       []state.MachineJob{state.JobManageModel, state.JobHostUnits},
		Series:     s.options.series,
	})
	err = s.state.InitializeController(controller)
	c.Assert(err, gc.IsNil)
}

func (s *FakeJujuService) TestStart(c *gc.C) {
}
