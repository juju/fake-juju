# Copyright 2016 Canonical Limited.  All rights reserved.

"""Support for interaction with fake-juju.

"fake-juju" is a combination of the juju and jujud commands that is
suitable for use in integration tests.  It exposes a limited subset
of the standard juju subcommands (see FakeJuju in this module for
specifics).  When called without any arguments it runs jujud (using
the dummy provider) with extra logging and testing hooks available to
control failures.  See https://launchpad.net/fake-juju for the project.

The binary is named with the Juju version for which it was built.
For example, for version 1.25.6 the file is named "fake-juju-1.25.6".

fake-juju uses the normal Juju local config directory.  This defaults
to ~/.local/shared/juju and may be set using the JUJU_DATA environment
variable (in 2.x, for 1.x it is JUJU_HOME).

In addition to all the normal Juju environment variables (e.g.
JUJU_DATA), fake-juju uses the following:

  FAKE_JUJU_FAILURES - the path to the failures file
      The Failures class below sets this to $JUJU_DATA/juju-failures.
  FAKE_JUJU_LOGS_DIR - the path to the logs directory
      This defaults to $JUJU_DATA.

fake-juju also creates several extra files:

  $FAKE_JUJU_LOGS_DIR/fake-juju.log - where fake-juju logs are written
  $JUJU_DATA/fakejuju - fake-juju's data cache
  $JUJU_DATA/fifo - a FIFO file that triggers jujud shutdown
  $JUJU_DATA/cert.ca - the API's CA certificate

Normal Juju logging for is written to $JUJU_DATA/fake-juju.log.

Failures may be injected into a running fake-juju (or set before
running).  They may be injected by adding them to the file identified
by $FAKE_JUJU_FAILURES.  The format is a single failure definition per
line.  The syntax of the failure definition depends on the failure.
The currently supported failures (with their definition syntax) are
listed here:

  * when adding a unit with a specific ID
    format: "unit-<ID>"  (e.g. unit-mysql/0)

"""
from pbr.version import VersionInfo

from .fakejuju import get_filename, set_envvars, FakeJuju


__all__ = [
    "__version__",
    "get_bootstrap_spec", "get_filename", "set_envvars",
    "FakeJuju",
    ]


_v = VersionInfo("fakejuju").semantic_version()
__version__ = _v.release_string()
version_info = _v.version_tuple()


def get_bootstrap_spec(name, admin_secret=None):
    """Return the BootstrapSpec instance for the given controller.

    @param name: The controller name to set up.
    @param admin_secret: The admin user password to use.
    """
    import txjuju.cli

    driver = "dummy"
    default_series = None  # Use the default.
    return txjuju.cli.BootstrapSpec(name, driver, default_series, admin_secret)
