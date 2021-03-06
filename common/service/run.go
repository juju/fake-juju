// Run the fake-juju service

package service

import (
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/juju/loggo"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/provider/dummy"

	coretesting "github.com/juju/juju/testing"
	jujutesting "github.com/juju/testing"
)

// Main entry point for running the fake-juju service. It will:
//
// - Create FakeJujuSuite instance with suitable parameters (yes, a
//   gocheck test suite, see its docstring). Its role is to set up
//   and tear down a controller backed by the "dummy" provider (see
//   the github.com/juju/juju/provider/dummy package).
//
// - Start an HTTP server serving a "control plane" API for
//   controlling fake-juju itself.
//
// - When a 'bootstrap' request is received by the control plane API, kick
//   off a run of our FakeJujuSuite instance, which will in turn create a
//   new controller and start a juju API server for it.
//
// - When a 'destroy' request is received by the control plan API, terminate
//   the FakeJujuSuite, which will stop the API server and clear the database
//   state.
//
// Additional control plane API endpoints can be used to further control
// the fake-jujud service, for example by requesting some units or machines
// to simulate certain errors.
func RunFakeJuju() int {

	// Command line options
	flags := flag.NewFlagSet("fake-jujud", flag.ExitOnError)
	mongo := flags.Int("mongo", 0, "Optional external MongoDB port to use (default is to spawn a new instance on a random free port)")
	port := flags.Int("port", 17099, "The port the API server will listent to")
	series := flags.String("series", "xenial", "Ubuntu series")
	debug := flags.Bool("debug", false, "Enable debug logging")
	flags.Parse(os.Args[1:])

	level := loggo.INFO
	if *debug {
		level = loggo.DEBUG
	}
	options := &FakeJujuOptions{
		Output: os.Stdout,
		Series: *series,
		Mongo:  *mongo,
		Level:  level,
		Port:   *port,
	}

	runner := NewFakeJujuRunner(options)
	runner.Run()
	result := runner.Wait()

	if result.Succeeded == 1 {
		return 0
	} else {
		return 1
	}
}

func NewFakeJujuRunner(options *FakeJujuOptions) *FakeJujuRunner {
	return &FakeJujuRunner{
		options:  options,
		commands: make(chan *command, 1),
		result:   make(chan *gc.Result, 1),
	}
}

type FakeJujuRunner struct {
	options *FakeJujuOptions

	// Control channel for sending commands to the main loop, for example
	// the "bootstrap" command will trigger new iterations in the main
	// loop (i.e. a new "bootstrap" process).
	commands chan *command

	// Channel for signalling that the main loop has terminated
	result chan *gc.Result

	// Control plane API port listener
	listener net.Listener
}

// Perform some setup tasks (logging, mongo, control plane API) and
// then start the main loop in a goroutine. The main loop (or the
// setup phase, in case of problems) will signal termination via the
// FakeJujuRunner.result channel, which will be sent a *gc.Result with
// information about whether the service completed cleanly or not.
//
// Consumer code will then typically invoke FakeJujuRunner.Wait() to
// wait for the main loop to terminate and gather such exit result.
func (f *FakeJujuRunner) Run() {

	setupLogging(f.options.Output, f.options.Level)
	log.Infof("Starting service")

	// Set the certificates that the service will use. This option
	// will be false only in unit tests (where we'll use a
	// randomly generated test certificate).
	if !f.options.UseRandomCert {
		if err := coretesting.SetCerts(); err != nil {
			f.result <- &gc.Result{RunError: err}
			return
		}
	}

	if f.options.Mongo > 0 { // Use an external MongoDB instance
		log.Infof("Using external MongoDB on port %d", f.options.Mongo)

		jujutesting.SetExternalMgoServer(
			"localhost", f.options.Mongo, coretesting.Certs)
	} else if f.options.Mongo == 0 { // Start a dedicated MongoDB instance
		log.Infof("Starting dedicated MongoDB instance")

		// The github.com/juju/testing/mgo.go list of possible mongod paths
		// doesn't include the path to juju's custom mongod package, so we
		// we force it via this environment variable.
		os.Setenv("JUJU_MONGOD", "/usr/lib/juju/mongo3.2/bin/mongod")

		err := jujutesting.MgoServer.Start(coretesting.Certs)
		if err != nil {
			f.result <- &gc.Result{RunError: err}
			return
		}
	}

	// Configure the test API server to listen to this port (the server
	// will be started only at "bootstrap" time, see TestMainLoop()).
	dummy.SetAPIPort(f.options.Port)

	// Start the control-plane API
	if err := f.serveControlPlaneAPI(); err != nil {
		f.result <- &gc.Result{RunError: err}
		return
	}

	// Start the main loop, waiting for 'bootstrap' commands
	conf := &gc.RunConf{
		Output: os.Stdout,
		Filter: "TestMainLoop",
	}

	go func() {
		result := gc.Run(f, conf)

		if f.options.Mongo == 0 {
			// Shutdown our dedicated MongoDB instance child
			// process.
			log.Infof("Stopping dedicated MongoDB instance")
			jujutesting.MgoServer.Destroy()
		}

		if f.listener != nil {
			f.stopControlPlaneAPI()
		}

		f.result <- result
	}()
}

// Main control loop of the fake-jujud service. It's prefixed with "Test"
// only because we need to convince the gocheck package that this is
// a test method, and thus have it invoked with a reference to a gc.C
// instance (that we'll use to setup/teardown our FakeJujuSuite instance).
func (f *FakeJujuRunner) TestMainLoop(c *gc.C) {

	log.Infof("Starting main loop")

	terminate := make(chan os.Signal, 2)
	signal.Notify(terminate, os.Interrupt, syscall.SIGTERM)

	suite := &FakeJujuSuite{
		options: f.options,
	}
	suite.SetUpSuite(c)

	// Process commands, typically coming from the control plane API.
	for {
		stop := false

		var err error
		var command *command

		select {
		case <-terminate:
		case command = <-f.commands:
		}
		if command == nil {
			// This means we either received a signal, let's
			// exit.
			f.Stop()
			continue
		}

		if command.code == commandCodeStop {
			log.Infof("Terminating service")
			stop = true
		} else if command.code == commandCodeBootstrap {
			log.Infof("Bootstrapping fake controller")
			suite.SetUpTest(c)
			go f.monitorWatchLoop(suite)
		} else if command.code == commandCodeDestroy {
			log.Infof("Destroying fake controller")
			suite.TearDownTest(c)
		}
		command.done <- err

		if stop {
			break
		}
	}

	if isControllerBootstrapped(suite) {
		// This means that we didn't have chance to tear down the test
		// because either the destroy command was not invoked or we
		// got interrupted by a signal. Let's force a clean up.
		suite.TearDownTest(c)
	}

	suite.TearDownSuite(c)
}

// Stop the main loop.
func (f *FakeJujuRunner) Stop() {
	f.commands <- newCommand(commandCodeStop)
}

// Wait for the main loop to complete and return the result.
func (f *FakeJujuRunner) Wait() *gc.Result {
	result := <-f.result
	logResult(result)
	return result
}

// Monitor the watch loop of FakeJujuService and catch unexpected
// errors.  If anything bad happens, we'll bail out. Otherwise, this
// goroutine will silently terminate when the delta watch loop in
// FakeJujuService gets stopped.
func (f *FakeJujuRunner) monitorWatchLoop(suite *FakeJujuSuite) {

	// Here we block until the watch loop terminates, either
	// successfully or not.
	err := suite.Wait()

	if err != nil {
		log.Errorf("Watch loop error: %s", err.Error())
		f.Stop()
	} else {
		log.Infof("Stop monitoring watch loop")
	}
}

// Log a summary of the service run
func logResult(result *gc.Result) {
	if !(result.Succeeded == 1) {
		log.Infof("Service finished uncleanly: %s", result.String())
	} else {
		log.Infof("Service finished cleanly")
	}
}

// Whether the dummy controller created by FakeJujuSuite has been bootstrapped.
func isControllerBootstrapped(suite *FakeJujuSuite) bool {
	// Use the presence of the RootDir has a flag indicating that the
	// suite setup has run.
	return suite.RootDir != ""
}
