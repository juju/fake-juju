GO_VERSION = 1.6
GO ?= /usr/lib/go-$(GO_VERSION)/bin/go
GO_PATH = $(CURDIR)

JUJU_VERSION = $(shell basename $(CURDIR))
JUJU_MAJOR = $(shell echo $(JUJU_VERSION) | cut -f1 -d.)
JUJU_MAJOR_MINOR = $(shell (echo $(JUJU_VERSION) | cut -f 1,2 -d . | cut -f 1 -d -))
JUJU_PATCH = juju-core.patch
JUJU_SRC = src
JUJU_UNPACKED_CLEAN = .unpacked-clean
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

	# Extract the original tarball, apply the fake-juju patch and create
	# synlinks in the original tree for the additional fake-juju sources.
	tar --strip=1 -z -xf $(JUJU_TARBALL)
	patch -p0 < $(JUJU_PATCH)
	for name in $(shell find ../common/core/ -name "*.go"); \
		do ln -s $$(pwd)/$$name $(JUJU_SRC)/$$(echo $$name | cut -d / -f 4-); \
	done

	GOPATH=$(GO_PATH) $(GO) build -v -i fake-jujud.go
	GOPATH=$(GO_PATH) $(GO) build -v -i fake-juju.go

.PHONY: unit-test
unit-test: $(JUJU_TARBALL) $(JUJU_PATCH)
	GOPATH=$(GO_PATH) $(GO) test ./service -gocheck.v

.PHONY: test
ifeq (1, $(SKIP_FUNCTIONAL_TESTS))
test:
else
test: $(JUJU_VERSION)
	cd .. && JUJU_VERSION=$(JUJU_VERSION) python3 -m unittest tests.test_fake
endif

.PHONY: install
install: $(JUJU_VERSION)
	install -D fake-jujud $(JUJU_INSTALLED)

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
	GOPATH=$(GO_PATH) $(GO) build -v fake-jujud.go
