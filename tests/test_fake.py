
import datetime
import json
import os
import os.path
import shutil
import subprocess
import tempfile
import unittest
import warnings

from . import _jujuclient


# Quiet pyflakes on Python 2.
try:
    ResourceWarning
except NameError:
    ResourceWarning = Warning


ROOT_DIR = os.path.dirname(os.path.dirname(__file__))

JUJU_VERSION = os.environ.get("JUJU_VERSION", "1.22.1")
JUJU_FAKE = os.path.join(ROOT_DIR, JUJU_VERSION, JUJU_VERSION)

DUMMY_CHARM = os.path.join(ROOT_DIR, "tests", "charms", "dummy")


class _JujuFakeTest:

    # These are set on the child classes.
    destroycmd = None
    bootstrap = None

    def setUp(self):
        super(_JujuFakeTest, self).setUp()

        self.env = os.environ.copy()
        self.juju_home = cfgdir = tempfile.mkdtemp(prefix="fake-juju-test-")
        _jujuclient.prepare("dummy", "dummy", cfgdir, self.env, JUJU_VERSION)

        endpoint, uuid, password = self.bootstrap("dummy", "dummy", self.env)
        self.api = _jujuclient.connect(endpoint, password, uuid, JUJU_VERSION)

    def tearDown(self):
        subprocess.check_call([JUJU_FAKE, self.destroycmd], env=self.env)
        self.api.close()

        shutil.rmtree(self.juju_home)

        super(_JujuFakeTest, self).tearDown()


@unittest.skipUnless(JUJU_VERSION.startswith("1."), "wrong Juju version")
class Juju1FakeTest(_JujuFakeTest, unittest.TestCase):

    destroycmd = "destroy-environment"

    def bootstrap(self, name, type, env):
        """Return the API endpoint after bootstrapping the controller."""
        subprocess.check_call([JUJU_FAKE, "bootstrap", "-e", name], env=env)

        output = subprocess.check_output([JUJU_FAKE, "api-info"], env=env)
        api_info = json.loads(output.decode())
        endpoint = str(api_info["state-servers"][0])
        uuid = api_info["environ-uuid"]
        password = "test"
        return endpoint, uuid, password

    def test_info(self):
        info = self.api.info()
        self.assertEqual("dummy", info["ProviderType"])

    def test_local_charm(self):
        with warnings.catch_warnings():
            warnings.simplefilter("ignore", ResourceWarning)
            charm = self.api.add_local_charm_dir(DUMMY_CHARM, "trusty")
        self.api.deploy("dummy", charm['CharmURL'], num_units=0)

    def test_run_on_all_machines(self):
        timeout = 5 * 10 ** 9
        result = self.api.run_on_all_machines("/foo/bar", timeout=timeout)

        self.assertEqual(
            {"Results": [
                {"Code": 0,
                 "Stdout": "",
                 "Stderr": "",
                 "MachineId": "0",
                 "Error": "",
                 "UnitId": ""}]},
            result)


@unittest.skipUnless(JUJU_VERSION.startswith("2."), "wrong Juju version")
class Juju2FakeTest(_JujuFakeTest, unittest.TestCase):

    destroycmd = "destroy-controller"

    def bootstrap(self, name, type, env):
        """Return the API endpoint after bootstrapping the controller."""
        args = [JUJU_FAKE, "bootstrap", "--no-gui", type, name]
        subprocess.check_call(args, env=env)

        args = [JUJU_FAKE, "show-controller", "--format", "json",
                "--show-password", name]
        output = subprocess.check_output(args, env=env)
        api_info = json.loads(output.decode())
        endpoint = str(api_info[name]["details"]["api-endpoints"][0])
        model = api_info[name]["current-model"].split("/", 1)[-1]
        uuid = api_info[name]["models"][model]["uuid"]
        password = api_info[name]["account"]["password"]
        return endpoint, uuid, password

    def test_info(self):
        info = self.api.info()
        self.assertEqual("dummy", info["provider-type"])

    def test_local_charm(self):
        with warnings.catch_warnings():
            warnings.simplefilter("ignore", ResourceWarning)
            charm = self.api.add_local_charm_dir(DUMMY_CHARM, "trusty")
        self.api.deploy("dummy", charm['charm-url'], num_units=0)

    def test_run_on_all_machines(self):
        timeout = 5 * 10 ** 9
        now = datetime.datetime.utcnow()
        result = self.api.run_on_all_machines("/foo/bar", timeout=timeout)

        tag = result["results"][0]["action"]["tag"]
        self.assertTrue(tag.startswith("action-"))
        self.maxDiff = None
        enqueuedstr = result["results"][0]["enqueued"]
        enqueued = datetime.datetime.strptime(
            enqueuedstr, "%Y-%m-%dT%H:%M:%SZ")
        self.assertLess(enqueued - now, datetime.timedelta(seconds=1))
        self.assertEqual(
            {"results": [
                {"action": {
                    "name": "juju-run",
                    "parameters": {
                        "command": "/foo/bar",
                        "timeout": 5000000000,
                        },
                    "receiver": "machine-0",
                    "tag": tag,
                    },
                 "completed": "0001-01-01T00:00:00Z",
                 "enqueued": enqueuedstr,
                 "started": "0001-01-01T00:00:00Z",
                 "status": "pending",
                 }]},
            result)
