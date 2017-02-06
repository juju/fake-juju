// Handle changes to machine entities

package service

import (
	"fmt"
	"time"

	"github.com/juju/juju/instance"
	"github.com/juju/juju/network"
	"github.com/juju/juju/state"
	"github.com/juju/juju/status"
	"github.com/juju/juju/version"
	semversion "github.com/juju/version"
)

// Handle a changed machine
func (s *FakeJujuService) handleMachineChanged(id string) error {
	log.Infof("Handling changed machine %s", id)

	// Get the machine
	machine, err := s.state.Machine(id)
	if err != nil {
		return err
	}

	st, err := machine.Status()
	if err != nil {
		return err
	}

	switch st.Status {
	case status.Pending:
		if err := s.startMachine(machine); err != nil {
			return err
		}
	case status.Started:
		if id == "0" && s.ready != nil {
			// Notify the Ready() method
			s.ready <- nil
			close(s.ready)
			s.ready = nil
		}
	}

	return nil
}

// Start a machine (i.e. transition it from pending to started)
func (s *FakeJujuService) startMachine(machine *state.Machine) error {

	log.Infof("Starting machine %s", machine.Id())

	now := time.Now()

	// Set network address
	address := network.NewScopedAddress("127.0.0.1", network.ScopeCloudLocal)
	if err := machine.SetProviderAddresses(address); err != nil {
		return err
	}

	// Set instance state
	if err := machine.SetProvisioned(s.newInstanceId(), "nonce", nil); err != nil {
		return err
	}
	if err := machine.SetInstanceStatus(status.StatusInfo{
		Status:  status.Running,
		Message: "",
		Since:   &now,
	}); err != nil {
		return err
	}

	// Set agent version
	currentVersion := version.Current.String()
	agentVersion, err := semversion.ParseBinary(currentVersion + "-xenial-amd64")
	if err != nil {
		return err
	}
	if err := machine.SetAgentVersion(agentVersion); err != nil {
		return err
	}

	// Set agent status
	if err := machine.SetStatus(status.StatusInfo{
		Status:  status.Started,
		Message: "",
		Since:   &now,
	}); err != nil {
		return err
	}

	// Set agent presence
	if _, err := machine.SetAgentPresence(); err != nil {
		return err
	}
	s.state.StartSync()
	if err := machine.WaitAgentPresence(MediumWait); err != nil {
		return err
	}

	return nil
}

func (s *FakeJujuService) newInstanceId() instance.Id {
	s.instanceCount += 1
	return instance.Id(fmt.Sprintf("id-%d", s.instanceCount))
}
