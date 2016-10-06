# Copyright 2016 Canonical Limited.  All rights reserved.

import unittest

import txjuju.cli

from fakejuju.fakejuju import get_bootstrap_spec


class HelperTests(unittest.TestCase):

    def test_get_bootstrap_spec_full(self):
        spec = get_bootstrap_spec("my-env", "pw")

        self.assertEqual(
            spec,
            txjuju.cli.BootstrapSpec("my-env", "dummy", admin_secret="pw"))

    def test_get_bootstrap_spec_minimal(self):
        spec = get_bootstrap_spec("my-env")

        self.assertEqual(spec, txjuju.cli.BootstrapSpec("my-env", "dummy"))
