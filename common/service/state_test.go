package service

import (
	"testing"

	gc "gopkg.in/check.v1"

	"github.com/juju/juju/agent"
	"github.com/juju/juju/state"
	"github.com/juju/juju/testing/factory"
	"github.com/juju/utils"
 	"github.com/juju/juju/version"

	coretesting "github.com/juju/juju/juju/testing"
	jujutesting "github.com/juju/juju/testing"
)

type FakeJujuStateSuite struct {
	coretesting.JujuConnSuite
	state *FakeJujuState
}

func (s *FakeJujuStateSuite) SetUpTest(c *gc.C) {
	s.JujuConnSuite.SetUpTest(c)

	s.state = &FakeJujuState{
		state:   s.State,
		api:     s.APIState,
		options: &FakeJujuOptions{},
	}
}

// The Initialize() method performs various initialization tasks.
func (s *FakeJujuStateSuite) TestInitialize(c *gc.C) {
	err := s.state.Initialize()
	c.Assert(err, gc.IsNil)
	c.Assert(utils.OutgoingAccessAllowed, gc.Equals, true)

	ports, err := s.State.APIHostPorts()
	c.Assert(err, gc.IsNil)
	c.Assert(string(ports[0][0].SpaceName), gc.Equals, "dummy-provider-network")

	//	config, err := s.State.ModelConfig()
	//	c.Assert(err, gc.IsNil)
	//	series, _ := config.DefaultSeries()
	//	c.Assert(defined, gc.Equals, true)
	//	c.Assert(series, gc.Equals, "xenial")
}

// The InitializeController() method configures the controller machine.
func (s *FakeJujuStateSuite) TestInitializeController(c *gc.C) {
	controller := s.Factory.MakeMachine(c, &factory.MachineParams{
		InstanceId: s.state.NewInstanceId(),
		Nonce:      agent.BootstrapNonce,
		Jobs:       []state.MachineJob{state.JobManageModel, state.JobHostUnits},
		Series:     "xenial",
	})
	err := s.state.InitializeController(controller)
	c.Assert(err, gc.IsNil)

	tools, err := controller.AgentTools()
	c.Assert(err, gc.IsNil)
	c.Assert(
		tools.Version.String(),
		gc.Equals,
		version.Current.String() + "-xenial-amd64")
}

var _ = gc.Suite(&FakeJujuStateSuite{})

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	jujutesting.MgoTestPackage(t)
}
