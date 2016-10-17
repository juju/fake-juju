# Copyright 2016 Canonical Limited.  All rights reserved.

import os
import os.path
import shutil
import tempfile
import unittest

from fakejuju.failures import Failures


class FailuresTests(unittest.TestCase):

    def setUp(self):
        super(FailuresTests, self).setUp()
        self.dirname = tempfile.mkdtemp(prefix="fakejuju-test-")

    def tearDown(self):
        shutil.rmtree(self.dirname)
        super(FailuresTests, self).tearDown()

    def test_full(self):
        """Failures() works correctly when given all args."""
        entities = [u"x", u"y", u"z"]
        failures = Failures(u"/some/dir", entities)

        self.assertEqual(failures.filename, u"/some/dir/juju-failures")
        self.assertEqual(failures.entities, set(entities))

    def test_minimal(self):
        """Failures() works correctly when given minimal args."""
        failures = Failures(u"/some/dir")

        self.assertEqual(failures.filename, u"/some/dir/juju-failures")
        self.assertEqual(failures.entities, set())

    def test_conversion(self):
        """Failures() doesn't convert any values."""
        failures_str = Failures("/some/dir", ["x", "y", "z"])
        failures_unicode = Failures(u"/some/dir", [u"x", u"y", u"z"])

        self.assertIsInstance(failures_str.filename, str)
        self.assertIsInstance(failures_unicode.filename, unicode)
        for id in failures_str.entities:
            self.assertIsInstance(id, str)
        for id in failures_unicode.entities:
            self.assertIsInstance(id, unicode)

    def test_file_not_created_initially(self):
        """Failures() doesn't create a missing cfgdir until necessary."""
        failures = Failures(self.dirname)

        self.assertFalse(os.path.exists(failures.filename))

    def test_cfgdir_created(self):
        """Failures() creates a missing cfgdir as soon as it's needed."""
        dirname = os.path.join(self.dirname, "subdir")
        self.assertFalse(os.path.exists(dirname))
        failures = Failures(dirname)
        failures.fail_entity("unit-xyz")

        self.assertTrue(os.path.exists(dirname))

    def test_fail_entity_one(self):
        """Failures,fail_entity() writes an initial entry to disk."""
        failures = Failures(self.dirname)
        failures.fail_entity("unit-abc")
        with open(failures.filename) as file:
            data = file.read()

        self.assertEqual(data, "unit-abc\n")

    def test_fail_entity_multiple(self):
        """Failures.fail_entity() correctly writes multiple entries to disk."""
        failures = Failures(self.dirname)
        failures.fail_entity("unit-abc")
        failures.fail_entity("unit-xyz")

        with open(failures.filename) as file:
            data = file.read()
        entities = set(tag for tag in data.splitlines() if tag)
        self.assertEqual(entities, failures.entities)
        self.assertTrue(data.endswith("\n"))

    def test_clear_exists(self):
        """Failures.clear() deletes the failures file if it exists."""
        failures = Failures(self.dirname)
        failures.fail_entity("unit-abc")
        self.assertTrue(os.path.exists(failures.filename))
        failures.clear()

        self.assertFalse(os.path.exists(failures.filename))
        self.assertEqual(failures.entities, set())

    def test_clear_not_exists(self):
        """Failures.clear() does nothing if the failures file is missing."""
        failures = Failures(self.dirname)
        self.assertFalse(os.path.exists(failures.filename))
        failures.clear()

        self.assertFalse(os.path.exists(failures.filename))
