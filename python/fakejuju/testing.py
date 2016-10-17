# Copyright 2016 Canonical Limited.  All rights reserved.

import txjuju
from fixtures import Fixture, TempDir
from testtools.content import content_from_file

from . import fakejuju


JUJU1_VER = "1.25.6"
JUJU2_VER = "2.0-beta17"
JUJU_VER = JUJU1_VER


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
        self._juju = fakejuju.FakeJuju.make(
            self._juju_home.path, self._version, self._logs_dir)

        if not self._logs_dir:
            # Attach logs as testtools details.
            self.addDetail("log-file", content_from_file(self._juju.logfile))

        spec = fakejuju.get_bootstrap_spec(self._controller, self._password)
        cfgfile = txjuju.prepare_for_bootstrap(
            spec, self._version, self._juju_home)
        cli = self._juju.cli()
        cli.bootstrap(spec, cfgfile=cfgfile)
        api_info = cli.api_info(spec.name)
        if self._version.startswith("1."):
            # fake-juju doesn't give us the password, so we have to
            # set it here.
            api_info = api_info._replace(password=self._password)
        self.api_info = api_info

    def cleanUp(self):
        self._juju.destroy_controller(self._controller)
        super(FakeJujuFixture, self).cleanUp()

    def add_failure(self, entity):
        """Make the given entity fail with an error status."""
        self._juju.failures.fail_entity(entity)
