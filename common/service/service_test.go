package service_test

import (
	"testing"

	gc "gopkg.in/check.v1"

	"github.com/juju/juju/agent"
	"github.com/juju/juju/state"
	"github.com/juju/juju/testing/factory"
	"github.com/juju/juju/version"
	"github.com/juju/utils"

	coretesting "github.com/juju/juju/juju/testing"
	"github.com/juju/juju/status"
	jujutesting "github.com/juju/juju/testing"

	"../service"
)

type FakeJujuServiceSuite struct {
	coretesting.JujuConnSuite
	service *service.FakeJujuService
}

func (s *FakeJujuServiceSuite) SetUpTest(c *gc.C) {
	s.JujuConnSuite.SetUpTest(c)

	options := &service.FakeJujuOptions{
		Mongo:  -1, // Use the MongoDB instance that the suite will setup
		Series: "xenial",
	}
	s.service = service.NewFakeJujuService(s.BackingState, s.APIState, options)
}

// The Initialize() method performs various initialization tasks.
func (s *FakeJujuServiceSuite) TestInitialize(c *gc.C) {
	err := s.service.Initialize()
	c.Assert(err, gc.IsNil)

	// We want to be able to access the charm store
	c.Assert(utils.OutgoingAccessAllowed, gc.Equals, true)

	// There's a space defined
	ports, err := s.State.APIHostPorts()
	c.Assert(err, gc.IsNil)
	c.Assert(string(ports[0][0].SpaceName), gc.Equals, "dummy-provider-network")
}

// The InitializeController() method configures the controller machine.
func (s *FakeJujuServiceSuite) TestInitializeController(c *gc.C) {
	machine := s.Factory.MakeMachine(c, &factory.MachineParams{
		InstanceId: s.service.NewInstanceId(),
		Nonce:      agent.BootstrapNonce,
		Jobs:       []state.MachineJob{state.JobManageModel, state.JobHostUnits},
		Series:     "xenial",
	})
	err := s.service.InitializeController(machine)
	c.Assert(err, gc.IsNil)

	tools, err := machine.AgentTools()
	c.Assert(err, gc.IsNil)
	c.Assert(
		tools.Version.String(),
		gc.Equals,
		version.Current.String()+"-xenial-amd64")

	// The machine machine is configured and started
	machineStatus, err := machine.Status()
	c.Check(err, gc.IsNil)
	c.Check(machineStatus.Status, gc.Equals, status.Started)

	instanceStatus, err := machine.InstanceStatus()
	c.Check(err, gc.IsNil)
	c.Check(instanceStatus.Status, gc.Equals, status.Running)

	s.State.StartSync()
	err = machine.WaitAgentPresence(jujutesting.ShortWait)
	c.Assert(err, gc.IsNil)
	alive, err := machine.AgentPresence()
	c.Assert(err, gc.IsNil)
	c.Assert(alive, gc.Equals, true)
}

// The watch loop can be started and stopped.
func (s *FakeJujuServiceSuite) TestStartAndStopWatchLoop(c *gc.C) {
	s.service.Start()
	err := s.service.Stop()
	c.Assert(err, gc.IsNil)
}

// In case an unexpected error occurs during the watch loop, the Wait()
// method will return it.
func (s *FakeJujuServiceSuite) TestWatchLoopError(c *gc.C) {
	s.service.Start()

	// Close the State object that the watch loop is connected to, this
	// will cause a wather error.
	s.BackingState.Close()

	err := s.service.Wait()
	c.Assert(err.Error(), gc.Equals, "shared state watcher was stopped")
}

var _ = gc.Suite(&FakeJujuServiceSuite{})

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	jujutesting.MgoTestPackage(t)
}
