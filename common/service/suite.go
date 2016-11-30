package service

import (
	gc "gopkg.in/check.v1"
	corecharm "gopkg.in/juju/charmrepo.v2-unstable"

	"github.com/juju/juju/agent"
	"github.com/juju/juju/juju/testing"
	"github.com/juju/juju/state"
	"github.com/juju/juju/testing/factory"
)

// Wrapper to setup and run the core FakeJujuService.
//
// It's implemented as a gocheck test suite because that's the easiest way
// to re-use all the code that sets up the dummy provider. Ideally such
// logic should be factored out from testing-related tooling and be made
// standalone.
type FakeJujuSuite struct {
	testing.JujuConnSuite

	options *FakeJujuOptions
	service *FakeJujuService
}

func (s *FakeJujuSuite) SetUpTest(c *gc.C) {
	s.JujuConnSuite.SetUpTest(c)

	// Note that LoggingSuite.SetUpTest (github.com/juju/testing/log.go),
	// called via s.JujuConnSuite.SetUpTest(), calls loggo.ResetLogging().
	// So we cannot set up logging before then, since any writer we
	// register will get removed.  Consequently we lose any logs that get
	// generated in the SetUpTest() call.
	setupLogging(s.options.Output, s.options.Level)

	s.PatchValue(&corecharm.CacheDir, c.MkDir())

	s.service = NewFakeJujuService(s.State, s.APIState, s.options)
	err := s.service.Initialize()
	c.Assert(err, gc.IsNil)

	controller := s.Factory.MakeMachine(c, &factory.MachineParams{
		InstanceId: s.service.NewInstanceId(),
		Nonce:      agent.BootstrapNonce,
		Jobs:       []state.MachineJob{state.JobManageModel, state.JobHostUnits},
		Series:     s.options.Series,
	})
	err = s.service.InitializeController(controller)
	c.Assert(err, gc.IsNil)
}

func (s *FakeJujuSuite) TestStart(c *gc.C) {
}
