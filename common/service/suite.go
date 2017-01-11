package service

import (
	"os"
	"os/signal"
	"syscall"

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
	log.Infof("Initializing test suite")
	s.JujuConnSuite.SetUpTest(c)

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
	log.Infof("Initializing TestStart of the test suite")
	// TODO: implement actual fake-juju logic. For now we just wait forever
	// until SIGINT (ctrl-c) or SIGTERM is received.
	channel := make(chan os.Signal, 2)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
	<-channel
	log.Infof("Terminating TestStart")
}
