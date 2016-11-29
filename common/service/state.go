package service

import (
	"fmt"

	"github.com/juju/juju/api"
	"github.com/juju/juju/instance"
	"github.com/juju/juju/state"
	"github.com/juju/utils"
 	"github.com/juju/juju/version"
 	"github.com/juju/juju/network"

 	semversion "github.com/juju/version"
)

type FakeJujuState struct {
	state         *state.State
	api           api.Connection
	options       *FakeJujuOptions
	instanceCount int
}

// Main initialization entry point
func (s *FakeJujuState) Initialize() error {

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

	config := map[string]interface{}{"default-series": s.options.series}
	err = s.state.UpdateModelConfig(config, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// Initialize the controller machine (aka machine 0).
func (s *FakeJujuState) InitializeController(controller *state.Machine) error {
	currentVersion := version.Current.String()

	agentVersion, err := semversion.ParseBinary(currentVersion + "-xenial-amd64")
	if err != nil {
		return err
	}

	controller.SetAgentVersion(agentVersion)

	address := network.NewScopedAddress("127.0.0.1", network.ScopeCloudLocal)
	return controller.SetProviderAddresses(address)
}

func (s *FakeJujuState) NewInstanceId() instance.Id {
	s.instanceCount += 1
	return instance.Id(fmt.Sprintf("id-%d", s.instanceCount))
}
