# Copyright 2016 Canonical Limited.  All rights reserved.

import unittest

import txjuju.cli

from fakejuju.fakejuju import get_bootstrap_spec, get_filename, set_envvars


class HelperTests(unittest.TestCase):

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

    def test_get_filename_full(self):
        """get_filename() works correctly when given all args."""
        filename = get_filename("1.25.6", "/spam")

        self.assertEqual(filename, "/spam/fake-juju-1.25.6")

    def test_get_filename_minimal(self):
        """get_filename() works correctly when given minimal args."""
        filename = get_filename("1.25.6")

        self.assertEqual(filename, "/usr/bin/fake-juju-1.25.6")

    def test_get_filename_empty_bindir(self):
        """get_filename() works correctly when given an empty string
        for bindir."""
        filename = get_filename("1.25.6", "")

        self.assertEqual(filename, "fake-juju-1.25.6")

    def test_get_filename_missing_version(self):
        """get_filename() fails if version is None or empty."""
        with self.assertRaises(ValueError):
            get_filename(None)
        with self.assertRaises(ValueError):
            get_filename("")

    def test_set_envvars_full(self):
        """set_envvars() works correctly when given all args."""
        envvars = {}
        set_envvars(envvars, "/spam/failures", "/eggs/logsdir")

        self.assertEqual(envvars, {
            "FAKE_JUJU_FAILURES": "/spam/failures",
            "FAKE_JUJU_LOGS_DIR": "/eggs/logsdir",
            })

    def test_set_envvars_minimal(self):
        """set_envvars() works correctly when given minimal args."""
        envvars = {}
        set_envvars(envvars)

        self.assertEqual(envvars, {
            "FAKE_JUJU_FAILURES": "",
            "FAKE_JUJU_LOGS_DIR": "",
            })

    def test_set_envvars_start_empty(self):
        """set_envvars() sets all values on an empty dict."""
        envvars = {}
        set_envvars(envvars, "x", "y")

        self.assertEqual(envvars, {
            "FAKE_JUJU_FAILURES": "x",
            "FAKE_JUJU_LOGS_DIR": "y",
            })

    def test_set_envvars_no_collisions(self):
        """set_envvars() sets all values when none are set yet."""
        envvars = {"SPAM": "eggs"}
        set_envvars(envvars, "x", "y")

        self.assertEqual(envvars, {
            "SPAM": "eggs",
            "FAKE_JUJU_FAILURES": "x",
            "FAKE_JUJU_LOGS_DIR": "y",
            })

    def test_set_envvars_empty_to_nonempty(self):
        """set_envvars() updates empty values."""
        envvars = {
            "FAKE_JUJU_FAILURES": "",
            "FAKE_JUJU_LOGS_DIR": "",
            }
        set_envvars(envvars, "x", "y")

        self.assertEqual(envvars, {
            "FAKE_JUJU_FAILURES": "x",
            "FAKE_JUJU_LOGS_DIR": "y",
            })

    def test_set_envvars_nonempty_to_nonempty(self):
        """set_envvars() overwrites existing values."""
        envvars = {
            "FAKE_JUJU_FAILURES": "spam",
            "FAKE_JUJU_LOGS_DIR": "ham",
            }
        set_envvars(envvars, "x", "y")

        self.assertEqual(envvars, {
            "FAKE_JUJU_FAILURES": "x",
            "FAKE_JUJU_LOGS_DIR": "y",
            })

    def test_set_envvars_nonempty_to_empty(self):
        """set_envvars() with no args "unsets" existing values."""
        envvars = {
            "FAKE_JUJU_FAILURES": "x",
            "FAKE_JUJU_LOGS_DIR": "y",
            }
        set_envvars(envvars)

        self.assertEqual(envvars, {
            "FAKE_JUJU_FAILURES": "",
            "FAKE_JUJU_LOGS_DIR": "",
            })
