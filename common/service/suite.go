package service

import (
	gc "gopkg.in/check.v1"
	corecharm "gopkg.in/juju/charmrepo.v2-unstable"

	"github.com/juju/juju/juju/testing"
	"github.com/juju/juju/state"

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
	log.Infof("Initializing fake-juju controller")
	s.JujuConnSuite.SetUpTest(c)

	s.PatchValue(&corecharm.CacheDir, c.MkDir())

	s.service = NewFakeJujuService(s.BackingState, s.APIState, s.options)
	err := s.service.Initialize()
	c.Assert(err, jc.ErrorIsNil)

	log.Infof("Creating controller machine")
	_, err = s.BackingState.AddOneMachine(state.MachineTemplate{
		Series:     s.options.Series,
		Jobs:       []state.MachineJob{state.JobManageModel, state.JobHostUnits},
	})
	c.Assert(err, gc.IsNil)

	log.Infof("Starting fake-juju watch loop")
	s.service.Start()
	err = s.service.Ready()
	c.Assert(err, gc.IsNil)
}

func (s *FakeJujuSuite) TearDownTest(c *gc.C) {
	log.Infof("Stopping fake-juju watch loop")

	c.Assert(s.service.Stop(), gc.IsNil)
	ClearFailures()
	s.JujuConnSuite.TearDownTest(c)
}

func (s *FakeJujuSuite) Wait() error {
	return s.service.Wait()
}
