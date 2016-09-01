
import datetime
import json
import os
import os.path
import shutil
import subprocess
import tempfile
from unittest import TestCase
from unittest.case import SkipTest

from . import _jujuclient


ROOT_DIR = os.path.dirname(os.path.dirname(__file__))

JUJU_VERSION = os.environ.get("JUJU_VERSION", "1.22.1")
JUJU_FAKE = os.path.join(ROOT_DIR, JUJU_VERSION, JUJU_VERSION)

DUMMY_CHARM = os.path.join(ROOT_DIR, "tests", "charms", "dummy")


def _bootstrap(name, type, env):
    """Return the API endpoint after bootstrapping the controller."""
    if JUJU_VERSION.startswith("1."):
        subprocess.check_call([JUJU_FAKE, "bootstrap", "-e", name], env=env)

        output = subprocess.check_output([JUJU_FAKE, "api-info"], env=env)
        api_info = json.loads(output.decode())
        endpoint = str(api_info["state-servers"][0])
        uuid = api_info["environ-uuid"]
        password = "test"
        return endpoint, uuid, password

    args = [JUJU_FAKE, "bootstrap", "--no-gui", name, type]
    subprocess.check_call(args, env=env)

    args = [JUJU_FAKE, "show-controller", "--format", "json",
            "--show-password", name]
    output = subprocess.check_output(args, env=env)
    api_info = json.loads(output.decode())
    endpoint = str(api_info[name]["details"]["api-endpoints"][0])
    model = api_info[name]["current-model"]
    uuid = api_info[name]["models"][model]["uuid"]
    password = api_info[name]["account"]["password"]
    return endpoint, uuid, password


class JujuFakeTest(TestCase):

    def setUp(self):
        super(JujuFakeTest, self).setUp()

        self.env = os.environ.copy()
        self.juju_home = cfgdir = tempfile.mkdtemp()
        _jujuclient.prepare("dummy", "dummy", cfgdir, self.env, JUJU_VERSION)

        endpoint, uuid, password = _bootstrap("dummy", "dummy", self.env)
        self.api = _jujuclient.connect(endpoint, password, uuid, JUJU_VERSION)

    def tearDown(self):
        destroycmd = "destroy-controller"
        if JUJU_VERSION.startswith("1."):
            destroycmd = "destroy-environment"
        subprocess.check_call([JUJU_FAKE, destroycmd], env=self.env)
        self.api.close()

        shutil.rmtree(self.juju_home)

        super(JujuFakeTest, self).tearDown()

    def test_info(self):
        info = self.api.info()

        if JUJU_VERSION.startswith("1."):
            self.assertEqual("dummy", info["ProviderType"])
        else:
            self.assertEqual("dummy", info["provider-type"])

    def test_local_charm(self):
        charm = self.api.add_local_charm_dir(DUMMY_CHARM, "trusty")
        key = "CharmURL" if JUJU_VERSION.startswith("1.") else "charm-url"
        self.api.deploy("dummy", charm[key], num_units=0)

    def test_run_on_all_machines(self):
        timeout = 5 * 10 ** 9
        now = datetime.datetime.utcnow()
        result = self.api.run_on_all_machines("/foo/bar", timeout=timeout)

        if JUJU_VERSION.startswith("1."):
            self.assertEqual(
                {"Results": [
                    {"Code": 0,
                     "Stdout": "",
                     "Stderr": "",
                     "MachineId": "0",
                     "Error": "",
                     "UnitId": ""}]},
                result)
        else:
            tag = result["results"][0]["action"]["tag"]
            self.assertTrue(tag.startswith("action-"))
            self.maxDiff = None
            enqueuedstr = result["results"][0]["enqueued"]
            enqueued = datetime.datetime.strptime(
                enqueuedstr,"%Y-%m-%dT%H:%M:%SZ")
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
