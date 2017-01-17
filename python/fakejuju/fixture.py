"""Fixtures for running long-lived fake-juju processes during a test run."""

import os
import ssl

from txfixtures import Service
from txfixtures.mongodb import MongoDB

ROOT = os.path.dirname(
    os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
CERT = os.path.join(ROOT, "cert")
SERVER_PEM = os.path.join(CERT, "server.pem")

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
    "--sslPEMKeyFile={serverPem}".format(serverPem=SERVER_PEM),
    "--sslPEMKeyPassword=ignored",
)

FAKE_JUJUD = "fake-jujud"


class JujuMongoDB(MongoDB):
    """
    Spawn a juju-mongodb server, suitable for acting as fake-juju backend.
    """

    def __init__(self):
        super(JujuMongoDB, self).__init__(
            # We honor the JUJU_MONGOD environment variable, the same one
            # supported by github.com/juju/testing/mgo.go, and set by the
            # fake-juju Travis job (see .travis.yml).
            mongod=os.environ.get("JUJU_MONGOD", JUJU_MONGOD),
            args=JUJU_MONGOD_ARGS,
        )
        self.setClientKwargs(ssl=True, ssl_cert_reqs=ssl.CERT_NONE)


class FakeJuju(Service):
    """Spawn a fake-juju process, pointing to the given mongodb port."""

    def __init__(self, mongo_port, binary=FAKE_JUJUD, port=17099, **kwargs):
        """
        :param mongo_port: Port of the MongoDB instance that this fake-jujud
            process should use.
        :param binary: Path to the fake-jujud binary to spawn.
        :param port: Port that the fake juju API server will listen to.
        """
        command = [
            binary, "-mongo", str(mongo_port), "-cert", CERT,
            "-port", str(port)]
        super(FakeJuju, self).__init__(command, **kwargs)
        self.expectOutput("Starting main loop")
        self.expectPort(port + 1)  # This is the control-plan API port

    @property
    def port(self):
        return self.protocol.expectedPort
