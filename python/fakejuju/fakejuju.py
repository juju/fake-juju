# Copyright 2016 Canonical Limited.  All rights reserved.

import os.path

import txjuju
import txjuju.cli

from .failures import Failures


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


def set_envvars(envvars, datadir=None, failures_filename=None, logsdir=None):
    """Return the environment variables with which to run fake-juju.

    @param envvars: The env dict to update.
    @param datadir: The fake-juju data directory.
    @param failures_filename: The path to the failures file that
        fake-juju will use.
    @params logsdir: The path to the directory where fake-juju will
        write its log files.
    """
    envvars["FAKE_JUJU_DATA_DIR"] = datadir or ""
    envvars["FAKE_JUJU_FAILURES"] = failures_filename or ""
    envvars["FAKE_JUJU_LOGS_DIR"] = logsdir or ""


class FakeJuju(object):
    """The fundamental details for fake-juju."""

    @classmethod
    def from_version(cls, version, datadir,
                     logsdir=None, failuresdir=None, bindir=None):
        """Return a new instance given the provided information.

        @param version: The Juju version to fake.
        @param datadir: The directory in which to store files specific
            to fake-juju.
        @param logsdir: The directory where logs will be written.
            This defaults to datadir.
        @params failuresdir: The directory where failure injection
            is managed.
        @param bindir: The directory containing the fake-juju binary.
            This defaults to /usr/bin.
        """
        if failuresdir is None:
            failuresdir = datadir
        filename = get_filename(version, bindir=bindir)
        failures = Failures(failuresdir)
        return cls(filename, version, datadir, logsdir, failures)

    def __init__(self, filename, version, datadir,
                 logsdir=None, failures=None):
        """
        @param filename: The path to the fake-juju binary.
        @param version: The Juju version to fake.
        @param datadir: The directory in which to store files specific
            to fake-juju.
        @param logsdir: The directory where logs will be written.
            This defaults to datadir.
        @param failures: The set of fake-juju failures to use.
        """
        logsdir = logsdir if logsdir is not None else datadir
        if failures is None and datadir:
            failures = Failures(datadir)

        if not filename:
            raise ValueError("missing filename")
        if not version:
            raise ValueError("missing version")
        if not datadir:
            raise ValueError("missing datadir")
        if not logsdir:
            raise ValueError("missing logsdir")
        if failures is None:
            raise ValueError("missing failures")

        self.filename = filename
        self.version = version
        self.datadir = datadir
        self.logsdir = logsdir
        self.failures = failures

    @property
    def logfile(self):
        """The path to fake-juju's log file."""
        return os.path.join(self.logsdir, "fake-juju.log")

    @property
    def infofile(self):
        """The path to fake-juju's data cache."""
        return os.path.join(self.datadir, "fakejuju")

    @property
    def fifo(self):
        """The path to the fifo file that triggers shutdown."""
        return os.path.join(self.datadir, "fifo")

    @property
    def cacertfile(self):
        """The path to the API server's certificate."""
        return os.path.join(self.datadir, "cert.ca")

    def cli(self, cfgdir, envvars=None):
        """Return the txjuju.cli.CLI for this fake-juju.

        Currently fake-juju supports only the following juju subcommands:

        * bootstrap
          Not that this only supports the dummy provider and the local
          system is only minimally impacted.
        * api-info
          Note that passwords are always omited, even if requested.
        * api-endpoints
        * destroy-environment

        Note that fake-juju ignores local config files.
        """
        if envvars is None:
            envvars = os.environ
        envvars = dict(envvars)
        set_envvars(
            envvars, self.datadir, self.failures._filename, self.logsdir)
        return txjuju.cli.CLI.from_version(
            self.filename, self.version, cfgdir, envvars)

    def bootstrap(self, name, cfgdir, admin_secret=None):
        """Return the CLI and APIInfo after bootstrapping from scratch."""
        from . import get_bootstrap_spec
        spec = get_bootstrap_spec(name, admin_secret)
        cfgfile = txjuju.prepare_for_bootstrap(spec, self.version, cfgdir)
        cli = self.cli(cfgdir)
        cli.bootstrap(spec, cfgfile=cfgfile)
        api_info = cli.api_info(spec.name)
        return cli, api_info

    def is_bootstrapped(self):
        """Return True if a fake-juju controller is running."""
        return os.path.exists(self.fifo)
