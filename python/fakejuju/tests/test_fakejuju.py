# Copyright 2016 Canonical Limited.  All rights reserved.

from contextlib import contextmanager
import json
import os
import shutil
import tempfile
import unittest

from txjuju import _juju1, _juju2
from txjuju._utils import Executable
import txjuju.cli
import yaml

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
        set_envvars(envvars, "/spam", "/spam/failures", "/eggs/logsdir")

        self.assertEqual(envvars, {
            "FAKE_JUJU_DATA_DIR": "/spam",
            "FAKE_JUJU_FAILURES": "/spam/failures",
            "FAKE_JUJU_LOGS_DIR": "/eggs/logsdir",
            })

    def test_minimal_args(self):
        """set_envvars() works correctly when given minimal args."""
        envvars = {}
        set_envvars(envvars)

        self.assertEqual(envvars, {
            "FAKE_JUJU_DATA_DIR": "",
            "FAKE_JUJU_FAILURES": "",
            "FAKE_JUJU_LOGS_DIR": "",
            })

    def test_start_empty(self):
        """set_envvars() sets all values on an empty dict."""
        envvars = {}
        set_envvars(envvars, "w", "x", "y")

        self.assertEqual(envvars, {
            "FAKE_JUJU_DATA_DIR": "w",
            "FAKE_JUJU_FAILURES": "x",
            "FAKE_JUJU_LOGS_DIR": "y",
            })

    def test_no_collisions(self):
        """set_envvars() sets all values when none are set yet."""
        envvars = {"SPAM": "eggs"}
        set_envvars(envvars, "w", "x", "y")

        self.assertEqual(envvars, {
            "SPAM": "eggs",
            "FAKE_JUJU_DATA_DIR": "w",
            "FAKE_JUJU_FAILURES": "x",
            "FAKE_JUJU_LOGS_DIR": "y",
            })

    def test_empty_to_nonempty(self):
        """set_envvars() updates empty values."""
        envvars = {
            "FAKE_JUJU_DATA_DIR": "",
            "FAKE_JUJU_FAILURES": "",
            "FAKE_JUJU_LOGS_DIR": "",
            }
        set_envvars(envvars, "w", "x", "y")

        self.assertEqual(envvars, {
            "FAKE_JUJU_DATA_DIR": "w",
            "FAKE_JUJU_FAILURES": "x",
            "FAKE_JUJU_LOGS_DIR": "y",
            })

    def test_nonempty_to_nonempty(self):
        """set_envvars() overwrites existing values."""
        envvars = {
            "FAKE_JUJU_DATA_DIR": "spam",
            "FAKE_JUJU_FAILURES": "spam",
            "FAKE_JUJU_LOGS_DIR": "ham",
            }
        set_envvars(envvars, "w", "x", "y")

        self.assertEqual(envvars, {
            "FAKE_JUJU_DATA_DIR": "w",
            "FAKE_JUJU_FAILURES": "x",
            "FAKE_JUJU_LOGS_DIR": "y",
            })

    def test_nonempty_to_empty(self):
        """set_envvars() with no args "unsets" existing values."""
        envvars = {
            "FAKE_JUJU_DATA_DIR": "w",
            "FAKE_JUJU_FAILURES": "x",
            "FAKE_JUJU_LOGS_DIR": "y",
            }
        set_envvars(envvars)

        self.assertEqual(envvars, {
            "FAKE_JUJU_DATA_DIR": "",
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
        self.assertEqual(juju.datadir, "/a/juju/home")
        self.assertEqual(juju.logsdir, "/logs/dir")
        self.assertEqual(juju.failures.filename, "/failures/dir/juju-failures")

    def test_from_version_minimal(self):
        """FakeJuju.from_version() works correctly when given minimal args."""
        juju = FakeJuju.from_version("1.25.6", "/my/juju/home")

        self.assertEqual(juju.filename, "/usr/bin/fake-juju-1.25.6")
        self.assertEqual(juju.version, "1.25.6")
        self.assertEqual(juju.datadir, "/my/juju/home")
        self.assertEqual(juju.logsdir, "/my/juju/home")
        self.assertEqual(juju.failures.filename, "/my/juju/home/juju-failures")

    def test_full(self):
        """FakeJuju() works correctly when given all args."""
        datadir = "/my/juju/home"
        failures = Failures(datadir)
        juju = FakeJuju(
            "/fake-juju", "1.25.6", datadir, "/some/logs", failures)

        self.assertEqual(juju.filename, "/fake-juju")
        self.assertEqual(juju.version, "1.25.6")
        self.assertEqual(juju.datadir, datadir)
        self.assertEqual(juju.logsdir, "/some/logs")
        self.assertIs(juju.failures, failures)

    def test_minimal(self):
        """FakeJuju() works correctly when given minimal args."""
        juju = FakeJuju("/fake-juju", "1.25.6", "/my/juju/home")

        self.assertEqual(juju.filename, "/fake-juju")
        self.assertEqual(juju.version, "1.25.6")
        self.assertEqual(juju.datadir, "/my/juju/home")
        self.assertEqual(juju.logsdir, "/my/juju/home")
        self.assertEqual(juju.failures.filename, "/my/juju/home/juju-failures")

    def test_conversions(self):
        """FakeJuju() doesn't convert the type of any value."""
        juju_str = FakeJuju(
            "/fake-juju", "1.25.6", "/x", "/y", Failures("/..."))
        juju_unicode = FakeJuju(
            u"/fake-juju", u"1.25.6", u"/x", u"/y", Failures(u"/..."))

        for name in ('filename version datadir logsdir'.split()):
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

    def test_missing_datadir(self):
        """FakeJuju() fails if datadir is None or empty."""
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
        cli = juju.cli("/y", {"SPAM": "eggs"})

        self.assertEqual(
            cli._exe,
            Executable("/fake-juju", {
                "SPAM": "eggs",
                "FAKE_JUJU_DATA_DIR": "/x",
                "FAKE_JUJU_FAILURES": "/x/juju-failures",
                "FAKE_JUJU_LOGS_DIR": "/x",
                "JUJU_HOME": "/y",
                }),
            )

    def test_cli_minimal(self):
        """FakeJuju.cli() works correctly when given minimal args."""
        juju = FakeJuju("/fake-juju", "1.25.6", "/x")
        cli = juju.cli("/y")

        self.assertEqual(
            cli._exe,
            Executable("/fake-juju", dict(os.environ, **{
                "FAKE_JUJU_DATA_DIR": "/x",
                "FAKE_JUJU_FAILURES": "/x/juju-failures",
                "FAKE_JUJU_LOGS_DIR": "/x",
                "JUJU_HOME": "/y",
                })),
            )

    def test_cli_juju1(self):
        """FakeJuju.cli() works correctly for Juju 1.x."""
        juju = FakeJuju.from_version("1.25.6", "/x")
        cli = juju.cli("/y")

        self.assertEqual(cli._exe.envvars["JUJU_HOME"], "/y")
        self.assertIsInstance(cli._juju, _juju1.CLIHooks)

    def test_cli_juju2(self):
        """FakeJuju.cli() works correctly for Juju 2.x."""
        juju = FakeJuju.from_version("2.0.0", "/x")
        cli = juju.cli("/y")

        self.assertEqual(cli._exe.envvars["JUJU_DATA"], "/y")
        self.assertIsInstance(cli._juju, _juju2.CLIHooks)

    def test_bootstrap(self):
        """FakeJuju.bootstrap() bootstraps from scratch using fake-juju."""
        expected = txjuju.cli.APIInfo(
                endpoints=['localhost:12727'],
                user='admin',
                password='dummy-secret',
                model_uuid='deadbeef-0bad-400d-8000-4b1d0d06f00d',
                )
        version = "1.25.6"
        with tempdir() as testdir:
            bindir = os.path.join(testdir, "bin")
            datadir = os.path.join(testdir, "fakejuju")
            cfgdir = os.path.join(testdir, ".juju")

            logfilename = write_fakejuju_script(
                version, bindir, datadir, cfgdir, expected)
            fakejuju = FakeJuju.from_version(version, cfgdir, bindir=bindir)

            cli, api_info = fakejuju.bootstrap("spam", cfgdir, "secret")

            files = []
            files.extend(os.path.join(os.path.basename(datadir), name)
                         for name in os.listdir(datadir))
            files.extend(os.path.join(os.path.basename(cfgdir), name)
                         for name in os.listdir(cfgdir))
            with open(os.path.join(cfgdir, "environments.yaml")) as envfile:
                data = envfile.read()

            cli.destroy_controller()
            with open(logfilename) as logfile:
                calls = [line.strip() for line in logfile]

        self.maxDiff = None
        self.assertEqual(api_info, {
            'controller': expected,
            None: expected._replace(model_uuid=None),
            })
        subcommands = []
        for call in calls:
            args = call.split()
            self.assertEqual(os.path.basename(args[0]), "fake-juju-" + version)
            subcommands.append(args[1])
        self.assertEqual(subcommands, [
            "bootstrap",
            "api-info",
            "destroy-environment",
            ])
        self.assertItemsEqual(files, [
            '.juju/environments',
            '.juju/environments.yaml',
            'fakejuju/cert.ca',
            'fakejuju/fake-juju.log',
            'fakejuju/fakejuju',
            'fakejuju/fifo',
            ])
        self.assertEqual(yaml.load(data), {
            "environments": {
                "spam": {
                    "admin-secret": "secret",
                    "default-series": "trusty",
                    "type": "dummy",
                    },
                },
            })

    def test_is_bootstrapped_true(self):
        """FakeJuju.is_bootstrapped() returns True if the fifo file exists."""
        with tempdir() as datadir:
            fakejuju = FakeJuju.from_version("1.25.6", datadir)
            with open(fakejuju.fifo, "w"):
                pass
            result = fakejuju.is_bootstrapped()

        self.assertTrue(result)

    def test_is_bootstrapped_false(self):
        """FakeJuju.is_bootstrapped() returns False if the fifo is gone."""
        with tempdir() as datadir:
            fakejuju = FakeJuju.from_version("1.25.6", datadir)
            result = fakejuju.is_bootstrapped()

        self.assertFalse(result)

    def test_is_bootstrapped_datadir_missing(self):
        """FakeJuju.is_bootstrapped() returns False if the data dir is gone."""
        fakejuju = FakeJuju.from_version("1.25.6", "/tmp/fakejuju-no-exist")
        result = fakejuju.is_bootstrapped()

        self.assertFalse(result)


FAKE_JUJU_SCRIPT = """\
#!/usr/bin/env python

import os.path
import sys

with open("{logfile}", "a") as logfile:
    logfile.write(" ".join(sys.argv) + "\\n")

if sys.argv[1] == "bootstrap":
    for filename in ("cert.ca", "fake-juju.log", "fakejuju", "fifo"):
        with open(os.path.join("{datadir}", filename), "w"):
            pass  # Touch the file.
    for filename in ("environments",):
        with open(os.path.join("{cfgdir}", filename), "w"):
            pass  # Touch the file.
elif sys.argv[1] in ("api-info", "show-controller"):
    print('''{output}''')

"""


def write_fakejuju_script(version, bindir, datadir, cfgdir, api_info):
    if version.startswith("1."):
        raw_api_info = {
            "state-servers": api_info.endpoints,
            "user": api_info.user,
            "password": api_info.password,
            "environ-uuid": api_info.model_uuid,
            }
    else:
        raw_api_info = {
            "details": {
                "api-endpoints": api_info.endpoints,
                },
            "account": {
                "user": api_info.user + "@local",
                "password": api_info.password,
                },
            "models": {
                "controller": {
                    "uuid": api_info.model_uuid,
                    },
                "default": {
                    "uuid": api_info.model_uuid,
                    },
                },
            }
    output = json.dumps(raw_api_info)

    logfile = os.path.join(bindir, "calls.log")
    script = FAKE_JUJU_SCRIPT.format(
        datadir=datadir, cfgdir=cfgdir, logfile=logfile, output=output)
    filename = get_filename(version, bindir)
    os.makedirs(os.path.dirname(filename))
    with open(filename, "w") as scriptfile:
        scriptfile.write(script)
    os.chmod(filename, 0o755)
    os.makedirs(datadir)

    return logfile


@contextmanager
def tempdir():
    cfgdir = tempfile.mkdtemp(prefix="fakejuju-test-")
    try:
        yield cfgdir
    finally:
        shutil.rmtree(cfgdir)
