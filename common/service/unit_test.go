package service_test

import (
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/state"
	"github.com/juju/juju/status"
	"github.com/juju/juju/testing/factory"

	"../service"
)

// The watch loop starts any new unit in the allocating state.
func (s *FakeJujuServiceSuite) TestWatchLoopStartUnit(c *gc.C) {
	s.service.Start()
	defer s.service.Stop()

	// Simulate a new unit being added.
	machine, err := s.BackingState.AddOneMachine(state.MachineTemplate{
		Series: "quantal",
		Jobs:   []state.MachineJob{state.JobHostUnits},
	})
	c.Assert(err, gc.IsNil)
	charm := s.Factory.MakeCharm(c, &factory.CharmParams{
		Series: "quantal",
	})
	application := s.Factory.MakeApplication(c, &factory.ApplicationParams{
		Charm: charm,
	})
	unit, err := application.AddUnit()
	c.Assert(err, gc.IsNil)
	c.Assert(unit.AssignToMachine(machine), gc.IsNil)

	// This will wait a couple of seconds for the agent to be up, or
	// bail out if it doesn't.
	unit, err = s.BackingState.Unit(unit.Name()) // Grab again to ensure it's fresh
	c.Assert(err, gc.IsNil)
	s.BackingState.StartSync()
	err = unit.WaitAgentPresence(service.MediumWait)
	c.Assert(err, gc.IsNil)

	// The agent is indeed alive.
	alive, err := unit.AgentPresence()
	c.Assert(err, gc.IsNil)
	c.Assert(alive, gc.Equals, true)

	// The agent is idle
	agentStatus, err := unit.AgentStatus()
	c.Check(err, gc.IsNil)
	c.Check(agentStatus.Status, gc.Equals, status.Idle)

	// The workload is active
	workloadStatus, err := unit.Status()
	c.Check(err, gc.IsNil)
	c.Check(workloadStatus.Status, gc.Equals, status.Active)
}
