# Copyright 2016 Canonical Limited.  All rights reserved.

import os
import unittest

from txjuju import _juju1, _juju2
from txjuju._utils import Executable

from fakejuju.failures import Failures
from fakejuju.fakejuju import get_filename, set_envvars, FakeJuju


class GetFilenameTests(unittest.TestCase):

    def test_all_args(self):
        """get_filename() works correctly when given all args."""
        filename = get_filename("1.25.6", "/spam")

        self.assertEqual(filename, "/spam/fake-juju-1.25.6")

    def test_minimal_args(self):
        """get_filename() works correctly when given minimal args."""
        filename = get_filename("1.25.6")

        self.assertEqual(filename, "/usr/bin/fake-juju-1.25.6")

    def test_empty_bindir(self):
        """get_filename() works correctly when given an empty string
        for bindir."""
        filename = get_filename("1.25.6", "")

        self.assertEqual(filename, "fake-juju-1.25.6")

    def test_missing_version(self):
        """get_filename() fails if version is None or empty."""
        with self.assertRaises(ValueError):
            get_filename(None)
        with self.assertRaises(ValueError):
            get_filename("")


class SetEnvvarsTests(unittest.TestCase):

    def test_all_args(self):
        """set_envvars() works correctly when given all args."""
        envvars = {}
        set_envvars(envvars, "/spam/failures", "/eggs/logsdir")

        self.assertEqual(envvars, {
            "FAKE_JUJU_FAILURES": "/spam/failures",
            "FAKE_JUJU_LOGS_DIR": "/eggs/logsdir",
            })

    def test_minimal_args(self):
        """set_envvars() works correctly when given minimal args."""
        envvars = {}
        set_envvars(envvars)

        self.assertEqual(envvars, {
            "FAKE_JUJU_FAILURES": "",
            "FAKE_JUJU_LOGS_DIR": "",
            })

    def test_start_empty(self):
        """set_envvars() sets all values on an empty dict."""
        envvars = {}
        set_envvars(envvars, "x", "y")

        self.assertEqual(envvars, {
            "FAKE_JUJU_FAILURES": "x",
            "FAKE_JUJU_LOGS_DIR": "y",
            })

    def test_no_collisions(self):
        """set_envvars() sets all values when none are set yet."""
        envvars = {"SPAM": "eggs"}
        set_envvars(envvars, "x", "y")

        self.assertEqual(envvars, {
            "SPAM": "eggs",
            "FAKE_JUJU_FAILURES": "x",
            "FAKE_JUJU_LOGS_DIR": "y",
            })

    def test_empty_to_nonempty(self):
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

    def test_nonempty_to_nonempty(self):
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

    def test_nonempty_to_empty(self):
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


class FakeJujuTests(unittest.TestCase):

    def test_from_version_full(self):
        """FakeJuju.from_version() works correctly when given all args."""
        juju = FakeJuju.from_version(
            "1.25.6", "/a/juju/home", "/logs/dir", "/failures/dir", "/bin/dir")

        self.assertEqual(juju.filename, "/bin/dir/fake-juju-1.25.6")
        self.assertEqual(juju.version, "1.25.6")
        self.assertEqual(juju.cfgdir, "/a/juju/home")
        self.assertEqual(juju.logsdir, "/logs/dir")
        self.assertEqual(juju.failures.filename, "/failures/dir/juju-failures")

    def test_from_version_minimal(self):
        """FakeJuju.from_version() works correctly when given minimal args."""
        juju = FakeJuju.from_version("1.25.6", "/my/juju/home")

        self.assertEqual(juju.filename, "/usr/bin/fake-juju-1.25.6")
        self.assertEqual(juju.version, "1.25.6")
        self.assertEqual(juju.cfgdir, "/my/juju/home")
        self.assertEqual(juju.logsdir, "/my/juju/home")
        self.assertEqual(juju.failures.filename, "/my/juju/home/juju-failures")

    def test_full(self):
        """FakeJuju() works correctly when given all args."""
        cfgdir = "/my/juju/home"
        failures = Failures(cfgdir)
        juju = FakeJuju("/fake-juju", "1.25.6", cfgdir, "/some/logs", failures)

        self.assertEqual(juju.filename, "/fake-juju")
        self.assertEqual(juju.version, "1.25.6")
        self.assertEqual(juju.cfgdir, cfgdir)
        self.assertEqual(juju.logsdir, "/some/logs")
        self.assertIs(juju.failures, failures)

    def test_minimal(self):
        """FakeJuju() works correctly when given minimal args."""
        juju = FakeJuju("/fake-juju", "1.25.6", "/my/juju/home")

        self.assertEqual(juju.filename, "/fake-juju")
        self.assertEqual(juju.version, "1.25.6")
        self.assertEqual(juju.cfgdir, "/my/juju/home")
        self.assertEqual(juju.logsdir, "/my/juju/home")
        self.assertEqual(juju.failures.filename, "/my/juju/home/juju-failures")

    def test_conversions(self):
        """FakeJuju() doesn't convert the type of any value."""
        juju_str = FakeJuju(
            "/fake-juju", "1.25.6", "/x", "/y", Failures("/..."))
        juju_unicode = FakeJuju(
            u"/fake-juju", u"1.25.6", u"/x", u"/y", Failures(u"/..."))

        for name in ('filename version cfgdir logsdir'.split()):
            self.assertIsInstance(getattr(juju_str, name), str)
            self.assertIsInstance(getattr(juju_unicode, name), unicode)

    def test_missing_filename(self):
        """FakeJuju() fails if filename is None or empty."""
        with self.assertRaises(ValueError):
            FakeJuju(None, "1.25.6", "/my/juju/home")
        with self.assertRaises(ValueError):
            FakeJuju("", "1.25.6", "/my/juju/home")

    def test_missing_version(self):
        """FakeJuju() fails if version is None or empty."""
        with self.assertRaises(ValueError):
            FakeJuju("/fake-juju", None, "/my/juju/home")
        with self.assertRaises(ValueError):
            FakeJuju("/fake-juju", "", "/my/juju/home")

    def test_missing_cfgdir(self):
        """FakeJuju() fails if cfgdir is None or empty."""
        with self.assertRaises(ValueError):
            FakeJuju("/fake-juju", "1.25.6", None)
        with self.assertRaises(ValueError):
            FakeJuju("/fake-juju", "1.25.6", "")

    def test_logfile(self):
        """FakeJuju.logfile returns the path to the fake-juju log file."""
        juju = FakeJuju("/fake-juju", "1.25.6", "/x", "/some/logs")

        self.assertEqual(juju.logfile, "/some/logs/fake-juju.log")

    def test_infofile(self):
        """FakeJuju.logfile returns the path to the fake-juju info file."""
        juju = FakeJuju("/fake-juju", "1.25.6", "/x")

        self.assertEqual(juju.infofile, "/x/fakejuju")

    def test_fifo(self):
        """FakeJuju.logfile returns the path to the fake-juju fifo."""
        juju = FakeJuju("/fake-juju", "1.25.6", "/x")

        self.assertEqual(juju.fifo, "/x/fifo")

    def test_cacertfile(self):
        """FakeJuju.cacertfile returns the path to the Juju API cert."""
        juju = FakeJuju("/fake-juju", "1.25.6", "/x")

        self.assertEqual(juju.cacertfile, "/x/cert.ca")

    def test_cli_full(self):
        """FakeJuju.cli() works correctly when given all args."""
        juju = FakeJuju("/fake-juju", "1.25.6", "/x")
        cli = juju.cli({"SPAM": "eggs"})

        self.assertEqual(
            cli._exe,
            Executable("/fake-juju", {
                "SPAM": "eggs",
                "FAKE_JUJU_FAILURES": "/x/juju-failures",
                "FAKE_JUJU_LOGS_DIR": "/x",
                "JUJU_HOME": "/x",
                }),
            )

    def test_cli_minimal(self):
        """FakeJuju.cli() works correctly when given minimal args."""
        juju = FakeJuju("/fake-juju", "1.25.6", "/x")
        cli = juju.cli()

        self.assertEqual(
            cli._exe,
            Executable("/fake-juju", dict(os.environ, **{
                "FAKE_JUJU_FAILURES": "/x/juju-failures",
                "FAKE_JUJU_LOGS_DIR": "/x",
                "JUJU_HOME": "/x",
                })),
            )

    def test_cli_juju1(self):
        """FakeJuju.cli() works correctly for Juju 1.x."""
        juju = FakeJuju.from_version("1.25.6", "/x")
        cli = juju.cli()

        self.assertEqual(cli._exe.envvars["JUJU_HOME"], "/x")
        self.assertIsInstance(cli._juju, _juju1.CLIHooks)

    def test_cli_juju2(self):
        """FakeJuju.cli() works correctly for Juju 2.x."""
        juju = FakeJuju.from_version("2.0.0", "/x")
        cli = juju.cli()

        self.assertEqual(cli._exe.envvars["JUJU_DATA"], "/x")
        self.assertIsInstance(cli._juju, _juju2.CLIHooks)
