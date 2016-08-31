
import os.path

import ssl
# XXX No support for cert files in Environment._http_conn, so
# add it via monkey patching.
if hasattr(ssl, "_create_unverified_context"):
    ssl._create_default_https_context = ssl._create_unverified_context

try:
    from jujuclient.juju1.environment import Environment as Juju1Client
    from jujuclient.juju2.environment import Environment as Juju2Client
except ImportError:
    from jujuclient import Environment as Juju1Client

    class Juju2Client(object):

        def __init__(self, endpoint):
            self.endpoint = endpoint

        def login(self, username):
            raise NotImplementedError

        def info(self):
            raise NotImplementedError

        def add_local_charm_dir(self, dirname, series):
            raise NotImplementedError

        def deploy(self, name, charmURL, num_units):
            raise NotImplementedError

        def run_on_all_machines(self, command, timeout):
            raise NotImplementedError


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
        raise NotImplementedError


def connect(endpoint, version=None):
    """Return a connected client to the Juju API endpoint."""
    JujuClient = Juju2Client
    if str(version).startswith("1"):
        JujuClient = Juju1Client

    # TODO make use of the cert
    # ca_cert = os.path.join(self.juju_home, "cert.ca")
    client = JujuClient(endpoint)
    client.login("test")
    return client
