// Handle test certificates used by fake-juju

package testing

import (
	"io/ioutil"
	"path/filepath"

	"github.com/juju/loggo"

	gitjujutesting "github.com/juju/testing"

	"github.com/juju/juju/cert"
)

var (
	filenames = []string{
		"ca.cert",
		"ca.key",
		"server.cert",
		"server.key",
	}

	certificates = []*string{
		&CACert,
		&CAKey,
		&ServerCert,
		&ServerKey,
	}
)

// Set the certificate global variables in the github.com/juju/juju/testing
// package, to make the test suite use a custom certificate (typically the one
// in the cert/ directory of the fake-juju tree) instead of the auto-generated
// one that would otherwise be set. This allows us to point to an external
// MongoDB process, spawned using the custom certificate.
func SetCerts(path string) error {

	loggo.GetLogger("").Debugf("Loading certificates from %s", path)

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
	caCertX509, _, err := cert.ParseCertAndKey(CACert, CAKey)
	if err != nil {
		return err
	}
	serverCert, serverKey, err := cert.ParseCertAndKey(ServerCert, ServerKey)
	if err != nil {
		return err
	}

	// Set the global Certs object in the Juju testing package
	Certs = &gitjujutesting.Certs{
		CACert:     caCertX509,
		ServerCert: serverCert,
		ServerKey:  serverKey,
	}

	return nil
}
