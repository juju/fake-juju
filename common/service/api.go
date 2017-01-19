// Control plane HTTP API for controlling fake-juju itself.

package service

import (
	"fmt"
	"net"
	"net/http"

	"github.com/bmizerany/pat"
)

// Start an HTTP server in a goroutine, exposing the control plane API.
func (f *FakeJujuRunner) serveControlPlaneAPI() error {

	mux := pat.New()

	// We want to use a port different than the one used for the
	// juju API server. Incrementing by one will do the trick and
	// also make the control-plan port predictable (if the API
	// server one is known)
	port := f.options.Port + 1

	var err error
	f.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	server := &http.Server{Handler: mux}

	go func() {
		log.Infof("Starting control plane API on port %d", port)
		server.Serve(f.listener)
	}()

	return nil
}
 
// Stop the control plane HTTP server.
func (f *FakeJujuRunner) stopControlPlaneAPI() {
	addr := f.listener.Addr()
	log.Infof("Stopping control plane API on address %s", addr.String())
	f.listener.Close()
}
