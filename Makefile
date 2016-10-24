
ifdef JUJU_VERSIONS
# If someone explicitly set JUJU_VERSIONS then we ignore JUJU_VERSION.
JUJU_VERSION =
endif


ifdef JUJU_VERSION  #############################
# for one version

GO_VERSION = 1.6
JUJU_TARBALL = juju-core_$(JUJU_VERSION).tar.gz
JUJU_PATCH = patches/juju-core_$(JUJU_VERSION).patch
INSTALLDIR = $(DESTDIR)/usr/bin
INSTALLED = $(INSTALLDIR)/fake-juju-$(JUJU_VERSION)

$(JUJU_VERSION)/$(JUJU_VERSION):
	case $(JUJU_VERSION) in \
		1.*) $(MAKE) build-common PATH=$$PATH JUJU_VERSION=$(JUJU_VERSION) ;;\
		2.*) $(MAKE) build-common PATH=/usr/lib/go-$(GO_VERSION)/bin:$$PATH JUJU_VERSION=$(JUJU_VERSION) ;;\
	esac

juju-core_%.tar.gz:
	case $* in \
		1.*) PROJECT="juju-core" ;;\
		2.*) PROJECT="juju" ;;\
	esac;\
	wget https://launchpad.net/$$PROJECT/$(shell (echo $* | cut -f 1 -d - | cut -f 1,2 -d .))/$*/+download/$@

.PHONY: build
build: $(JUJU_VERSION)/$(JUJU_VERSION) py-build

.PHONY: build-common
build-common: $(JUJU_TARBALL) $(JUJU_PATCH)
	rm -rf $(JUJU_VERSION)/src  # Go doesn't play nice with existing files.
	tar -C $(JUJU_VERSION) --strip=1 -z -xf $(JUJU_TARBALL)
	patch -p0 < $(JUJU_PATCH)
	cd $(JUJU_VERSION) && GOPATH=$(shell pwd)/$(JUJU_VERSION) PATH=$(PATH) go build

.PHONY: install
install: $(JUJU_VERSION)/$(JUJU_VERSION) py-install-dev
	install -D $(JUJU_VERSION)/$(JUJU_VERSION) $(INSTALLED)

.PHONY: install-dev
install-dev: $(JUJU_VERSION)/$(JUJU_VERSION) py-install-dev
	mkdir -p $(INSTALLDIR)
	ln -s --backup=existing --suffix .orig $(shell pwd)/$(JUJU_VERSION)/$(JUJU_VERSION) $(INSTALLED)

.PHONY: clean
clean: py-clean
	rm -f $(JUJU_TARBALL)
	rm -rf $(JUJU_VERSION)/src
	rm -f $(JUJU_VERSION)/$(JUJU_VERSION)
	rm -rf _trial_temp
	rm -f tests/*.pyc

.PHONY: test
test: $(JUJU_VERSION)/$(JUJU_VERSION) py-test
	env JUJU_VERSION=$(JUJU_VERSION) python3 -m unittest tests.test_fake


else  ###########################################
# for all versions

ifndef JUJU_VERSIONS
JUJU1_VERSIONS = 1.24.7 1.25.6
JUJU2_VERSIONS = 2.0-beta17
JUJU_VERSIONS = $(JUJU1_VERSIONS) $(JUJU2_VERSIONS)
endif
BUILT_VERSIONS = $(foreach version,$(JUJU_VERSIONS),$(version)/$(version))

$(BUILT_VERSIONS):
	for VERSION in $(JUJU_VERSIONS); do \
	    $(MAKE) build JUJU_VERSION=$$VERSION SKIP_PYTHON_LIB=TRUE; \
	done

.PHONY: build
build: $(BUILT_VERSIONS) py-build

.PHONY: install
install: py-install
	for VERSION in $(JUJU_VERSIONS); do \
	    $(MAKE) install JUJU_VERSION=$$VERSION SKIP_PYTHON_LIB=TRUE; \
	done

.PHONY: install-dev
install-dev: py-install-dev
	for VERSION in $(JUJU_VERSIONS); do \
	    $(MAKE) install-dev JUJU_VERSION=$$VERSION SKIP_PYTHON_LIB=TRUE; \
	done

.PHONY: clean
clean: py-clean
	for VERSION in $(JUJU_VERSIONS) ; do \
		$(MAKE) clean JUJU_VERSION=$$VERSION SKIP_PYTHON_LIB=TRUE; \
	done	

.PHONY: test
# Use xargs here so that we don't throw away the return codes, and correctly fail if any of the tests fail
test: $(BUILT_VERSIONS) py-test
	#@echo -n $(JUJU_VERSIONS) | xargs -t -d' ' -I {} env JUJU_VERSION={} python3 -m unittest tests.test_fake
	@echo -n $(JUJU_VERSIONS) | xargs -t -d' ' -I {} $(MAKE) test JUJU_VERSION={} SKIP_PYTHON_LIB=TRUE


endif  ##########################################

# for the Python library

PYTHON = python
ifndef PYTHON_INSTALLDIR
PYTHON_INSTALLDIR = $(DESTDIR)/usr/lib/python2.7/dist-packages
endif
PYTHON_INSTALL_OPTION = --install-lib $(PYTHON_INSTALLDIR)

PYTHON_LIB_ROOT = $(shell pwd)/python
# TODO: read from python/fakejuju/__init__.py
PYTHON_LIB_VERSION = 0.9.0b1
PYTHON_LIB_SOURCE_TARBALL = $(PYTHON_LIB_ROOT)/dist/fakejuju-$(PYTHON_LIB_VERSION).tar.gz

$(PYTHON_LIB_SOURCE_TARBALL): python/fakejuju
	echo $(PYTHON_LIB_SOURCE_TARBALL)
	cd python; \
	$(PYTHON) setup.py sdist

.PHONY: py-build
py-build: $(PYTHON_LIB_SOURCE_TARBALL)

.PHONY: py-install
py-install: py-build
	if [ ! "$(SKIP_PYTHON_LIB)" ]; then \
		mkdir -p $(PYTHON_INSTALL_DIR); \
		cd python; \
		$(PYTHON) setup.py install $(PYTHON_INSTALL_OPTION); \
	fi

.PHONY: py-install-dev
py-install-dev:
	if [ ! "$(SKIP_PYTHON_LIB)" ]; then \
		ln -snv $(PYTHON_LIB_ROOT)/fakejuju $(PYTHON_INSTALLDIR)/fakejuju; \
	fi

.PHONY: py-clean
py-clean:
	if [ ! "$(SKIP_PYTHON_LIB)" ]; then \
		rm -rf python/dist; \
		rm -rf python/fakejuju.egg-info; \
		rm -f python/fakejuju/*.pyc; \
		rm -f python/fakejuju/tests/*.pyc; \
	fi

.PHONY: py-test
py-test:
	if [ ! "$(SKIP_PYTHON_LIB)" ]; then \
		$(PYTHON) -m unittest discover -t $(PYTHON_LIB_ROOT) -s $(PYTHON_LIB_ROOT)/fakejuju; \
	fi


#################################################
# other targets

JUJUCLIENT_DOWNLOADS = $(shell pwd)/tests/jujuclient-archive
JUJUCLIENT_REQ = $(JUJUCLIENT_DOWNLOADS)/requirements

.PHONY: ci-test
ci-test:
	sudo apt-get -y install --force-yes \
		wget \
		python3-pip \
		golang-go \
		golang-1.6
	# See tests/jujuclient-archive/UPGRADE when a newer jujuclient version is needed.
	sudo python3 -m pip install \
		--ignore-installed \
        --no-cache-dir \
		--no-index \
		--find-links $(JUJUCLIENT_DOWNLOADS) \
		--requirement $(JUJUCLIENT_REQ)
	$(MAKE) test
