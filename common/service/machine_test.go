package service_test

import (
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/state"
	"github.com/juju/juju/status"
	"github.com/juju/juju/version"

	"../service"
)

// The watch loop starts any new machine in the pending state.
func (s *FakeJujuServiceSuite) TestWatchLoopStartMachine(c *gc.C) {
	s.service.Start()
	defer s.service.Stop()

	// Simulate a new machine being added.
	machine, err := s.BackingState.AddOneMachine(state.MachineTemplate{
		Series: "xenial",
		Jobs:   []state.MachineJob{state.JobHostUnits},
	})
	c.Assert(err, gc.IsNil)

	// This will wait a couple of seconds for the agent to be up, or
	// bail out if it doesn't.
	s.BackingState.StartSync()
	err = machine.WaitAgentPresence(service.MediumWait)
	c.Assert(err, gc.IsNil)

	// The agent is indeed alive.
	alive, err := machine.AgentPresence()
	c.Assert(err, gc.IsNil)
	c.Assert(alive, gc.Equals, true)

	// The machine is started.
	machineStatus, err := machine.Status()
	c.Check(err, gc.IsNil)
	c.Check(machineStatus.Status, gc.Equals, status.Started)

	// The associated dummy instance is running.
	instanceStatus, err := machine.InstanceStatus()
	c.Check(err, gc.IsNil)
	c.Check(instanceStatus.Status, gc.Equals, status.Running)

	// Tools are at the correct version
	machine, err = s.BackingState.Machine("0") // Grab again to ensure it's fresh
	c.Assert(err, gc.IsNil)
	tools, err := machine.AgentTools()
	c.Assert(err, gc.IsNil)
	c.Assert(
		tools.Version.String(),
		gc.Equals,
		version.Current.String()+"-xenial-amd64")

}
