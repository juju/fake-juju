package main

import (
	"fmt"
	"testing"
	gc "gopkg.in/check.v1"
	"os"
	"os/exec"
	"bufio"
	"time"
	"path/filepath"
	"syscall"
	"io"
	"io/ioutil"
	"errors"
	"log"
	"encoding/json"

	"github.com/juju/juju/environs"
	"github.com/juju/juju/environs/configstore"
	"github.com/juju/juju/juju/osenv"
	"github.com/juju/juju/state"
	"github.com/juju/juju/agent"
	"github.com/juju/juju/network"
	"github.com/juju/juju/api"
	"github.com/juju/names"
	_ "github.com/juju/juju/provider/maas"
	coretesting "github.com/juju/juju/testing"
	jujutesting "github.com/juju/juju/juju/testing"
	"github.com/juju/juju/testing/factory"
	"github.com/juju/juju/version"
	corecharm "gopkg.in/juju/charm.v5/charmrepo"
	"github.com/juju/juju/instance"
	goyaml "gopkg.in/yaml.v1"
)

func main() {
	if len(os.Args) > 1 {
		code := 0
		err := handleCommand(os.Args[1])
		if err != nil {
			fmt.Println(err.Error())
			code = 1
		}
		os.Exit(code)
	}
	t := &testing.T{}
	coretesting.MgoTestPackage(t)
}

type processInfo struct {
        WorkDir string
	EndpointAddr string
	Uuid string
	CACert string
}

func handleCommand(command string) error {
	if command == "bootstrap" {
		return bootstrap()
	}
	if command == "api-endpoints" {
		return apiEndpoints()
	}
	if command == "api-info" {
		return apiInfo()
	}
	if command == "destroy-environment" {
		return destroyEnvironment()
	}
	return errors.New("command not found")
}

func bootstrap() error {
	envName, password, err := environmentNameAndPassword()
	if err != nil {
		return err
	}
	command := exec.Command(os.Args[0])
	command.Env = os.Environ()
	command.Env = append(command.Env, "ADMIN_PASSWORD=" + password)
        stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	command.Start()	
	apiInfo, err := parseApiInfo(envName, stdout)
	if err != nil {
		return err
	}
	dialOpts := api.DialOpts{
		DialAddressInterval: 50 * time.Millisecond,
		Timeout:             5 * time.Second,
		RetryDelay:          2 * time.Second,
	}
	state, err := api.Open(apiInfo, dialOpts)
	if err != nil {
		return err
	}
	client := state.Client()
	watcher, err := client.WatchAll()
	if err != nil {
		return err
	}
	deltas, err := watcher.Next()
	if err != nil {
		return err
	}
	for _, delta := range deltas {
		entityId := delta.Entity.EntityId()
		if entityId.Kind == "machine" {
			if entityId.Id == "0" {
				return nil
			}
		}
	}
	return errors.New("invalid delta")
}

func apiEndpoints() error {
	info, err := readProcessInfo()
	if err != nil {
		return err
	}
	fmt.Println(info.EndpointAddr)
	return nil
}

func apiInfo() error {
	info, err := readProcessInfo()
	if err != nil {
		return err
	}
	fmt.Printf("{\"environ-uuid\": \"%s\", \"state-servers\": [\"%s\"]}\n", info.Uuid, info.EndpointAddr)
	return nil
}

func destroyEnvironment() error {
	info, err := readProcessInfo()
	if err != nil {
		return err
	}
	fifoPath := filepath.Join(info.WorkDir, "fifo")
        fd, err := os.OpenFile(fifoPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer fd.Close()
	_, err = fd.WriteString("destroy\n")
	if err != nil {
		return err
	}
	return nil
}

func environmentNameAndPassword() (string, string, error) {
	jujuHome := os.Getenv("JUJU_HOME")
	osenv.SetJujuHome(jujuHome)
	environs, err := environs.ReadEnvirons(
		filepath.Join(jujuHome, "environments.yaml"))
	if err != nil {
		return "", "", err
	}
	envName := environs.Names()[0]
	config, err := environs.Config(envName)
	if err != nil {
		return "", "", err
	}
	return envName, config.AdminSecret(), nil
}

func parseApiInfo(envName string, stdout io.ReadCloser) (*api.Info, error) {
        buffer := bufio.NewReader(stdout)
	line, _, err := buffer.ReadLine()
	if err != nil {
		return nil, err
	}
	uuid := string(line)
	environTag := names.NewEnvironTag(uuid)
	line, _, err = buffer.ReadLine()
	if err != nil {
		return nil, err
	}
	workDir := string(line)
	store, err := configstore.NewDisk(workDir)
	if err != nil {
		return nil, err
	}
	info, err := store.ReadInfo("dummyenv")
	if err != nil {
		return nil, err
	}
	credentials := info.APICredentials()
	endpoint := info.APIEndpoint()
	addresses := endpoint.Addresses
	apiInfo := &api.Info{
		Addrs: addresses,
		Tag: names.NewLocalUserTag(credentials.User),
		Password: credentials.Password,
		CACert: endpoint.CACert,
		EnvironTag: environTag,
	}
	err = writeProcessInfo(envName, &processInfo{
		WorkDir: workDir,
		EndpointAddr: addresses[0],
		Uuid: uuid,
		CACert: endpoint.CACert,
	})
	if err != nil {
		return nil, err
	}
	return apiInfo, nil
}

func readProcessInfo() (*processInfo, error) {
	infoPath := filepath.Join(os.Getenv("JUJU_HOME"), "fakejuju")
	data, err := ioutil.ReadFile(infoPath)
	if err != nil {
		return nil, err
	}
	info := &processInfo{}
	err = goyaml.Unmarshal(data, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func writeProcessInfo(envName string, info *processInfo) error {
	jujuHome := os.Getenv("JUJU_HOME")
	infoPath := filepath.Join(jujuHome, "fakejuju")
	logPath := filepath.Join(jujuHome, "fake-juju.log")
	caCertPath := filepath.Join(jujuHome, "cert.ca")
	envPath := filepath.Join(jujuHome, "environments")
	os.Mkdir(envPath, 0755)
	jEnvPath := filepath.Join(envPath, envName + ".jenv")
	data, _ := goyaml.Marshal(info)
	err := os.Symlink(filepath.Join(info.WorkDir, "fake-juju.log"), logPath)
	if err != nil {
		return err
	}
	err = os.Symlink(filepath.Join(info.WorkDir, "environments/dummyenv.jenv"), jEnvPath)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(infoPath, data, 0644)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(caCertPath, []byte(info.CACert), 0644)
}

type FakeJujuSuite struct {
	jujutesting.JujuConnSuite

	instanceCount int
	machineStarted map[string]bool
	fifoPath string
	logFile *os.File
}

var _ = gc.Suite(&FakeJujuSuite{})

func (s *FakeJujuSuite) SetUpTest(c *gc.C) {
	s.JujuConnSuite.SetUpTest(c)

	ports := s.APIState.APIHostPorts()
	ports[0][0].NetworkName = "dummy-provider-network"
	err := s.State.SetAPIHostPorts(ports)
	c.Assert(err, gc.IsNil)

	s.machineStarted = make(map[string]bool)
	s.PatchValue(&corecharm.CacheDir, c.MkDir())
	password := "dummy-password"
	if os.Getenv("ADMIN_PASSWORD") != "" {
		password = os.Getenv("ADMIN_PASSWORD")
	}
	_, err = s.State.AddUser("admin", "Admin", password, "dummy-admin")
	c.Assert(err, gc.IsNil)
	_, err = s.State.AddEnvironmentUser(
		names.NewLocalUserTag("admin"), names.NewLocalUserTag("dummy-admin"), "Admin")
	c.Assert(err, gc.IsNil)

	// Create a machine to manage the environment.
	stateServer := s.Factory.MakeMachine(c, &factory.MachineParams{
		InstanceId: s.newInstanceId(),
		Nonce:    agent.BootstrapNonce,
		Jobs:       []state.MachineJob{state.JobManageEnviron, state.JobHostUnits},
		Series: "trusty",
	})
	c.Assert(stateServer.SetAgentVersion(version.Current), gc.IsNil)
	address := network.NewScopedAddress("127.0.0.1", network.ScopeCloudLocal)
	c.Assert(stateServer.SetProviderAddresses(address), gc.IsNil)
	c.Assert(stateServer.SetStatus(state.StatusStarted, "", nil), gc.IsNil)
	_, err = stateServer.SetAgentPresence()
	c.Assert(err, gc.IsNil)
	s.State.StartSync()
	err = stateServer.WaitAgentPresence(coretesting.LongWait)
	c.Assert(err, gc.IsNil)

	apiInfo := s.APIInfo(c)
	//fmt.Println(apiInfo.Addrs[0])
	jujuHome := osenv.JujuHome()
	fmt.Println(apiInfo.EnvironTag.Id())
	fmt.Println(jujuHome)

	binPath := filepath.Join(jujuHome, "bin")
	os.Mkdir(binPath, 0755)
	fakeSSHData := []byte("#!/bin/sh\nsleep 1\n")
	fakeSSHPath := filepath.Join(binPath, "ssh")
	err = ioutil.WriteFile(fakeSSHPath, fakeSSHData, 0755)
	c.Assert(err, gc.IsNil)
	os.Setenv("PATH", binPath + ":" + os.Getenv("PATH"))

	s.fifoPath = filepath.Join(jujuHome, "fifo")
	syscall.Mknod(s.fifoPath, syscall.S_IFIFO|0666, 0)


	// Logging
	logPath := filepath.Join(jujuHome, "fake-juju.log")
	s.logFile, err = os.OpenFile(logPath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	c.Assert(err, gc.IsNil)

	log.SetOutput(s.logFile)
	log.Println("Set up completed")

}

func (s *FakeJujuSuite) TearDownTest(c *gc.C) {
	s.JujuConnSuite.TearDownTest(c)
	s.logFile.Close()
}

func (s *FakeJujuSuite) TestStart(c *gc.C) {
	watcher := s.State.Watch()
	go func() {
		fd, err := os.Open(s.fifoPath)
		c.Assert(err, gc.IsNil)
		scanner := bufio.NewScanner(fd)
		scanner.Scan()
		watcher.Stop()
	}()
	for {
		deltas, err := watcher.Next()
		log.Println("Got deltas")
		if err != nil {
			if err.Error() == "watcher was stopped" {
				log.Println("Watcher stopped")
				break
			}
			log.Println("Unexpected error", err.Error())
		}
		c.Assert(err, gc.IsNil)
		for _, d := range deltas {

			entity, err := json.MarshalIndent(d.Entity, "", "  ")
			c.Assert(err, gc.IsNil)
			verb := "change"
			if d.Removed {
				verb = "remove"
			}
			log.Println("Processing delta", verb, d.Entity.EntityId().Kind, string(entity[:]))

			entityId := d.Entity.EntityId()
			if entityId.Kind == "machine" {
				machineId := entityId.Id
				c.Assert(s.handleAddMachine(machineId), gc.IsNil)
			}
			if entityId.Kind == "unit" {
				unitId := entityId.Id
				c.Assert(s.handleAddUnit(unitId), gc.IsNil)
			}
			log.Println("Done processing delta")
		}
	}
}

func (s *FakeJujuSuite) handleAddMachine(id string) error  {
	machine, err := s.State.Machine(id)
	if err != nil {
		return err
	}
	if instanceId, _ := machine.InstanceId(); instanceId == "" {
		err = machine.SetProvisioned(s.newInstanceId(), agent.BootstrapNonce, nil)
		if err != nil {
			return err
		}
		address := network.NewScopedAddress("127.0.0.1", network.ScopeCloudLocal)
		err = machine.SetProviderAddresses(address)
		if err != nil {
			return err
		}
	}
	status, _ := machine.Status()
	if status.Status == state.StatusPending {
		if err = s.startMachine(machine); err != nil {
			return err
		}
	} else if status.Status == state.StatusStarted {
		if _, ok := s.machineStarted[id]; !ok {
			s.machineStarted[id] = true
			if err = s.startUnits(machine); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *FakeJujuSuite) handleAddUnit(id string) error  {
	unit, err := s.State.Unit(id)
	if err != nil {
		return err
	}
	machineId, err := unit.AssignedMachineId()
	if err != nil {
		return nil
	}
	log.Println("Got machineId", machineId)
	machine, err := s.State.Machine(machineId)
	if err != nil {
		return err
	}
	machineStatus, _ := machine.Status()
	if machineStatus.Status != state.StatusStarted {
		return nil
	}
	status, _ := unit.Status()
	if status.Status != state.StatusActive {
		if err = s.startUnit(unit); err != nil {
			return err
		}
	}
	return nil
}

func (s *FakeJujuSuite) startMachine(machine *state.Machine) error  {
	time.Sleep(500 * time.Millisecond)
	err := machine.SetStatus(state.StatusStarted, "", nil)
	if err != nil {
		return err
	}
	err = machine.SetAgentVersion(version.Current)
	if err != nil {
		return err
	}
	_, err = machine.SetAgentPresence()
	if err != nil {
		return err
	}
	s.State.StartSync()
	err = machine.WaitAgentPresence(coretesting.LongWait)
	if err != nil {
		return err
	}
	return nil
}

func (s *FakeJujuSuite) startUnits(machine *state.Machine) error  {
	units, err := machine.Units()
	if err != nil {
		return err
	}
	return nil
	for _, unit := range units {
		unitStatus, _ := unit.Status()
		if unitStatus.Status != state.StatusActive {
			if err = s.startUnit(unit); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *FakeJujuSuite) startUnit(unit *state.Unit) error  {
	err := unit.SetStatus(state.StatusActive, "", nil)
	if err != nil {
		return err
	}
	_, err = unit.SetAgentPresence()
	if err != nil {
		return err
	}
	s.State.StartSync()
	err = unit.WaitAgentPresence(coretesting.LongWait)
	if err != nil {
		return err
	}
	err = unit.SetAgentStatus(state.StatusIdle, "", nil)
	if err != nil {
		return err
	}
	return nil
}


func (s *FakeJujuSuite) newInstanceId() instance.Id {
	s.instanceCount += 1
	return instance.Id(fmt.Sprintf("id-%d", s.instanceCount))
}
