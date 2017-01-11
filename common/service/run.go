// Run the fake-juju service

package service

import (
	"flag"
	"io/ioutil"
	"os"

	gc "gopkg.in/check.v1"

	"github.com/juju/loggo"

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
	result := runner.Run()

	if result.Succeeded == 1 {
		return 0
	} else {
		return 1
	}
}

func NewFakeJujuRunner(suite interface{}, options *FakeJujuOptions) *FakeJujuRunner {
	return &FakeJujuRunner{
		suite:   suite,
		options: options,
	}
}

type FakeJujuRunner struct {
	suite   interface{} // A FakeJujuSuite instance
	options *FakeJujuOptions
}

func (f *FakeJujuRunner) Run() *gc.Result {

	setupLogging(f.options.Output, f.options.Level)
	log.Infof("Starting service")

	if f.options.Mongo > 0 { // Use an external MongoDB instance
		log.Infof("Using external MongoDB on port %d", f.options.Mongo)

		// Set the certificates that the service will use
		err := SetCerts(f.options.Cert)
		if err != nil {
			result := &gc.Result{RunError: err}
			logResult(result)
			return result
		}
		jujutesting.SetExternalMgoServer(
			"localhost", f.options.Mongo, coretesting.Certs)
	} else if f.options.Mongo == 0 { // Start a dedicated MongoDB instance
		log.Infof("Starting dedicated MongoDB instance")

		err := jujutesting.MgoServer.Start(coretesting.Certs)
		if err != nil {
			return &gc.Result{RunError: err}
		}
		defer jujutesting.MgoServer.Destroy()
	}

	conf := &gc.RunConf{
		Output: ioutil.Discard, // We don't want any output from gocheck
		Filter: "TestStart",
	}
	result := gc.Run(f.suite, conf)

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
