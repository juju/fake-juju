Code organization
=================

The repository keeps multiple versions of fake-juju, each building
against the relevant real-juju code.

In this way, fake-juju can be used to write integration tests for
exercising code meant to support multiple Juju versions, typically
with different behavior associated with each version.

The ``common`` directory
------------------------

All fake-juju code that is common across all versions is kept under
the ``common/`` directory. In particular look at the following
entries:

patches/
  A series of quilt-managed patches, each introducing a change in
  the regular Juju core code. The rational for the change is described
  in the header of the file.

core/
  An "overlay" GOPATH directory tree, that will add fake-juju extensions
  to Juju's core code tree. All files in this directory are named using
  the pattern ``<real-file-name>-fakejuju.go``. The idea is to avoid
  patching real code as much as possible, and put extra code in these
  standalone files, that won't conflict with the original code.

service/
  The Go packages used to build the fake-jujud binary.

fake-jujud.go
  Main Go file for building the fake-jujud binary.

fake-juju.go
  Main Go file for building the fake-juju binary.

The ``x.y.z`` directories
-------------------------

These directories hold version-specific code (if any) and are used as
workspace for extracting the original Juju source tree of a particular
version, applying the patches and build the relevant fake-juju
binaries.

Their version-controlled files are typically symlinks to files under
the ``common/`` directory.

The ``python`` directory
------------------------

This is where Python code for driving fake-juju lives, as well as
integration tests for fake-juju itself.
