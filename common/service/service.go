// Core logic for the fake-jujud process.
//
// It mainly implements a watch loop that reacts to changes in juju
// entities. For example it will automatically transition new machines
// from the "pending" to the "started" state.
//
// The logic in this file is only about the top-level watch loop. The
// logic for handling specific entities is implemented in the files
// named after the entities (machine.go, unit.go etc).
package service

import (
	"io"
	"time"

	"github.com/juju/juju/api"
	"github.com/juju/juju/state"
	"github.com/juju/juju/state/multiwatcher"
	"github.com/juju/loggo"
	"github.com/juju/utils"
)

// Value used when waiting for events like agent presence synchronization.
const MediumWait = 2 * time.Second

// Runtime options for the fake-juju service
type FakeJujuOptions struct {
	Output io.Writer   // Where to write logs
	Level  loggo.Level // The log level

	// MongoDB port to connect to (on localhost). If set to 0,
	// a dedicated MongoDB child process will be spawned. If
	// set to -1, no setup will be done at all (for tests)
	Mongo int

	Port   int    // Port for the API server
	Series string // Default Ubuntu series

	// Whether to use a random certificate for the juju API server. This
	// is set to true only by unit tests, where we want to leverage the
	// custom certificate that JujuConnSuite generates, and that the
	// rest of testing facilities execpt.
	UseRandomCert bool

	// Whether to automatically start machines for units that don't appear
	// to have one.
	AutoStartMachines bool
}

// The core fake-juju service
func NewFakeJujuService(
	state *state.State, api api.Connection, options *FakeJujuOptions) *FakeJujuService {

	return &FakeJujuService{
		state:   state,
		api:     api,
		options: options,
		ready:   make(chan error, 1),
		done:    make(chan error, 1),
	}
}

type FakeJujuService struct {

	// The "backing" State object, i.e. the one that API server
	// uses (as opposed to the one that the test suite uses).
	state *state.State

	api     api.Connection
	options *FakeJujuOptions
	watcher *state.Multiwatcher

	// Monotonically incrementing counter for generating instance IDs.
	instanceCount int

	// A channel that will be filled with nil if machine 0 could be
	// started cleanly, or with an error otherwise.
	ready chan error

	// A channel that will be filled with nil if the FakeJujuService
	// completes cleanly, or with an error otherwise.
	done chan error
}

// Main initialization entry point
func (s *FakeJujuService) Initialize() error {
	log.Infof("Initializing the service")

	// Juju needs internet access to reach the charm store.  This is
	// necessary to download charmstore charms (e.g. when adding a
	// service.  In the case of downloading charms, the controller
	// does not try to download if the charm is already in the database.
	// So internet access could go back to disabled if we were to add
	// the charm directly before using any API methods that try to
	// download it.  Here are some ideas on how to do that:
	//  * use the API's "upload charm" HTTP endpoint (POST on /charms);
	//    this requires disabling the prohibition against uploading
	//    charmstore charms (which we've done already in fake-juju);
	//    this could be accomplished using txjuju, python-jujuclient,
	//    or manually with httplib (all have auth complexity)
	//  * add an "upload-charm" command to fake-juju that forces a
	//    charmstore charm into the DB
	//  * add the service with a "local" charm schema and then forcibly
	//    change the charm's schema to "cs" in the DB
	// In the meantime, we allow fake-juju to have internet access.
	// XXX (lp:1639276): Remove this special case.
	utils.OutgoingAccessAllowed = true

	ports := s.api.APIHostPorts()
	ports[0][0].SpaceName = "dummy-provider-network"
	err := s.state.SetAPIHostPorts(ports)
	if err != nil {
		return err
	}

	config := map[string]interface{}{"default-series": s.options.Series}
	err = s.state.UpdateModelConfig(config, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// Start the service. It will watch for changes and react accordingly.
func (s *FakeJujuService) Start() {
	s.watcher = s.state.Watch()
	go s.watch()
}

// Wait for the service to be ready, i.e. wait for machine 0 to transition
// to the "started" state.
func (s *FakeJujuService) Ready() error {
	return <-s.ready
}

// Stop the service, cancelling our delta watcher. This method will wait
// for the watch loop to terminate, and return any error occurring while
// shutting down.
func (s *FakeJujuService) Stop() error {
	if err := s.watcher.Stop(); err != nil {
		return err
	}
	return s.Wait()
}

// Wait for the service to terminate, either cleanly or with an error.
func (s *FakeJujuService) Wait() error {
	return <-s.done
}

// Watch the model and react to changes. The loop will terminate when
// the Stop() method is called, or an unexpected error occurs.
func (s *FakeJujuService) watch() {
	for {
		deltas, err := s.watcher.Next()
		if err != nil {
			if err.Error() != state.ErrStopped.Error() {
				log.Errorf("Watcher error: %s", err.Error())
				s.done <- err
			}
			break
		}
		for _, delta := range deltas {
			if err := s.handleDelta(delta); err != nil {
				log.Errorf("Delta error: %s (%v)", err.Error(), delta)
				s.done <- err
			}
		}
	}
	log.Infof("Watch loop terminated")

	// This will unblock any caller of Wait(), and make it return "nil"
	// to signal a clean termination.
	close(s.done)
}

// Handle an entity delta
func (s *FakeJujuService) handleDelta(delta multiwatcher.Delta) error {
	entity := delta.Entity.EntityId()
	log.Infof("Delta for %s-%s (removed: %t)", entity.Kind, entity.Id, delta.Removed)
	if delta.Removed {
		return nil
	} else {
		return s.handleEntityChanged(entity)
	}
}

// Handle a changed entity
func (s *FakeJujuService) handleEntityChanged(entity multiwatcher.EntityId) error {
	switch entity.Kind {
	case "machine":
		return s.handleMachineChanged(entity.Id)
	case "unit":
		return s.handleUnitChanged(entity.Id)
	case "action":
		return s.handleActionChanged(entity.Id)
	default:
		log.Infof("Ignoring kind %s", entity.Kind)
		return nil
	}
}
