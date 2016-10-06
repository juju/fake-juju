# Copyright 2016 Canonical Limited.  All rights reserved.

import os.path

import txjuju.cli


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
