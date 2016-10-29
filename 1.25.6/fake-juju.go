package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	gc "gopkg.in/check.v1"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/juju/juju/agent"
	"github.com/juju/juju/api"
	"github.com/juju/juju/environs"
	"github.com/juju/juju/environs/config"
	"github.com/juju/juju/environs/configstore"
	"github.com/juju/juju/instance"
	"github.com/juju/juju/juju/osenv"
	jujutesting "github.com/juju/juju/juju/testing"
	"github.com/juju/juju/network"
	_ "github.com/juju/juju/provider/maas"
	"github.com/juju/juju/state"
	coretesting "github.com/juju/juju/testing"
	"github.com/juju/juju/testing/factory"
	"github.com/juju/juju/version"
	"github.com/juju/loggo"
	"github.com/juju/names"
	"github.com/juju/utils"
	corecharm "gopkg.in/juju/charm.v5/charmrepo"
	goyaml "gopkg.in/yaml.v1"
)

const (
	envDataDir = "FAKE_JUJU_DATA_DIR"
)

func main() {
	code := 0
	if len(os.Args) > 1 {
		err := handleCommand(os.Args[1])
		if err != nil {
			fmt.Println(err.Error())
			code = 1
		}
	} else {
		// This kicks off the daemon.  See FakeJujuSuite below.
		t := &testing.T{}
		coretesting.MgoTestPackage(t)
	}
	os.Exit(code)
}

func handleCommand(command string) error {
	filenames := newFakeJujuFilenames("", "", "")
	if command == "bootstrap" {
		return bootstrap(filenames)
	}
	if command == "api-endpoints" {
		return apiEndpoints(filenames)
	}
	if command == "api-info" {
		return apiInfo(filenames)
	}
	if command == "destroy-environment" {
		return destroyEnvironment(filenames)
	}
	return errors.New("command not found")
}

func bootstrap(filenames fakejujuFilenames) (returnedErr error) {
	if err := filenames.ensureDirsExist(); err != nil {
		return err
	}
	envName, config, err := environmentNameAndConfig()
	if err != nil {
		return err
	}
	password := config.AdminSecret()

	command := exec.Command(os.Args[0])
	command.Env = os.Environ()
	command.Env = append(command.Env, "ADMIN_PASSWORD="+password)
	defaultSeries, _ := config.DefaultSeries()
	command.Env = append(command.Env, "DEFAULT_SERIES="+defaultSeries)
	command.Env = append(command.Env, envDataDir+"="+filenames.datadir)
	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	command.Start()

	var whence string
	defer func() {
		if returnedErr != nil {
			if err := destroyEnvironment(filenames); err != nil {
				fmt.Printf("could not destroy environment when %s failed: %v\n", whence, err)
			}
			returnedErr = fmt.Errorf("bootstrap failed while %s: %v", whence, returnedErr)
		}
	}()

	result, err := parseApiInfo(stdout)
	if err != nil {
		whence = "parsing bootstrap result"
		return err
	}
	// Get the API info before changing it.  The new values might
	// not work yet.
	apiInfo := result.apiInfo()
	// We actually want to report the API user we added in SetUpTest().
	result.username = "admin"
	if password != "" {
		result.password = password
	}
	if err := result.apply(filenames, envName); err != nil {
		whence = "setting up fake-juju files"
		return err
	}

	whence = "waiting-for-ready"
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

func apiEndpoints(filenames fakejujuFilenames) error {
	info, err := readProcessInfo(filenames)
	if err != nil {
		return err
	}
	fmt.Println(info.EndpointAddr)
	return nil
}

func apiInfo(filenames fakejujuFilenames) error {
	info, err := readProcessInfo(filenames)
	if err != nil {
		return err
	}
	fmt.Printf("{\"user\": \"%s\", \"password\": \"%s\", \"environ-uuid\": \"%s\", \"state-servers\": [\"%s\"]}\n", info.Username, info.Password, info.Uuid, info.EndpointAddr)
	return nil
}

func destroyEnvironment(filenames fakejujuFilenames) error {
	fd, err := os.OpenFile(filenames.fifoFile(), os.O_APPEND|os.O_WRONLY, 0600)
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

func environmentNameAndConfig() (string, *config.Config, error) {
	jujuHome := os.Getenv("JUJU_HOME")
	osenv.SetJujuHome(jujuHome)
	environs, err := environs.ReadEnvirons(
		filepath.Join(jujuHome, "environments.yaml"))
	if err != nil {
		return "", nil, err
	}
	envName := environs.Names()[0]
	config, err := environs.Config(envName)
	if err != nil {
		return "", nil, err
	}
	return envName, config, nil
}

// processInfo holds all the information that fake-juju uses internally.
type processInfo struct {
	Username     string
	Password     string
	WorkDir      string
	EndpointAddr string
	Uuid         string
	CACert       []byte
}

func readProcessInfo(filenames fakejujuFilenames) (*processInfo, error) {
	infoPath := filenames.infoFile()
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

func (info processInfo) write(infoPath string) error {
	data, _ := goyaml.Marshal(&info)
	if err := ioutil.WriteFile(infoPath, data, 0644); err != nil {
		return err
	}
	return nil
}

// fakejujuFilenames encapsulates the paths to all the directories and
// files that are relevant to fake-juju.
type fakejujuFilenames struct {
	datadir string
	logsdir string
}

func newFakeJujuFilenames(datadir, logsdir, jujucfgdir string) fakejujuFilenames {
	if datadir == "" {
		datadir = os.Getenv(envDataDir)
		if datadir == "" {
			if jujucfgdir == "" {
				jujucfgdir = os.Getenv("JUJU_HOME")
			}
			datadir = jujucfgdir
		}
	}
	if logsdir == "" {
		logsdir = os.Getenv("FAKE_JUJU_LOGS_DIR")
		if logsdir == "" {
			logsdir = datadir
		}
	}
	return fakejujuFilenames{datadir, logsdir}
}

func (fj fakejujuFilenames) ensureDirsExist() error {
	if err := os.MkdirAll(fj.datadir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(fj.logsdir, 0755); err != nil {
		return err
	}
	return nil
}

// infoFile() returns the path to the file that fake-juju uses as
// its persistent storage for internal data.
func (fj fakejujuFilenames) infoFile() string {
	return filepath.Join(fj.datadir, "fakejuju")
}

// logsFile() returns the path to the file where fake-juju writes
// its logs.  Note that the normal Juju logs are not written here.
func (fj fakejujuFilenames) logsFile() string {
	return filepath.Join(fj.logsdir, "fake-juju.log")
}

// jujudLogsFile() returns the path to the file where fake-juju writes
// the jujud logs.
func (fj fakejujuFilenames) jujudLogsFile() string {
	return filepath.Join(fj.logsdir, "jujud.log")
}

// fifoFile() returns the path to the FIFO file used by fake-juju.
// The FIFO is used by the fake-juju subcommands to interact with
// the daemon.
func (fj fakejujuFilenames) fifoFile() string {
	return filepath.Join(fj.datadir, "fifo")
}

// caCertFile() returns the path to the file holding the CA certificate
// used by the Juju API server.  fake-juju writes the cert there as a
// convenience for users.  It is not actually used for anything.
func (fj fakejujuFilenames) caCertFile() string {
	return filepath.Join(fj.datadir, "cert.ca")
}

// bootstrapResult encapsulates all significant information that came
// from bootstrapping an environment.
type bootstrapResult struct {
	dummyEnvName string
	cfgdir       string
	uuid         string
	username     string
	password     string
	addresses    []string
	caCert       []byte
}

// apiInfo() composes the Juju API info corresponding to the result.
func (br bootstrapResult) apiInfo() *api.Info {
	return &api.Info{
		Addrs:      br.addresses,
		Tag:        names.NewLocalUserTag(br.username),
		Password:   br.password,
		CACert:     string(br.caCert),
		EnvironTag: names.NewEnvironTag(br.uuid),
	}
}

// fakeJujuInfo() composes, from the result, the set of information
// that fake-juju should use internally.
func (br bootstrapResult) fakeJujuInfo() *processInfo {
	return &processInfo{
		Username:     br.username,
		Password:     br.password,
		WorkDir:      br.cfgdir,
		EndpointAddr: br.addresses[0],
		Uuid:         br.uuid,
		CACert:       br.caCert,
	}
}

// jenvSymlinkFilenames() determines the source and target paths for
// a symlink to the .jenv file for the identified environment.
func (br bootstrapResult) jenvSymlinkFilenames(jujuHome, envName string) (source, target string) {
	if jujuHome == "" || envName == "" {
		return "", ""
	}

	source = filepath.Join(br.cfgdir, "environments", br.dummyEnvName+".jenv")
	target = filepath.Join(jujuHome, "environments", envName+".jenv")
	return source, target
}

// apply() writes out the information from the bootstrap result to the
// various files identified by the provided filenames.
func (br bootstrapResult) apply(filenames fakejujuFilenames, envName string) error {
	if err := br.fakeJujuInfo().write(filenames.infoFile()); err != nil {
		return err
	}

	jenvSource, jenvTarget := br.jenvSymlinkFilenames(os.Getenv("JUJU_HOME"), envName)
	if jenvSource != "" && jenvTarget != "" {
		if err := os.MkdirAll(filepath.Dir(jenvTarget), 0755); err != nil {
			return err
		}
		if err := os.Symlink(jenvSource, jenvTarget); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(filenames.caCertFile(), br.caCert, 0644); err != nil {
		return err
	}

	return nil
}

// See github.com/juju/juju/blob/juju/testing/conn.go.
const dummyEnvName = "dummyenv"

func parseApiInfo(stdout io.ReadCloser) (*bootstrapResult, error) {
	buffer := bufio.NewReader(stdout)

	line, _, err := buffer.ReadLine()
	if err != nil {
		return nil, err
	}
	uuid := string(line)
	if !utils.IsValidUUIDString(uuid) {
		data, _ := ioutil.ReadAll(stdout)
		return nil, fmt.Errorf("%s\n%s", line, data)
	}

	line, _, err = buffer.ReadLine()
	if err != nil {
		return nil, err
	}
	workDir := string(line)

	result := &bootstrapResult{
		dummyEnvName: dummyEnvName,
		cfgdir:       workDir,
		uuid:         uuid,
	}

	store, err := configstore.NewDisk(workDir)
	if err != nil {
		return nil, err
	}
	info, err := store.ReadInfo(dummyEnvName)
	if err != nil {
		return nil, err
	}
	credentials := info.APICredentials()
	endpoint := info.APIEndpoint()
	result.username = credentials.User
	result.password = credentials.Password
	result.addresses = endpoint.Addresses
	result.caCert = []byte(endpoint.CACert)

	return result, nil
}

// Read the failures info file pointed by the FAKE_JUJU_FAILURES environment
// variable, if any. The format of the file is one entity name per line. If
// entity is found there, the code in FakeJujuSuite.TestStart will make that
// entity transition to an error state.
func readFailuresInfo() (map[string]bool, error) {
	log.Println("Checking for forced failures")
	failuresPath := os.Getenv("FAKE_JUJU_FAILURES")
	if failuresPath == "" {
		log.Println("No FAKE_JUJU_FAILURES env variable set")
	}
	log.Println("Reading failures file", failuresPath)
	failuresInfo := map[string]bool{}
	if _, err := os.Stat(failuresPath); os.IsNotExist(err) {
		log.Println("No failures file found")
		return failuresInfo, nil
	}
	file, err := os.Open(failuresPath)
	if err != nil {
		log.Println("Error opening failures file", err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var entity string
	for scanner.Scan() {
		entity = scanner.Text()
		log.Println("Add failure:", entity)
		failuresInfo[entity] = true
	}
	if err := scanner.Err(); err != nil {
		log.Println("Error reading failures file", err)
		return nil, err
	}
	return failuresInfo, nil
}

//===================================================================
// The fake-juju daemon (started by bootstrap) is found here.  It is
// implemented as a test suite.

type FakeJujuSuite struct {
	jujutesting.JujuConnSuite

	instanceCount     int
	machineStarted    map[string]bool
	filenames         fakejujuFilenames
	toCloseOnTearDown []io.Closer
}

var _ = gc.Suite(&FakeJujuSuite{})

func (s *FakeJujuSuite) SetUpTest(c *gc.C) {
	var CommandOutput = (*exec.Cmd).CombinedOutput
	s.JujuConnSuite.SetUpTest(c)

	c.Assert(os.Getenv(envDataDir), gc.Not(gc.Equals), "")
	s.filenames = newFakeJujuFilenames("", "", "")
	// Note that LoggingSuite.SetUpTest (github.com/juju/testing/log.go),
	// called via s.JujuConnSuite.SetUpTest(), calls loggo.ResetLogging().
	// So we cannot set up logging before then, since any writer we
	// register will get removed.  Consequently we lose any logs that get
	// generated in the SetUpTest() call.
	logFile, jujudLogFile := setUpLogging(c, s.filenames)
	s.toCloseOnTearDown = append(s.toCloseOnTearDown, logFile, jujudLogFile)

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
	defaultSeries := "trusty"
	if os.Getenv("DEFAULT_SERIES") != "" {
		defaultSeries = os.Getenv("DEFAULT_SERIES")
	}
	_, err = s.State.AddUser("admin", "Admin", password, "dummy-admin")
	c.Assert(err, gc.IsNil)
	_, err = s.State.AddEnvironmentUser(
		names.NewLocalUserTag("admin"), names.NewLocalUserTag("dummy-admin"), "Admin")
	c.Assert(err, gc.IsNil)
	err = s.State.UpdateEnvironConfig(
		map[string]interface{}{"default-series": defaultSeries}, nil, nil)
	c.Assert(err, gc.IsNil)

	// Create a machine to manage the environment.
	stateServer := s.Factory.MakeMachine(c, &factory.MachineParams{
		InstanceId: s.newInstanceId(),
		Nonce:      agent.BootstrapNonce,
		Jobs:       []state.MachineJob{state.JobManageEnviron, state.JobHostUnits},
		Series:     defaultSeries,
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
	jujuHome := osenv.JujuHome()

	binPath := filepath.Join(jujuHome, "bin")
	os.Mkdir(binPath, 0755)
	fakeSSHData := []byte("#!/bin/sh\nsleep 1\n")
	fakeSSHPath := filepath.Join(binPath, "ssh")
	err = ioutil.WriteFile(fakeSSHPath, fakeSSHData, 0755)
	c.Assert(err, gc.IsNil)
	os.Setenv("PATH", binPath+":"+os.Getenv("PATH"))

	// Once this FIFO is created, users can start sending commands.
	// The actual handling doesn't start until TestStart() runs.
	syscall.Mknod(s.filenames.fifoFile(), syscall.S_IFIFO|0666, 0)

	// Send the info back to the bootstrap command.
	reportInfo(apiInfo.EnvironTag.Id(), jujuHome)

	dpkgCmd := exec.Command(
		"dpkg-query", "--showformat='${Version}'", "--show", "fake-juju")
	out, err := CommandOutput(dpkgCmd)
	fakeJujuDebVersion := strings.Trim(string(out), "'")
	log.Printf("Started fake-juju-%s for %s\nJUJU_HOME=%s", fakeJujuDebVersion, version.Current, jujuHome)
}

func setUpLogging(c *gc.C, filenames fakejujuFilenames) (*os.File, *os.File) {
	c.Assert(filenames.logsdir, gc.Not(gc.Equals), "")

	// fake-juju logging
	logPath := filenames.logsFile()
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	c.Assert(err, gc.IsNil)
	log.SetOutput(logFile)

	// jujud logging
	logPath = filenames.jujudLogsFile()
	jujudLogFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	c.Assert(err, gc.IsNil)
	err = loggo.RegisterWriter("fake-juju-jujud-logs", loggo.NewSimpleWriter(jujudLogFile, &loggo.DefaultFormatter{}), loggo.TRACE)
	c.Assert(err, gc.IsNil)
	logger := loggo.GetLogger("fake-juju")
	logger.Infof("--- starting logging ---")

	return logFile, jujudLogFile
}

func reportInfo(uuid, jujuCfgDir string) {
	// IMPORTANT: don't remove this logging because it's used by the
	// bootstrap command.
	fmt.Println(uuid)
	fmt.Println(jujuCfgDir)
}

func (s *FakeJujuSuite) TearDownTest(c *gc.C) {
	log.Println("Tearing down processes")
	s.JujuConnSuite.TearDownTest(c)
	log.Println("Closing log files")
	for _, closer := range s.toCloseOnTearDown {
		closer.Close()
	}
}

func (s *FakeJujuSuite) TestStart(c *gc.C) {
	fifoPath := s.filenames.fifoFile()
	watcher := s.State.Watch()
	go func() {
		log.Println("Open commands FIFO", fifoPath)
		fd, err := os.Open(fifoPath)
		if err != nil {
			log.Println("Failed to open commands FIFO")
		}
		c.Assert(err, gc.IsNil)
		defer func() {
			if err := fd.Close(); err != nil {
				c.Logf("failed closing FIFO file: %s", err)
			}
			// Mark the controller as destroyed by renaming some files.
			if err := os.Rename(fifoPath, fifoPath+".destroyed"); err != nil {
				c.Logf("failed renaming FIFO file: %s", err)
			}
			infofile := s.filenames.infoFile()
			if err := os.Rename(infofile, infofile+".destroyed"); err != nil {
				c.Logf("failed renaming info file: %s", err)
			}
		}()
		scanner := bufio.NewScanner(fd)
		log.Println("Listen for commands on FIFO", fifoPath)
		scanner.Scan()
		log.Println("Stopping fake-juju")
		watcher.Stop()
	}()
	for {
		log.Println("Watching deltas")
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
	log.Println("Stopping fake-juju")
}

func (s *FakeJujuSuite) handleAddMachine(id string) error {
	machine, err := s.State.Machine(id)
	log.Println("Handle machine", id)
	if err != nil {
		return err
	}
	if instanceId, _ := machine.InstanceId(); instanceId == "" {
		err = machine.SetProvisioned(s.newInstanceId(), agent.BootstrapNonce, nil)
		if err != nil {
			log.Println("Got error with SetProvisioned", err)
			return err
		}
		address := network.NewScopedAddress("127.0.0.1", network.ScopeCloudLocal)
		err = machine.SetProviderAddresses(address)
		if err != nil {
			log.Println("Got error with SetProviderAddresses", err)
			return err
		}
	}
	status, _ := machine.Status()
	log.Println("Machine has status:", string(status.Status), status.Message)
	if status.Status == state.StatusPending {
		if err = s.startMachine(machine); err != nil {
			log.Println("Got error with startMachine:", err)
			return err
		}
	} else if status.Status == state.StatusStarted {
		log.Println("Starting units on machine", id)
		if _, ok := s.machineStarted[id]; !ok {
			s.machineStarted[id] = true
			if err = s.startUnits(machine); err != nil {
				log.Println("Got error with startUnits", err)
				return err
			}
		}
	}
	return nil
}

func (s *FakeJujuSuite) handleAddUnit(id string) error {
	unit, err := s.State.Unit(id)
	log.Println("Handle unit", id)
	if err != nil {
		log.Println("Got error with get unit", err)
		return err
	}
	machineId, err := unit.AssignedMachineId()
	if err != nil {
		return nil
	}
	log.Println("Got machineId", machineId)
	machine, err := s.State.Machine(machineId)
	if err != nil {
		log.Println("Got error with unit AssignedMachineId", err)
		return err
	}
	machineStatus, _ := machine.Status()
	if machineStatus.Status != state.StatusStarted {
		return nil
	}
	status, _ := unit.Status()
	log.Println("Unit has status", string(status.Status), status.Message)
	if status.Status != state.StatusActive && status.Status != state.StatusError {
		log.Println("Start unit", id)
		err = s.startUnit(unit)
		if err != nil {
			log.Println("Got error changing unit status", id, err)
			return err
		}
	} else if status.Status != state.StatusError {
		failuresInfo, err := readFailuresInfo()
		if err != nil {
			return err
		}
		if _, ok := failuresInfo["unit-"+id]; ok {
			agentStatus, err := unit.AgentStatus()
			if err != nil {
				log.Println("Got error checking agent status", id, err)
				return err
			}
			if agentStatus.Status != state.StatusError {
				log.Println("Error unit", id)
				err = s.errorUnit(unit)
				if err != nil {
					log.Println("Got error erroring unit status", id, err)
					return err
				}
			}
		}
	}
	return nil
}

func (s *FakeJujuSuite) startMachine(machine *state.Machine) error {
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

func (s *FakeJujuSuite) errorMachine(machine *state.Machine) error {
	time.Sleep(500 * time.Millisecond)
	err := machine.SetStatus(state.StatusError, "machine errored", nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *FakeJujuSuite) startUnits(machine *state.Machine) error {
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

func (s *FakeJujuSuite) startUnit(unit *state.Unit) error {
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

func (s *FakeJujuSuite) errorUnit(unit *state.Unit) error {
	log.Println("Erroring unit", unit.Name())
	err := unit.SetAgentStatus(state.StatusError, "unit errored", nil)
	if err != nil {
		return err
	}
	log.Println("Done eroring unit", unit.Name())
	return nil
}

func (s *FakeJujuSuite) newInstanceId() instance.Id {
	s.instanceCount += 1
	return instance.Id(fmt.Sprintf("id-%d", s.instanceCount))
}
