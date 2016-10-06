# Copyright 2016 Canonical Limited.  All rights reserved.

import unittest

import txjuju.cli

from fakejuju.fakejuju import get_bootstrap_spec, get_filename


class HelperTests(unittest.TestCase):

    def test_get_bootstrap_spec_full(self):
        spec = get_bootstrap_spec("my-env", "pw")

        self.assertEqual(
            spec,
            txjuju.cli.BootstrapSpec("my-env", "dummy", admin_secret="pw"))

    def test_get_bootstrap_spec_minimal(self):
        spec = get_bootstrap_spec("my-env")

        self.assertEqual(spec, txjuju.cli.BootstrapSpec("my-env", "dummy"))

    def test_get_filename_full(self):
        filename = get_filename("1.25.6", "/spam")

        self.assertEqual(filename, "/spam/fake-juju-1.25.6")

    def test_get_filename_minimal(self):
        filename = get_filename("1.25.6")

        self.assertEqual(filename, "/usr/bin/fake-juju-1.25.6")

    def test_get_filename_empty_bindir(self):
        filename = get_filename("1.25.6", "")

        self.assertEqual(filename, "fake-juju-1.25.6")

    def test_get_filename_missing_version(self):
        with self.assertRaises(ValueError):
            get_filename(None)
        with self.assertRaises(ValueError):
            get_filename("")
