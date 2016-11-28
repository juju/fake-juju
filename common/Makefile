GO_VERSION = 1.6
GO_PATH = $(shell pwd)

JUJU_VERSION = $(shell basename $(shell pwd))
JUJU_MAJOR = $(shell echo $(JUJU_VERSION) | cut -f1 -d.)
JUJU_MAJOR_MINOR = $(shell (echo $(JUJU_VERSION) | cut -f 1,2 -d .))
JUJU_PATCH = juju-core.patch
JUJU_SRC = $(GO_PATH)/src
JUJU_UNPACKED_CLEAN = $(GO_PATH)/.unpacked-clean
JUJU_INSTALLDIR = $(DESTDIR)/usr/bin
JUJU_INSTALLED = $(JUJU_INSTALLDIR)/fake-juju-$(JUJU_VERSION)

ifeq (1, $(JUJU_MAJOR))
  JUJU_PROJECT=juju-core
else
  JUJU_PROJECT=juju
endif

JUJU_TARBALL = juju-core_$(JUJU_VERSION).tar.gz
JUJU_TARBALL_URL=https://launchpad.net/$(JUJU_PROJECT)/$(JUJU_MAJOR_MINOR)/$(JUJU_VERSION)/+download/$(JUJU_TARBALL)

.PHONY: build
build: $(JUJU_TARBALL) $(JUJU_PATCH)
	rm -rf $(JUJU_SRC)  # Go doesn't play nice with existing files.
	tar --strip=1 -z -xf $(JUJU_TARBALL)
	patch -p0 < $(JUJU_PATCH)
	GOPATH=$(GO_PATH) PATH=$(PATH) go build

.PHONY: test
test: $(JUJU_VERSION)
	cd .. && JUJU_VERSION=$(JUJU_VERSION) python3 -m unittest tests.test_fake

.PHONY: install
install: $(JUJU_VERSION)
	install -D $(JUJU_VERSION) $(JUJU_INSTALLED)

.PHONY: clean
clean:
	rm -f $(JUJU_TARBALL)
	rm -rf $(JUJU_SRC)
	rm -f $(JUJU_VERSION)

.PHONY: update-patch
update-patch: $(JUJU_SRC) $(JUJU_UNPACKED_CLEAN)
	diff -U 3 -r --no-dereference $(JUJU_UNPACKED_CLEAN) $(JUJU_SRC) > $(JUJU_PATCH); \
		echo " -- diff exited with $$? --"
	rm -rf $(JUJU_UNPACKED_CLEAN)

$(JUJU_TARBALL):
	wget $(JUJU_TARBALL_URL)

$(JUJU_UNPACKED_CLEAN): $(JUJU_TARBALL)
	mkdir -p $(JUJU_UNPACKED_CLEAN)
	tar -C $(JUJU_UNPACKED_CLEAN) --strip=2 -z -xf $(JUJU_TARBALL)

$(JUJU_VERSION):
	GOPATH=$(GO_PATH) PATH=$(PATH) go build
