# Copyright 2016 Canonical Limited.  All rights reserved.

from collections import namedtuple
import os.path

import txjuju.cli

from .failures import Failures


def get_bootstrap_spec(name, admin_secret=None):
    """Return the BootstrapSpec instance for the given controller.

    @param name: The controller name to set up.
    @param admin_secret: The admin user password to use.
    """
    type = "dummy"
    default_series = None  # Use the default.
    return txjuju.cli.BootstrapSpec(name, type, default_series, admin_secret)


def get_filename(version, bindir=None):
    """Return the full path to the fake-juju binary for the given version.

    @param version: The Juju version to use.
    @param bindir: The directory containing the fake-juju binary.
        This defaults to /usr/bin.
    """
    if not version:
        raise ValueError("version not provided")
    filename = "fake-juju-{}".format(version)
    if bindir is None:
        # XXX Search $PATH.
        bindir = "/usr/bin"
    return os.path.join(bindir, filename)


def set_envvars(envvars, failures_filename=None, logsdir=None):
    """Return the environment variables with which to run fake-juju.

    @param envvars: The env dict to update.
    @param failures_filename: The path to the failures file that
        fake-juju will use.
    @params logsdir: The path to the directory where fake-juju will
        write its log files.
    """
    envvars["FAKE_JUJU_FAILURES"] = failures_filename or ""
    envvars["FAKE_JUJU_LOGS_DIR"] = logsdir or ""


class FakeJuju(
        namedtuple("FakeJuju", "filename version cfgdir logsdir failures")):
    """The fundamental details for fake-juju."""

    @classmethod
    def from_version(cls, version, cfgdir,
             logsdir=None, failuresdir=None, bindir=None):
        """Return a new instance given the provided information.

        @param version: The Juju version to fake.
        @param cfgdir: The "juju home" directory to use.
        @param logsdir: The directory where logs will be written.
            This defaults to cfgdir.
        @params failuresdir: The directory where failure injection
            is managed.
        @param bindir: The directory containing the fake-juju binary.
            This defaults to /usr/bin.
        """
        if logsdir is None:
            logsdir = cfgdir
        if failuresdir is None:
            failuresdir = cfgdir
        filename = get_filename(version, bindir=bindir)
        failures = Failures(failuresdir)
        return cls(filename, version, cfgdir, logsdir, failures)

    def __new__(cls, filename, version, cfgdir, logsdir=None, failures=None):
        """
        @param filename: The path to the fake-juju binary.
        @param version: The Juju version to fake.
        @param cfgdir: The "juju home" directory to use.
        @param logsdir: The directory where logs will be written.
            This defaults to cfgdir.
        @param failures: The set of fake-juju failures to use.
        """
        filename = unicode(filename) if filename else None
        version = unicode(version) if version else None
        cfgdir = unicode(cfgdir) if cfgdir else None
        logsdir = unicode(logsdir) if logsdir is not None else cfgdir
        if failures is None and cfgdir:
            failures = Failures(cfgdir)
        return super(FakeJuju, cls).__new__(
            cls, filename, version, cfgdir, logsdir, failures)

    def __init__(self, *args, **kwargs):
        if not self.filename:
            raise ValueError("missing filename")
        if not self.version:
            raise ValueError("missing version")
        if not self.cfgdir:
            raise ValueError("missing cfgdir")
        if not self.logsdir:
            raise ValueError("missing logsdir")
        if self.failures is None:
            raise ValueError("missing failures")

    @property
    def logfile(self):
        """The path to fake-juju's log file."""
        return os.path.join(self.logsdir, "fake-juju.log")

    @property
    def infofile(self):
        """The path to fake-juju's data cache."""
        return os.path.join(self.cfgdir, "fakejuju")

    @property
    def fifo(self):
        """The path to the fifo file that triggers shutdown."""
        return os.path.join(self.cfgdir, "fifo")

    @property
    def cacertfile(self):
        """The path to the API server's certificate."""
        return os.path.join(self.cfgdir, "cert.ca")
