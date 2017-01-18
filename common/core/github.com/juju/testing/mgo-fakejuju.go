package testing

import (
	"fmt"
)

// Support pointing MgoSuite to an external MongoDB server,
// instead of spawning a new one as child process.
func SetExternalMgoServer(addr string, port int, certs *Certs) error {
	MgoServer.addr = fmt.Sprintf("%s:%d", addr, port)
	MgoServer.port = port
	MgoServer.certs = certs
	return nil
}
