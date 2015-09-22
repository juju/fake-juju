import os
import tempfile
import shutil
import subprocess
import json


from jujuclient import Environment

from unittest import TestCase

ROOT_DIR = os.path.dirname(os.path.dirname(__file__))

JUJU_VERSION = os.environ.get("JUJU_VERSION", "1.22.1")
JUJU_FAKE = os.path.join(ROOT_DIR, JUJU_VERSION, JUJU_VERSION)

DUMMY_CHARM = os.path.join(ROOT_DIR, "tests", "charms", "dummy")

ENVIRONMENTS_YAML = """environments:
    test:
        admin-secret: test
        default-series: trusty
        type: dummy
"""


class JujuFakeTest(TestCase):

    def setUp(self):
        super(JujuFakeTest, self).setUp()
        self.juju_home = tempfile.mkdtemp()
        environments_yaml = os.path.join(self.juju_home, "environments.yaml")
        with open(environments_yaml, "w") as fd:
            fd.write(ENVIRONMENTS_YAML)
        self.env = os.environ.copy()
        self.env["JUJU_HOME"] = self.juju_home
        self.juju_fake = os.path.join(JUJU_VERSION, JUJU_VERSION)
        subprocess.check_call([JUJU_FAKE, "bootstrap"], env=self.env)
        output = subprocess.check_output([JUJU_FAKE, "api-info"], env=self.env)
        api_info = json.loads(output)
        endpoint = "wss://" + str(api_info["state-servers"][0]) + "/"
        self.environment = Environment(endpoint)
        self.environment.login("test")

    def tearDown(self):
        subprocess.check_call([JUJU_FAKE, "destroy-environment"], env=self.env)
        shutil.rmtree(self.juju_home)
        super(JujuFakeTest, self).tearDown()

    def test_info(self):
        info = self.environment.info()
        self.assertEqual("dummy", info["ProviderType"])

    def test_local_charm(self):
        charm = self.environment.add_local_charm_dir(DUMMY_CHARM, "trusty")
        self.environment.deploy("dummy", charm["CharmURL"], num_units=0)

    def test_run_on_all_machines(self):
        timeout = 5 * 10 ** 9
        result = self.environment.run_on_all_machines(
            "/foo/bar", timeout=timeout)
        self.assertEqual(
            {"Results": [
                {"Code": 0,
                 "Stdout": "",
                 "Stderr": "",
                 "MachineId": "0",
                 "Error": "",
                 "UnitId": ""}]},
            result)
