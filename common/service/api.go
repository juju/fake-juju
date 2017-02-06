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
	mux.Post("/bootstrap", http.HandlerFunc(f.bootstrap))
	mux.Post("/destroy", http.HandlerFunc(f.destroy))
	mux.Post("/fail/:entity", http.HandlerFunc(f.fail))

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

// Bootstrap the controller by starting machine 0
func (f *FakeJujuRunner) bootstrap(w http.ResponseWriter, req *http.Request) {
	command := newCommand(commandCodeBootstrap)
	f.commands <- command
	writeResponse(w, <-command.done)
}

// Destroy the controller by tearing down the FakeJujuSuite, which under the
// hood will stop the test juju API server and reset the database.
func (f *FakeJujuRunner) destroy(w http.ResponseWriter, req *http.Request) {
	command := newCommand(commandCodeDestroy)
	f.commands <- command
	writeResponse(w, <-command.done)
}

// Mark the given entity as doomed to fail
func (f *FakeJujuRunner) fail(w http.ResponseWriter, req *http.Request) {
	SetFailure(req.URL.Query().Get(":entity"))
}

// Write the response, in case of error the message is provided in the body.
func writeResponse(w http.ResponseWriter, err error) {
	var body string
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		body = err.Error()
	} else {
		body = "done"
	}
	w.Write([]byte(fmt.Sprintf("%s\n", body)))
}
