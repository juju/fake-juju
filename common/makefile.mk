GO_VERSION = 1.6
GO ?= /usr/lib/go-$(GO_VERSION)/bin/go
GO_PATH = $(CURDIR)

JUJU_VERSION = $(shell basename $(CURDIR))
JUJU_MAJOR = $(shell echo $(JUJU_VERSION) | cut -f1 -d.)
JUJU_MAJOR_MINOR = $(shell (echo $(JUJU_VERSION) | cut -f 1,2 -d . | cut -f 1 -d -))
JUJU_SRC = src

SERVER = fake-jujud-$(JUJU_VERSION)
CLIENT = fake-juju-$(JUJU_VERSION)

ifeq (1, $(JUJU_MAJOR))
  JUJU_PROJECT=juju-core
else
  JUJU_PROJECT=juju
endif

JUJU_TARBALL = juju-core_$(JUJU_VERSION).tar.gz
JUJU_TARBALL_URL=https://launchpad.net/$(JUJU_PROJECT)/$(JUJU_MAJOR_MINOR)/$(JUJU_VERSION)/+download/$(JUJU_TARBALL)

.PHONY: init
init:
	ln -s ../common/patches .
	ln -s ../common/service .
	ln -s ../common/fake-jujud.go $(SERVER).go
	ln -s ../common/fake-juju.go $(CLIENT).go

.PHONY: build
build: overlay patch
	GOPATH=$(GO_PATH) $(GO) build -v -i $(SERVER).go
	GOPATH=$(GO_PATH) $(GO) build -v -i $(CLIENT).go

.PHONY: unit-test
unit-test: overlay patch
	GOPATH=$(GO_PATH) $(GO) test ./service -timeout 5m -gocheck.vv

.PHONY: test
ifeq (1, $(SKIP_FUNCTIONAL_TESTS))
test:
else
test: $(JUJU_VERSION)
	cd .. && JUJU_VERSION=$(JUJU_VERSION) python3 -m unittest tests.test_fake
endif

.PHONY: install
install:
	install -D $(SERVER) $(DESTDIR)/usr/bin/$(SERVER)
	install -D $(CLIENT) $(DESTDIR)/usr/bin/$(CLIENT)

# Copy fake-juju Go code extensions to the upstream source tree
.PHONY: overlay
overlay: src
	# First clean up the upstream source tree from any overlay symlink
	# we might have created in previous runs.
	find src/ -name *-fakejuju.go -delete

	# Then create new symlinks against our overlay Go source files.
	for name in $(shell find ../common/core/ -name "*.go"); \
		do ln -s $$(pwd)/$$name src/$$(echo $$name | sed -e "s|^../common/core/||"); \
	done

# Apply quilt patches to the upstream source tree
.PHONY: patch
patch: src
	quilt --quiltrc=/dev/null applied > /dev/null 2>&1 || quilt --quiltrc=/dev/null push -a

# Extract the upstream tarball
.PHONY: src
src: $(JUJU_TARBALL)
	test -e src || tar --strip=1 -z -xf $(JUJU_TARBALL)

# Download the upstream tarball
$(JUJU_TARBALL):
	wget $(JUJU_TARBALL_URL)

.PHONY: clean
clean:
	rm -f $(JUJU_TARBALL)
	rm -rf src
	rm -rf .pc
	rm -rf pkg
	rm -f $(SERVER)
	rm -f $(CLIENT)
