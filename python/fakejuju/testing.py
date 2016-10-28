# Copyright 2016 Canonical Limited.  All rights reserved.

from fixtures import Fixture, TempDir
from testtools.content import content_from_file

from . import fakejuju


JUJU1_VER = "1.25.6"
JUJU2_VER = "2.0.0"
JUJU_VER = JUJU2_VER


class FakeJujuFixture(Fixture):
    """Manages a fake-juju process."""

    CONTROLLER = "test"
    ADMIN_SECRET = "sekret"
    VERSION = JUJU_VER

    def __init__(self, controller=None, password=None, logs_dir=None,
                 version=None):
        """
        @param logs_dir: If given, copy logs to this directory upon cleanup,
            otherwise, print it as test plain text detail upon failure.
        """
        if controller is None:
            controller = self.CONTROLLER
        if password is None:
            password = self.ADMIN_SECRET
        if version is None:
            version = self.VERSION

        self._controller = controller
        self._password = password
        self._logs_dir = logs_dir
        self._version = version

    def setUp(self):
        super(FakeJujuFixture, self).setUp()
        self._juju_home = self.useFixture(TempDir())
        self.fakejuju = fakejuju.FakeJuju.from_version(
            self._version, self._juju_home.path, self._logs_dir)

        if not self._logs_dir:
            # Attach logs as testtools details.
            self.addDetail("log-file", content_from_file(self._juju.logfile))

        self._juju, self.all_api_info = self.fakejuju.bootstrap(
            self._controller, self._password)

    def cleanUp(self):
        self._juju.destroy_controller(self._controller)
        super(FakeJujuFixture, self).cleanUp()
