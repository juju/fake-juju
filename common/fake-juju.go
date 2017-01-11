// Command line for fake-juju, mimicking a real juju command line but always
// using fake-juju and the dummy provider as backend (regardless of user's
// configuration).

package main

import (
	"os"

	"github.com/juju/loggo"

	"github.com/juju/juju/cmd/juju/commands"
	components "github.com/juju/juju/component/all"
	// Import the providers.
	_ "github.com/juju/juju/provider/all"
)

var log = loggo.GetLogger("juju.cmd.juju")

func init() {
	if err := components.RegisterForClient(); err != nil {
		log.Criticalf("unable to register client components: %v", err)
		os.Exit(1)
	}
}

func main() {
	os.Exit(commands.Main(os.Args))
}
