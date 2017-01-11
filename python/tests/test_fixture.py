import os

from testtools import TestCase

from fixtures import FakeLogger

from txfixtures import Reactor

from fakejuju.fixture import (
    ROOT,
    JujuMongoDB,
    FakeJuju,
)

FAKE_JUJUD = os.path.join(ROOT, "2.0.2", "fake-jujud")


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
            self.mongodb.port, fake_jujud=FAKE_JUJUD))

    def test_up(self):
        """The fake-juju service is connected to the mongodb one."""
        self.assertIn(
            "Using external MongoDB on port %d" % self.mongodb.port,
            self.logger.output)
