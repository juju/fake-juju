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

# TODO: put in place some logic to automatically figure out what the
#       latest version is, instead of hard-coding it.
JUJU_VERSION = "2.0.2"


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

        # The JUJU_VERSION variable is a hook that can be used to force tests
        # run against a specific version.
        version = os.environ.get("JUJU_VERSION", JUJU_VERSION)

        # This will force the FakeJuju fixture to use the locally built
        # binaries, instead of the system ones.
        env = {"PATH": "{}:{}".format(
            os.path.join(ROOT, version), os.environ["PATH"])}

        self.fake_juju = self.useFixture(FakeJuju(version=version, env=env))

    def test_up(self):
        """The fake-juju service is connected to the mongodb one."""
        self.assertIn(
            "Using external MongoDB on port %d" % self.fake_juju.mongo.port,
            self.logger.output)

    def test_cli_bootstrap(self):
        """
        The fake-juju command line tool can perform a fake bootstrap. After the
        bootstrap the controller machine (machine 0) is present.
        """
        cli = self.fake_juju.cli()
        cli.execute("bootstrap", "foo", "bar")

        status = json.loads(cli.execute("status", "--format=json"))

        self.assertEqual(
            "running", status["machines"]["0"]["machine-status"]["current"])
        self.assertEqual(
            "started", status["machines"]["0"]["juju-status"]["current"])

    def test_cli_destroy_controller(self):
        """
        The fake-juju command line tool can destroy the fake controller, using
        the normal "destroy-controller" subcommand.
        """
        cli = self.fake_juju.cli()
        cli.execute("bootstrap", "foo", "bar")
        cli.execute("destroy-controller", "-y", "bar")
        self.assertRaises(CalledProcessError, cli.execute, "status")
