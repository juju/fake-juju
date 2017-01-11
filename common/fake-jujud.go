// Main fake-juju process, acting as Juju controller.

package main

import (
	"os"
	"./service"
)

func main() {
	os.Exit(service.RunFakeJuju())
}
