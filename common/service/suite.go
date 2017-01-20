package service

import (
	gc "gopkg.in/check.v1"
	corecharm "gopkg.in/juju/charmrepo.v2-unstable"

	"github.com/juju/juju/agent"
	"github.com/juju/juju/juju/testing"
	"github.com/juju/juju/state"
	"github.com/juju/juju/testing/factory"

	jc "github.com/juju/testing/checkers"
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
	log.Infof("Initializing test suite")
	s.JujuConnSuite.SetUpTest(c)

	s.PatchValue(&corecharm.CacheDir, c.MkDir())

	s.service = NewFakeJujuService(s.BackingState, s.APIState, s.options)
	err := s.service.Initialize()
	c.Assert(err, jc.ErrorIsNil)

	controller := s.Factory.MakeMachine(c, &factory.MachineParams{
		InstanceId: s.service.NewInstanceId(),
		Nonce:      agent.BootstrapNonce,
		Jobs:       []state.MachineJob{state.JobManageModel, state.JobHostUnits},
		Series:     s.options.Series,
	})
	err = s.service.InitializeController(controller)
	c.Assert(err, gc.IsNil)
}
