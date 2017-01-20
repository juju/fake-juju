// Simple HTTP client for the fake-jujud control-plane API

package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/juju/errors"
)

type FakeJujuClient struct {

	// The fake-jujud control plane API port on localhost to connect to
	port int
}

// Get a new client using the default port or the one from the FAKE_JUJUD_PORT
// environment variable.
func NewFakeJujuClient() (*FakeJujuClient, error) {
	port, err := GetFakeJujudPort()
	if err != nil {
		return nil, err
	}
	return NewFakeJujuClientWithPort(port), nil
}

// Get a new client pointing to the given port.
func NewFakeJujuClientWithPort(port int) *FakeJujuClient {
	return &FakeJujuClient{port: port}
}

// Perform a new controller bootstrap
func (c *FakeJujuClient) Bootstrap() error {
	return c.post("bootstrap")
}

// Destroy the running controller
func (c *FakeJujuClient) Destroy() error {
	return c.post("destroy")
}

// Perform a POST HTTP request against the fake-juju control API
func (c *FakeJujuClient) post(path string) error {
	url := fmt.Sprintf("http://127.0.0.1:%d/%s", c.port, path)

	logger.Debugf("Performing fake-juju request at %s", url)
	response, err := http.Post(url, "", bytes.NewBuffer([]byte("")))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	logger.Debugf("Got response Status: %s", response.Status)

	if response.StatusCode != 200 {
		data, err := ioutil.ReadAll(response.Body)
		message := string(data)
		if err != nil {
			message = err.Error()
		}
		return errors.New(
			fmt.Sprintf("Failed fake-juju request: %s", message))
	}

	return nil
}

// Figure the port that fake-jujud is listening to.
func GetFakeJujudPort() (port int, err error) {
	port = 17100 // the default
	if os.Getenv("FAKE_JUJUD_PORT") != "" {
		port, err = strconv.Atoi(os.Getenv("FAKE_JUJUD_PORT"))
		if err != nil {
			return 0, errors.Annotate(err, "invalid port number")
		}
	}
	return port, nil
}
