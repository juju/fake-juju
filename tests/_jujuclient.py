
import os.path
import ssl
# XXX No support for cert files in Environment._http_conn, so
# add it via monkey patching.
if hasattr(ssl, "_create_unverified_context"):
    ssl._create_default_https_context = ssl._create_unverified_context
import sys

try:
    from jujuclient.juju1.environment import Environment as Juju1Client
    from jujuclient.juju2.environment import Environment as Juju2Client
except ImportError:
    sys.exit("latest jujuclient not installed "
             "(try python -m pip install jujuclient)")


ENVIRONMENTS_YAML = """environments:
    {name}:
        admin-secret: test
        default-series: trusty
        type: {type}
"""


def prepare(name, type, cfgdir, env, version=None):

    if str(version).startswith("1"):
        environments_yaml = os.path.join(cfgdir, "environments.yaml")
        with open(environments_yaml, "w") as fd:
            fd.write(ENVIRONMENTS_YAML.format(**locals()))
        env["JUJU_HOME"] = cfgdir
    else:
        env["JUJU_DATA"] = cfgdir


def connect(endpoint, password, uuid, version=None):
    """Return a connected client to the Juju API endpoint."""
    JujuClient = Juju2Client
    midpath = "model"
    if str(version).startswith("1."):
        JujuClient = Juju1Client
        midpath = "environment"
    endpoint = "/".join(["wss:/", endpoint, midpath, uuid, "api"])

    # TODO make use of the cert
    # ca_cert = os.path.join(self.juju_home, "cert.ca")
    client = JujuClient(endpoint)
    client.login(password)
    return client


class Juju2Client(Juju2Client):
    """Fixes busted code."""

    def run_on_all_machines(self, command, timeout=None):
        """Run the given shell command on all machines in the environment."""
        return self._rpc({
            "type": self.actions.name,
            "version": self.actions.version,
            "request": "RunOnAllMachines",
            "params": {"commands": command,
                       "timeout": timeout}})
