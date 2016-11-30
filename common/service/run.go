// Run the fake-juju service

package service

import (
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
func RunFakeJuju() {

	series := "xenial"
	if os.Getenv("DEFAULT_SERIES") != "" {
		series = os.Getenv("DEFAULT_SERIES")
	}

	options := &FakeJujuOptions{
		Output: os.Stdout,
		Level:  loggo.INFO,
		Series: series,
		Mongo:  true,
	}
	suite := &FakeJujuSuite{
		options: options,
	}
	runner := NewFakeJujuRunner(suite, options)
	runner.Run()
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

	// Conditionally start mongo (we don't want this for unit tests).
	if f.options.Mongo {
		certs := coretesting.Certs
		if err := jujutesting.MgoServer.Start(certs); err != nil {
			return &gc.Result{RunError: err}
		}
		defer jujutesting.MgoServer.Destroy()
	}

	conf := &gc.RunConf{
		Output: ioutil.Discard, // We don't want any output from gocheck
		Filter: "TestStart",
	}
	result := gc.Run(f.suite, conf)

	if !(result.Succeeded == 1) {
		message := "No error message"
		if result.RunError != nil {
			message = result.RunError.Error()
		}
		log.Infof("Service exited uncleanly: %s", message)
	}
	log.Infof("Stopping service")

	return result

}
