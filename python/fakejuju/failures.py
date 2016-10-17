# Copyright 2016 Canonical Limited.  All rights reserved.

import errno
import os
import os.path


class Failures(object):
    """The collection of injected failures to use with a fake-juju.

    The failures are tracked here as well as injected into any
    fake-juju using the initial config dir (aka "juju home").

    Note that fake-juju provides only limited capability for
    failure injection.
    """

    def __init__(self, cfgdir, entities=None):
        """
        @param cfgdir: The "juju home" directory into which the
            failures will be registered for injection.
        @param entities: The entity names to start with, if any.
        """
        filename = os.path.join(cfgdir, "juju-failures")
        entities = set(entities or ())

        self._filename = filename
        self._entities = entities

    @property
    def filename(self):
        """The path to the failures file the fake-juju reads."""
        return self._filename

    @property
    def entities(self):
        """The IDs of the failing entities."""
        return set(self._entities)

    def _flush(self):
        """Write the failures to disk."""
        data = "\n".join(self._entities) + "\n"
        try:
            file = open(self._filename, "w")
        except IOError:
            dirname = os.path.dirname(self._filename)
            if not os.path.exists(dirname):
                os.makedirs(dirname)
            file = open(self._filename, "w")
        with file:
            file.write(data)

    def fail_entity(self, tag):
        """Inject a global failure for the identified Juju entity."""
        self._entities.add(tag)
        self._flush()

    def clear(self):
        """Remove all injected failures."""
        try:
            os.remove(self._filename)
        except OSError as e:
            if e.errno != errno.ENOENT:
                raise
        self._entities.clear()
