// Run the fake-juju service

package service

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/juju/loggo"
	gc "gopkg.in/check.v1"

	coretesting "github.com/juju/juju/testing"
	jujutesting "github.com/juju/testing"
)

// Main entry point for running the fake-juju service. It will create
// a FakeJujuSuite instance (yes, a gocheck test suite, see its docstring)
// and run it (i.e. invoke its single TestStart test method, which will
// spin a FakeJujuService indefinitely).
func RunFakeJuju() int {

	// Command line options
	flags := flag.NewFlagSet("fake-jujud", flag.ExitOnError)
	mongo := flags.Int("mongo", 0, "Optional external MongoDB port to use (default is to spawn a new instance on a random free port)")
	cert := flags.String("cert", "/usr/share/fake-juju/cert", "Certificate directory")
	series := flags.String("series", "xenial", "Ubuntu series")
	flags.Parse(os.Args[1:])

	options := &FakeJujuOptions{
		Output: os.Stdout,
		Series: *series,
		Mongo:  *mongo,
		Cert:   *cert,
		Level:  loggo.INFO,
	}
	suite := &FakeJujuSuite{
		options: options,
	}

	runner := NewFakeJujuRunner(suite, options)
	runner.Run()
	result := runner.Wait()

	if result.Succeeded == 1 {
		return 0
	} else {
		return 1
	}
}

func NewFakeJujuRunner(suite *FakeJujuSuite, options *FakeJujuOptions) *FakeJujuRunner {
	return &FakeJujuRunner{
		suite:    suite,
		options:  options,
		commands: make(chan *command, 1),
		result:   make(chan *gc.Result, 1),
	}
}

type FakeJujuRunner struct {
	suite   *FakeJujuSuite
	options *FakeJujuOptions

	// Control channel for sending commands to the main loop, for example
	// the "bootstrap" command will trigger new iterations in the main
	// loop (i.e. a new "bootstrap" process).
	commands chan *command

	// Channel for signalling that the main loop has terminated
	result chan *gc.Result
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

	if f.options.Mongo > 0 { // Use an external MongoDB instance
		log.Infof("Using external MongoDB on port %d", f.options.Mongo)

		// Set the certificates that the service will use
		err := SetCerts(f.options.Cert)
		if err != nil {
			f.result <- &gc.Result{RunError: err}
			return
		}
		jujutesting.SetExternalMgoServer(
			"localhost", f.options.Mongo, coretesting.Certs)
	} else if f.options.Mongo == 0 { // Start a dedicated MongoDB instance
		log.Infof("Starting dedicated MongoDB instance")

		err := jujutesting.MgoServer.Start(coretesting.Certs)
		if err != nil {
			f.result <- &gc.Result{RunError: err}
			return
		}
		defer jujutesting.MgoServer.Destroy()
	}

	conf := &gc.RunConf{
		Output: ioutil.Discard, // We don't want any output from gocheck
		Filter: "TestMainLoop",
	}

	go func() {
		result := gc.Run(f, conf)
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
		}
		command.done <- err

		if stop {
			break
		}
	}
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

// Log a summary of the service run
func logResult(result *gc.Result) {

	if !(result.Succeeded == 1) {
		message := "Unknown error"
		if result.RunError != nil {
			message = result.RunError.Error()
		}
		log.Infof("Service finished uncleanly: %s", message)
	} else {
		log.Infof("Service finished cleanly")
	}
}
