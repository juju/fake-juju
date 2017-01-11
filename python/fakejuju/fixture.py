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
    """
    Spawn a fake-juju process, pointing the given mongodb port.
    """

    def __init__(self, mongo_port, fake_jujud=FAKE_JUJUD, **kwargs):
        command = [fake_jujud, "-mongo", str(mongo_port), "-cert", CERT]
        super(FakeJuju, self).__init__(command, **kwargs)
        self.expectOutput("preferred public address changed from")
