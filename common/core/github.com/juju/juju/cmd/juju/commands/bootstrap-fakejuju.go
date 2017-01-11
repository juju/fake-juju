package commands

import (
	"fmt"
	"os"
	"strconv"

	"github.com/juju/errors"

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

	return nil
}

// Write a fake controllers.yaml
func writeControllersFile(store jujuclient.ClientStore, controller string) error {
	port, err := getFakeJujudPort()
	if err != nil {
		return err
	}
	details := jujuclient.ControllerDetails{
		ControllerUUID: testing.ControllerTag.Id(),
		CACert:         testing.CACert,
		AgentVersion:   version.Current.String(),
		APIEndpoints:   []string{fmt.Sprintf("localhost:%d", port)},
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

// Figure the port that fake-jujud is listening to.
func getFakeJujudPort() (port int, err error) {
	port = 17079  // the default
	if os.Getenv("FAKE_JUJUD_PORT") != "" {
		port, err = strconv.Atoi(os.Getenv("FAKE_JUJUD_PORT"))
		if err != nil {
			return 0, errors.Annotate(err, "invalid port number")
		}
	}
	return port, nil
}
