// Handle test certificates used by fake-juju

package service

import (
	"io/ioutil"
	"path/filepath"

	gitjujutesting "github.com/juju/testing"

	"github.com/juju/juju/cert"
	"github.com/juju/juju/testing"
)

var (
	filenames = []string{
		"ca.cert",
		"ca.key",
		"server.cert",
		"server.key",
	}

	certificates = []*string{
		&testing.CACert,
		&testing.CAKey,
		&testing.ServerCert,
		&testing.ServerKey,
	}
)

// Set the certificate global variables in the github.com/juju/juju/testing
// package, to make the test suite use a custom certificate (typically the one
// in the cert/ directory of the fake-juju tree) instead of the auto-generated
// one that would otherwise be set. This allows us to point to an external
// MongoDB process, spawned using the custom certificate.
func SetCerts(path string) error {

	log.Infof("Loading certificates from %s", path)

	// Read the certificate bytes, overwriting the relevant variables
	// from github.com/juju/juju/testing.
	for i, filename := range filenames {
		data, err := ioutil.ReadFile(filepath.Join(path, filename))
		if err != nil {
			return err
		}
		*certificates[i] = string(data)
	}

	// Parse the bytes converting them into Go objects
	caCertX509, _, err := cert.ParseCertAndKey(
		testing.CACert, testing.CAKey)
	if err != nil {
		return err
	}
	serverCert, serverKey, err := cert.ParseCertAndKey(
		testing.ServerCert, testing.ServerKey)

	if err != nil {
		return err
	}

	// Set the global Certs object in the Juju testing package
	testing.Certs = &gitjujutesting.Certs{
		CACert:     caCertX509,
		ServerCert: serverCert,
		ServerKey:  serverKey,
	}

	return nil
}
