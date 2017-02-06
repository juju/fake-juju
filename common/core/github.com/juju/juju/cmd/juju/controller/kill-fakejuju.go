package controller

import (
	"github.com/juju/juju/api"
)

// Custom destroy-controller logic for fake juju.
func (c *killCommand) fakeJujuKillController() error {

	// Clean all JUJU_DATA controller-related files
	controller := c.ControllerName()
	store := c.ClientStore()
	if err := store.RemoveController(controller); err != nil {
		return err
	}

	client, err := api.NewFakeJujuClient()
	if err != nil {
		return err
	}
	return client.Destroy()
}
