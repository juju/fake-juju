.. image:: https://launchpadlibrarian.net/248604805/FakeJuju64x64.png
    :align: right
    :alt: Project logo

Fake juju
=========

.. image:: https://travis-ci.org/juju/fake-juju.svg?branch=master
    :target: https://travis-ci.org/juju/fake-juju
    :alt: Build Status

This package implements a fake Juju controller/cli. It's very close to the
"real" Juju, with the main difference being that it runs with the dummy
provider as backend.

It is meant as a helper in integration tests for services that consume Juju
in some way (typically by talking to its websockets API).

Adding a new version
---------------------

1) Edit the Makefile and add the desired version on the very first line.
2) Create a new version subdirectory, copy fake-jujud.go and other files
   there (e.g. Makefile and juju-core.patch)

Dependencies
------------

To run tests, the code will need the latest python-jujuclient and python-txjuju
packages installed. These builds are available from the juju-stable and
txjuju-daily PPAs and can be added with::

  sudo add-apt-repository -y ppa:juju/stable
  sudo add-apt-repository -y ppa:landscape/txjuju-daily
  sudo apt-get update && sudo apt-get install python-jujuclient python-txjuju

Building
---------

Generally speaking, this package is built and delivered to PPAs from a recipe_.

You can build it locally running::

  $ dpkg-buildpackage -b

.. _recipe: https://code.launchpad.net/fake-juju/+recipes