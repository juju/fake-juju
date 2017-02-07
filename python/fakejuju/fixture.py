"""Fixtures for running long-lived fake-juju processes during a test run."""

import os
import ssl
import requests

from subprocess import (
    check_output,
    STDOUT,
)

from fixtures import (
    Fixture,
    TempDir,
)

from txfixtures import Service
from txfixtures.mongodb import MongoDB

# Default fake-juju version that code consuming the FakeJuju fixture will use.
# It can be tweaked by using a JUJU_VERSION environment variable.
DEFAULT_VERSION = os.environ.get("JUJU_VERSION", "2.0.2")

# Directory where the test certificate to use lives.
DEFAULT_CERT_DIR = "/usr/share/fake-juju/cert"

# Path to the root of the local fake-juju Git source tree
ROOT = os.path.dirname(
    os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

JUJU_MONGOD = "/usr/lib/juju/mongo3.2/bin/mongod"
if not os.path.exists(JUJU_MONGOD):  # Trusty has no mongo 3.2
    JUJU_MONGOD = "/usr/lib/juju/bin/mongod"

JUJU_MONGOD_ARGS = (  # See github.com/juju/testing/mgo.go
    "--nssize=1",
    "--noprealloc",
    "--smallfiles",
    "--nohttpinterface",
    "--oplogSize=10",
    "--ipv6",
    "--setParameter=enableTestCommands=1",
    "--nounixsocket",
    "--sslOnNormalPorts",
    "--sslPEMKeyPassword=ignored",
)

FAKE_JUJUD = "fake-jujud"

# The test model has a stable UUID. See:
# github.com/juju/juju/testing/environ.go
MODEL_UUID = "deadbeef-0bad-400d-8000-4b1d0d06f00d"

USER = "user-admin"
PASSWORD = "dummy-secret"


class JujuMongoDB(MongoDB):
    """
    Spawn a juju-mongodb server, suitable for acting as fake-juju backend.
    """

    def __init__(self, reactor, timeout=None):
        # We honor the JUJU_MONGOD environment variable, the same one
        # supported by github.com/juju/testing/mgo.go, and set by the
        # fake-juju Travis job (see .travis.yml).
        command = os.environ.get("JUJU_MONGOD", JUJU_MONGOD)
        serverPem = os.path.join(_get_cert_dir(), "server.pem")
        args = list(JUJU_MONGOD_ARGS) + [
            "--sslPEMKeyFile={serverPem}".format(serverPem=serverPem)]
        super(JujuMongoDB, self).__init__(
            reactor, command=command, args=args, timeout=timeout)
        self.setClientKwargs(ssl=True, ssl_cert_reqs=ssl.CERT_NONE)


class FakeJuju(Service):
    """Spawn a fake-juju process, pointing to the given mongodb port."""

    def __init__(self, reactor, version=DEFAULT_VERSION, mongo=None, env=None,
                 timeout=None):
        """
        :param version: The version number of the specific fake-juju to use.
        :param mongo: The JujuMongoDB fixture instance managing the mongod
            process to point this fake-juju to. If None, a brand new one
            will be created.
        """
        self.version = version
        command = "fake-jujud-{}".format(self.version)
        super(FakeJuju, self).__init__(
            reactor, command, timeout=timeout, env=env)
        self.mongo = mongo
        self.expectPort(17100)
        self.expectOutput("Starting main loop")

    def _setUp(self):
        if self.mongo is None:
            self.mongo = self.useFixture(
                JujuMongoDB(self.reactor, timeout=self.protocol.timeout))
        super(FakeJuju, self)._setUp()

    @property
    def address(self):
        return "localhost:%d" % (self.port - 1)

    @property
    def port(self):
        return self.protocol.expectedPort

    def cli(self):
        """Return a FakeJujuCLI fixture matching this fake-juju version."""
        return self.useFixture(FakeJujuCLI(self.version, env=self.env))

    def fail(self, entity):
        """Mark the given entity as failing.

        It will be transitioned to the error state as soon as it gets created.

        Currently only failing units is supported.

        :param entity: A string of the form "<kind>-<id>", for example
            "unit-postgresql-1".
        """
        requests.post("http://localhost:{}/fail/{}".format(self.port, entity))

    def _extraArgs(self):
        return [
            "-mongo", str(self.mongo.port),
            "-port", str(self.port - 1)]


class FakeJujuCLI(Fixture):
    """A convenience fixture for executing fake-juju CLI commands."""

    def __init__(self, version, env=None):
        """
        :param version: The version number of the specific fake-juju to use.
        :param env: Optionally, the environment variables to use.
        """
        self.command = "fake-juju-{}".format(version)
        self.env = env or os.environ

    def _setUp(self):
        self.data_dir = self.useFixture(TempDir())

    def execute(self, *args):
        """Execute the fake-juju command line with the given argument."""
        env = self.env.copy()
        env.update({"JUJU_DATA": self.data_dir.path})
        return check_output((self.command,) + args, env=env, stderr=STDOUT)


def _get_cert_dir():
    """Return the path to the test certificate directory to use.

    This function will detect if we're being invoked inside fake-juju source
    checkout and, if so, return the path the local cert/ directory. Otherwise,
    it will return the default system-wide test certificate directory installed
    by the fake-juju Debian package.
    """
    cert_dir = os.path.join(ROOT, "cert")
    if not os.path.exists(cert_dir):
        cert_dir = DEFAULT_CERT_DIR
    return cert_dir
