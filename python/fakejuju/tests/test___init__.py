# Copyright 2016 Canonical Limited.  All rights reserved.

import unittest

import txjuju.cli

from fakejuju import get_bootstrap_spec


class GetBootstrapTests(unittest.TestCase):

    def test_get_bootstrap_spec_full(self):
        """get_bootstrap_spec() works correctly when given all args."""
        spec = get_bootstrap_spec("my-env", "pw")

        self.assertEqual(
            spec,
            txjuju.cli.BootstrapSpec("my-env", "dummy", admin_secret="pw"))

    def test_get_bootstrap_spec_minimal(self):
        """get_bootstrap_spec() works correctly when given minimal args."""
        spec = get_bootstrap_spec("my-env")

        self.assertEqual(spec, txjuju.cli.BootstrapSpec("my-env", "dummy"))
