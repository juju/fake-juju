// Handle changes to action entities

package service

import (
	"github.com/juju/juju/state"
)

// Handle a changed action
func (s *FakeJujuService) handleActionChanged(id string) error {
	log.Infof("Handling changed action %s", id)

	// Get the action
	action, err := s.state.Action(id)
	if err != nil {
		return err
	}

	if action.Status() == state.ActionPending {
		return s.completeAction(action)
	}

	return nil
}

// Start a action (i.e. transition it from waiting to active)
func (s *FakeJujuService) completeAction(action state.Action) error {
	log.Infof("Completing action %s", action.Id())

	output := map[string]interface{}{"output": "action ran successfully"}
	results := state.ActionResults{
		Status:  state.ActionCompleted,
		Results: output,
	}
	_, err := action.Finish(results)
	return err
}
