# Copyright 2016 Canonical Limited.  All rights reserved.

import txjuju.cli


def get_bootstrap_spec(name, admin_secret=None):
    """Return the BootstrapSpec instance for the given controller.

    @param name: The controller name to set up.
    @param admin_secret: The admin user password to use.
    """
    type = "dummy"
    default_series = None  # Use the default.
    return txjuju.cli.BootstrapSpec(name, type, default_series, admin_secret)
