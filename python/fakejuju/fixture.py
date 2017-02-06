"""Fixtures for running long-lived fake-juju processes during a test run."""

import os
import ssl

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

# Directory where the test certificate to use lives.
DEFAULT_CERT_DIR = "/usr/share/fake-juju/cert"

# Path to the root of the local fake-juju Git source tree
ROOT = os.path.dirname(
    os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

JUJU_MONGOD = "/usr/lib/juju/mongo3.2/bin/mongod"
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

    def __init__(self):
        serverPem = os.path.join(_get_cert_dir(), "server.pem")
        args = JUJU_MONGOD_ARGS + (
            "--sslPEMKeyFile={serverPem}".format(serverPem=serverPem),)
        super(JujuMongoDB, self).__init__(
            # We honor the JUJU_MONGOD environment variable, the same one
            # supported by github.com/juju/testing/mgo.go, and set by the
            # fake-juju Travis job (see .travis.yml).
            mongod=os.environ.get("JUJU_MONGOD", JUJU_MONGOD),
            args=args,
        )
        self.setClientKwargs(ssl=True, ssl_cert_reqs=ssl.CERT_NONE)


class FakeJuju(Service):
    """Spawn a fake-juju process, pointing to the given mongodb port."""

    def __init__(self, version="2.0.2", mongo=None, port=17100, **kwargs):
        """
        :param version: The version number of the specific fake-juju to use.
        :param mongo: The JujuMongoDB fixture instance managing the mongod
            process to point this fake-juju to. If None, a brand new one
            will be created.
        :param port: Port that the fake juju control plane API will listen to.
            The juju API server bootstrapped by fake-jujud itself will listen
            to "port -1 ".
        """
        self.version = version
        command = ["fake-jujud-{}".format(self.version)]
        super(FakeJuju, self).__init__(command, **kwargs)
        self.mongo = mongo
        self.expectOutput("Starting main loop")
        self.expectPort(port)  # This is the control-plane API port

    def _setUp(self):
        if self.mongo is None:
            self.mongo = self.useFixture(JujuMongoDB())
        super(FakeJuju, self)._setUp()

    @property
    def port(self):
        return self.protocol.expectedPort

    def cli(self):
        """Return a FakeJujuCLI fixture matching this fake-juju version."""
        return self.useFixture(FakeJujuCLI(self.version, env=self.env))

    @property
    def _args(self):
        return self.command + [
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
