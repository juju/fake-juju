Building
========

Here's how to build the various fake-juju artifacts.

All versions
------------

To build all fake-juju versions:

  $ make build

This will produce two binaries (``fake-jujud`` and ``fake-juju``) in
each ``x.y.z`` version directory.

Specific version
----------------

Get into the version's directory and run ``make build``, e.g.:

  $ cd 2.0.2
  $ make build

Debian package
--------------

Generally speaking, this package is built and delivered to PPAs from a recipe_.

You can build it locally running::

  $ dpkg-buildpackage -b

.. _recipe: https://code.launchpad.net/fake-juju/+recipes

Developing
==========

Here follow some directions for common development tasks.

Adding a new version
--------------------

1. Edit the Makefile and add the desired version on the very first line.

2. Create a new version subdirectory, create a minimal Makefile and
   create the various symlinks by running "make init", e.g.:

   $ mkdir 2.0.0
   $ cd 2.0.0
   $ echo "include ../common/makefile.mk" > Makefile
   $ make init

Adding a new patch
------------------

Use quilt to add a new patch to Juju core code, for instance:

  $ cd 2.0.0
  $ make build
  $ quilt new my-new-hack.diff
  $ quilt add src/github.com/juju/some/source/file.go
  $ <edit src/github.com/juju/some/source/file.go>
  $ quilt refresh

At this point a new ``my-new-hack.diff`` file has been added to the
``patches/`` directory, and you can add it to git.
