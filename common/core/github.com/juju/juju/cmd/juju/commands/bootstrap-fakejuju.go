package commands

import (
	"fmt"

	"github.com/juju/juju/api"
	"github.com/juju/juju/cmd/modelcmd"
	"github.com/juju/juju/jujuclient"
	"github.com/juju/juju/testing"
	"github.com/juju/juju/version"
)

// Custom bootstrap logic for fake juju, which essentially:
//
// - populates the JUJU_DATA directory with files pointing to the
//   fake-jujud process, as identified by the port number in the
//   FAKE_JUJUD_PORT environment variable.
//
// - connects to fake-jujud an gets the agent version to ensure that
//   all parameters are correct.
//
// Once the bootstrap is complete, the CLI can be used like a regular
// juju one.
func (c *bootstrapCommand) fakeJujuBootstrap() error {
	store := c.ClientStore()
	controller := c.controllerName
	model := "controller"

	logger.Debugf("bootstrapping %s:%s", controller, model)

	if err := testing.SetCerts(); err != nil {
		return err
	}

	if err := writeControllersFile(store, controller); err != nil {
		return err
	}

	if err := writeAccountsFile(store, controller); err != nil {
		return err
	}

	if err := writeModelsFile(store, controller, model); err != nil {
		return err
	}

	if err := store.SetCurrentModel(controller, model); err != nil {
		return err
	}

	// Connect to fake-jujud and create a new controller
	if err := performBootstrap(); err != nil {
		return err
	}

	// Ensure that the setup is valid
	if err := c.SetModelName(modelcmd.JoinModelName(controller, model)); err != nil {
		return err
	}
	client, err := c.NewAPIClient()
	if err != nil {
		return err
	}
	version, err := client.AgentVersion()
	if err != nil {
		return err
	}
	logger.Debugf("fake-jujud agent version %s", version.String())

	return nil
}

// Write a fake controllers.yaml
func writeControllersFile(store jujuclient.ClientStore, controller string) error {
	port, err := api.GetFakeJujudPort()
	if err != nil {
		return err
	}
	details := jujuclient.ControllerDetails{
		ControllerUUID: testing.ControllerTag.Id(),
		CACert:         testing.CACert,
		AgentVersion:   version.Current.String(),
		APIEndpoints:   []string{fmt.Sprintf("localhost:%d", port - 1)},
	}
	return store.AddController(controller, details)
}

// Write a fake accounts.yaml
func writeAccountsFile(store jujuclient.ClientStore, controller string) error {
	details := jujuclient.AccountDetails{
		User:     "admin",
		Password: "dummy-secret",
	}
	return store.UpdateAccount(controller, details)
}

// Write a fake models.yaml
func writeModelsFile(store jujuclient.ClientStore, controller, model string) error {
	details := jujuclient.ModelDetails{
		ModelUUID: testing.ModelTag.Id(),
	}
	return store.UpdateModel(controller, model, details)
}

// Perform a fake bootstrap by connecting to the faje-jujud service
// using the control API.
func performBootstrap() error {

	// Perform a bootstrap request against fake-juju
	client, err := api.NewFakeJujuClient()
	if err != nil {
		return err
	}
	return client.Bootstrap()
}
