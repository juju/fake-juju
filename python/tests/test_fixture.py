import os

import yaml

from subprocess import check_call

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
        The command line performs a fake bootstrap and populates JUJU_DATA.
        """
        juju_data = self.useFixture(TempDir())
        self.useFixture(EnvironmentVariable("JUJU_DATA", juju_data.path))
        check_call([FAKE_JUJU, "bootstrap", "foo", "bar"])

        with open(juju_data.join("controllers.yaml")) as fd:
            controllers = yaml.load(fd)
            self.assertEqual(
                CONTROLLER_UUID, controllers["controllers"]["bar"]["uuid"])

        with open(juju_data.join("accounts.yaml")) as fd:
            accounts = yaml.load(fd)
            self.assertEqual(
                {"bar": {"password": "dummy-secret", "user": "admin"}},
                accounts["controllers"])

        with open(juju_data.join("models.yaml")) as fd:
            models = yaml.load(fd)
            self.assertEqual(
                {"admin/controller": {"uuid": MODEL_UUID}},
                models["controllers"]["bar"]["models"])
            self.assertEqual(
                "admin/controller",
                models["controllers"]["bar"]["current-model"])
