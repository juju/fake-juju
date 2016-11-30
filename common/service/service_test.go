package service_test

import (
	"testing"

	gc "gopkg.in/check.v1"

	"github.com/juju/juju/agent"
	"github.com/juju/juju/state"
	"github.com/juju/juju/testing/factory"
	"github.com/juju/juju/version"
	"github.com/juju/utils"

	coretesting "github.com/juju/juju/juju/testing"
	jujutesting "github.com/juju/juju/testing"

	"../service"
)

type FakeJujuServiceSuite struct {
	coretesting.JujuConnSuite
	service *service.FakeJujuService
}

func (s *FakeJujuServiceSuite) SetUpTest(c *gc.C) {
	s.JujuConnSuite.SetUpTest(c)

	options := &service.FakeJujuOptions{}
	s.service = service.NewFakeJujuService(s.State, s.APIState, options)
}

// The Initialize() method performs various initialization tasks.
func (s *FakeJujuServiceSuite) TestInitialize(c *gc.C) {
	err := s.service.Initialize()
	c.Assert(err, gc.IsNil)
	c.Assert(utils.OutgoingAccessAllowed, gc.Equals, true)

	ports, err := s.State.APIHostPorts()
	c.Assert(err, gc.IsNil)
	c.Assert(string(ports[0][0].SpaceName), gc.Equals, "dummy-provider-network")
}

// The InitializeController() method configures the controller machine.
func (s *FakeJujuServiceSuite) TestInitializeController(c *gc.C) {
	controller := s.Factory.MakeMachine(c, &factory.MachineParams{
		InstanceId: s.service.NewInstanceId(),
		Nonce:      agent.BootstrapNonce,
		Jobs:       []state.MachineJob{state.JobManageModel, state.JobHostUnits},
		Series:     "xenial",
	})
	err := s.service.InitializeController(controller)
	c.Assert(err, gc.IsNil)

	tools, err := controller.AgentTools()
	c.Assert(err, gc.IsNil)
	c.Assert(
		tools.Version.String(),
		gc.Equals,
		version.Current.String()+"-xenial-amd64")
}

var _ = gc.Suite(&FakeJujuServiceSuite{})

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	jujutesting.MgoTestPackage(t)
}
