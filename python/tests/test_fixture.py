import os
import json

import yaml

from subprocess import (
    check_output,
    check_call,
    STDOUT,
    PIPE,
    CalledProcessError,
)

from testtools import TestCase

from fixtures import (
    FakeLogger,
    EnvironmentVariable,
    TempDir
)

from txfixtures import Reactor

from fakejuju.fixture import (
    ROOT,
    JujuMongoDB,
    FakeJuju,
)

JUJU_VERSION = "2.0.2"

FAKE_JUJUD = os.path.join(ROOT, JUJU_VERSION, "fake-jujud")
FAKE_JUJU = os.path.join(ROOT, JUJU_VERSION, "fake-juju")

# Constants from github.com/juju/juju/testing/environ.go
CONTROLLER_UUID = "deadbeef-1bad-500d-9000-4b1d0d06f00d"
MODEL_UUID = "deadbeef-0bad-400d-8000-4b1d0d06f00d"


class JujuMongoDBIntegrationTest(TestCase):

    def setUp(self):
        super(JujuMongoDBIntegrationTest, self).setUp()
        self.logger = self.useFixture(FakeLogger())
        self.useFixture(Reactor())
        self.mongodb = self.useFixture(JujuMongoDB())

    def test_client(self):
        """The mongo instance is up and we can connect fine with the client."""
        self.assertIsNotNone(self.mongodb.client.server_info())


class FakeJujuIntegrationTest(TestCase):

    def setUp(self):
        super(FakeJujuIntegrationTest, self).setUp()
        self.logger = self.useFixture(FakeLogger())
        self.useFixture(Reactor())
        self.mongodb = self.useFixture(JujuMongoDB())
        self.fake_juju = self.useFixture(FakeJuju(
            self.mongodb.port, binary=FAKE_JUJUD))

    def test_up(self):
        """The fake-juju service is connected to the mongodb one."""
        self.assertIn(
            "Using external MongoDB on port %d" % self.mongodb.port,
            self.logger.output)

    def test_cli_bootstrap(self):
        """
        The fake-juju command line tool can perform a fake bootstrap. After the
        bootstrap the controller machine (machine 0) is present.
        """
        juju_data = self.useFixture(TempDir())
        self.useFixture(EnvironmentVariable("JUJU_DATA", juju_data.path))
        check_call([FAKE_JUJU, "bootstrap", "foo", "bar"])
        check_call([FAKE_JUJU, "switch", "bar"], stdout=PIPE, stderr=STDOUT)

        output = check_output([FAKE_JUJU, "status", "--format=json"])
        status = json.loads(output)

        self.assertEqual(
            "running", status["machines"]["0"]["machine-status"]["current"])

        self.assertEqual(
            "started", status["machines"]["0"]["juju-status"]["current"])

    def test_cli_destroy_controller(self):
        """
        The fake-juju command line tool can destroy the fake controller, using
        the normal "destroy-controller" subcommand.
        """
        juju_data = self.useFixture(TempDir())
        self.useFixture(EnvironmentVariable("JUJU_DATA", juju_data.path))
        check_call([FAKE_JUJU, "bootstrap", "foo", "bar"])
        check_call([FAKE_JUJU, "switch", "bar"], stdout=PIPE, stderr=STDOUT)
        check_call([FAKE_JUJU, "destroy-controller", "-y", "bar"])
        self.assertRaises(
            CalledProcessError,
            check_call, [FAKE_JUJU, "status"], stdout=PIPE, stderr=STDOUT)
