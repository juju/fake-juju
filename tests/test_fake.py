
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
        endpoint = "wss://" + str(api_info["state-servers"][0]) + "/"
        return endpoint

    args = [JUJU_FAKE, "bootstrap", "--no-gui", name, type]
    subprocess.check_call(args, env=env)

    args = [JUJU_FAKE, "show-controller", "--format", "json", name]
    output = subprocess.check_output(args, env=env)
    api_info = json.loads(output.decode())
    endpoints = api_info[name]["details"]["api-endpoints"]
    endpoint = "wss://" + str(endpoints[0]) + "/"
    return endpoint


class JujuFakeTest(TestCase):

    def setUp(self):
        super(JujuFakeTest, self).setUp()
        if JUJU_VERSION.startswith("2.0"):
            raise SkipTest("Juju 2.0 still not fully supported")

        self.env = os.environ.copy()
        self.juju_home = cfgdir = tempfile.mkdtemp()
        _jujuclient.prepare("dummy", "dummy", cfgdir, self.env, JUJU_VERSION)

        endpoint = _bootstrap("dummy", "dummy", self.env)
        self.api = _jujuclient.connect(endpoint, JUJU_VERSION)

    def tearDown(self):
        subprocess.check_call([JUJU_FAKE, "destroy-environment"], env=self.env)
        shutil.rmtree(self.juju_home)
        super(JujuFakeTest, self).tearDown()

    def test_info(self):
        info = self.api.info()
        self.assertEqual("dummy", info["ProviderType"])

    def test_local_charm(self):
        charm = self.api.add_local_charm_dir(DUMMY_CHARM, "trusty")
        self.api.deploy("dummy", charm["CharmURL"], num_units=0)

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
