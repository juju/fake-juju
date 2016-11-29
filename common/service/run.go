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
// a FakeJujuService instance (which is actually a gocheck test suite)
// and run it (i.e. invoke its single TestStart test method, which will
// spin the "service" indefinitely).
func RunFakeJuju() {

	series := "xenial"
	if os.Getenv("DEFAULT_SERIES") != "" {
		series = os.Getenv("DEFAULT_SERIES")
	}

	options := &FakeJujuOptions{
		output: os.Stdout,
		level: loggo.INFO,
		series: series,
		mongo:  true,
	}
	service := &FakeJujuService{
		options: options,
	}
	runner := &FakeJujuRunner{
		service: gc.Suite(service),
		options: options,
	}
	result := runner.Run()

	if !(result.Succeeded == 1) {
		log.Infof("Service exited uncleanly: %s", result.RunError.Error())
	}
	log.Infof("Stopping service")
}

type FakeJujuRunner struct {

	// A FakeJujuService instance
	service interface{}
	options *FakeJujuOptions
}

func (f *FakeJujuRunner) Run() *gc.Result {

	setupLogging(f.options.output, f.options.level)
	log.Infof("Starting service")

	// Conditionally start mongo (we don't want this for unit tests).
	if f.options.mongo {
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
	return gc.Run(f.service, conf)
}
