// Handle changes to unit entities

package service

import (
	"time"

	"github.com/juju/juju/state"
	"github.com/juju/juju/status"
)

// Handle a changed unit
func (s *FakeJujuService) handleUnitChanged(id string) error {
	log.Infof("Handling changed unit %s", id)

	// Get the unit
	unit, err := s.state.Unit(id)
	if err != nil {
		return err
	}

	workloadStatus, err := unit.Status()
	if err != nil {
		return err
	}

	agentStatus, err := unit.AgentStatus()
	if err != nil {
		return err
	}

	if agentStatus.Status == status.Allocating {
		return s.startUnit(unit)
	} else if workloadStatus.Status != status.Error && ShouldFail("unit", id) {
		return s.errorUnit(unit)
	}

	return nil
}

// Start a unit (i.e. transition it from waiting to active)
func (s *FakeJujuService) startUnit(unit *state.Unit) error {
	log.Infof("Starting unit %s", unit.Name())

	now := time.Now()

	if _, err := unit.AssignedMachineId(); err != nil {
		if s.options.AutoStartMachines {
			// If the unit has no machine assigned, we'll create one
			// for it. We should eventually get another delta about
			// the unit, and at that point this if branch won't be
			// taken anymore, because there's an assigned machine.
			return s.addMachineForUnit(unit)
		} else {
			// Just no-op, we'll retry as soon as the unit gets
			// associated with a machine.
			return nil
		}
	}

	if err := unit.SetAgentStatus(status.StatusInfo{
		Status:  status.Idle,
		Message: "",
		Since:   &now,
	}); err != nil {
		return err
	}

	if err := unit.SetStatus(status.StatusInfo{
		Status:  status.Active,
		Message: "",
		Since:   &now,
	}); err != nil {
		return err
	}

	// Set agent presence
	if _, err := unit.SetAgentPresence(); err != nil {
		return err
	}
	s.state.StartSync()
	if err := unit.WaitAgentPresence(MediumWait); err != nil {
		return err
	}
	return nil
}

// Mark a unit as failed (i.e. transition it to the errored state)
func (s *FakeJujuService) errorUnit(unit *state.Unit) error {
	log.Infof("Erroring unit %s", unit.Name())

	now := time.Now()

	return unit.SetAgentStatus(status.StatusInfo{
		Status:  status.Error,
		Message: "unit errored",
		Since:   &now,
	})
}

// Create a machine for a unit that doesn't have one yet
func (s *FakeJujuService) addMachineForUnit(unit *state.Unit) error {
	log.Infof("Adding new machine for unit %s", unit.Name())
	machine, err := s.state.AddOneMachine(state.MachineTemplate{
		Series: s.options.Series,
		Jobs:   []state.MachineJob{state.JobHostUnits},
	})
	if err != nil {
		return err
	}
	return unit.AssignToMachine(machine)
}
