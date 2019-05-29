ifeq ($(GOPATH),)
$(error "***** Please set your GOPATH environment variable")
endif

ifneq ($(GOPATH)/src/github.com/ciena/voltctl,$(shell pwd))
$(warning "***** Your GOPATH environment variable may not be set correctly. Your current directory should be $$GOPATH/src/github.com/ciena/voltctl")
endif

help:

commands/voltha_v1_pb.go: protosets/voltha_v1.pb
	@echo "package commands" > $@
	@echo "" >> $@
	@echo "var V1Descriptor = []byte{" >> $@
	hexdump -ve '1/1 "0x%02x,"' protosets/voltha_v1.pb | fold -w 60 -s >> $@
	@echo "}" >> $@
	@go fmt $@

commands/voltha_v2_pb.go: protosets/voltha_v2.pb
	@echo "package commands" > $@
	@echo "" >> $@
	@echo "var V2Descriptor = []byte{" >> $@
	hexdump -ve '1/1 "0x%02x,"' protosets/voltha_v2.pb | fold -w 60 -s >> $@
	@echo "}" >> $@
	@go fmt $@

encode-protosets: commands/voltha_v1_pb.go commands/voltha_v2_pb.go

VERSION=$(shell cat $(GOPATH)/src/github.com/ciena/voltctl/VERSION)
GITCOMMIT=$(shell git rev-parse HEAD)
ifeq ($(shell git ls-files --others --modified --exclude-standard 2>/dev/null | wc -l | sed -e 's/ //g'),0)
GITDIRTY=false
else
GITDIRTY=true
endif
GOVERSION=$(shell go version 2>&1 | sed -E  's/.*(go[0-9]+\.[0-9]+\.[0-9]+).*/\1/g')
HOST_OS=$(shell uname -s | tr A-Z a-z)
ifeq ($(shell uname -m),x86_64)
	HOST_ARCH ?= amd64
else
	HOST_ARCH ?= $(shell uname -m)
endif
BUILDTIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags \
	'-X "github.com/ciena/voltctl/cli/version.Version=$(VERSION)"  \
	 -X "github.com/ciena/voltctl/cli/version.VcsRef=$(GITCOMMIT)"  \
	 -X "github.com/ciena/voltctl/cli/version.VcsDirty=$(GITDIRTY)"  \
	 -X "github.com/ciena/voltctl/cli/version.GoVersion=$(GOVERSION)"  \
	 -X "github.com/ciena/voltctl/cli/version.Os=$(HOST_OS)" \
	 -X "github.com/ciena/voltctl/cli/version.Arch=$(HOST_ARCH)" \
	 -X "github.com/ciena/voltctl/cli/version.BuildTime=$(BUILDTIME)"'

# Release related items
# Generates binaries in $RELEASE_DIR with name $RELEASE_NAME-$RELEASE_OS_ARCH
# Inspired by: https://github.com/kubernetes/minikube/releases
RELEASE_DIR     ?= release
RELEASE_NAME    ?= voltctl
RELEASE_OS_ARCH ?= linux-amd64 windows-amd64 darwin-amd64
RELEASE_BINS    := $(foreach rel,$(RELEASE_OS_ARCH),$(RELEASE_DIR)/$(RELEASE_NAME)-$(subst -dev,_dev,$(VERSION))-$(rel))

# Functions to extract the OS/ARCH
rel_ver   = $(word 2, $(subst -, ,$(notdir $@)))
rel_os    = $(word 3, $(subst -, ,$(notdir $@)))
rel_arch  = $(word 4, $(subst -, ,$(notdir $@)))

dependencies:
	[ -d "vendor" ] || dep ensure

$(RELEASE_BINS): dependencies
	mkdir -p $(RELEASE_DIR)
	GOPATH=$(GOPATH) GOOS=$(rel_os) GOARCH=$(rel_arch) \
	       go build -v $(LDFLAGS) -o "$@" cmd/voltctl.go

release: $(RELEASE_BINS)

build: dependencies
	GOPATH=$(GOPATH) \
	       go build $(LDFLAGS) \
	       cmd/voltctl.go

install: dependencies
	GOPATH=$(GOPATH) GOBIN=$(GOPATH)/bin  go install $(LDFLAGS) \
	       cmd/voltctl.go

run: dependencies
	GOPATH=$(GOPATH) go run $(LDFLAGS) github.com/ciena/voltctl/cmd $(CMD) 

lint: dependencies
	GOPATH=$(GOPATH) find $(GOPATH)/src/github.com/ciena/voltctl -name "*.go" -not -path '$(GOPATH)/src/github.com/ciena/voltctl/vendor/*' | xargs gofmt -l
	GOPATH=$(GOPATH) go vet github.com/ciena/voltctl/...
	dep check

test: dependencies
	GOPATH=$(GOPATH) go test $(TEST_ARGS) -cover -coverprofile=voltctl.cp github.com/ciena/voltctl/...

view-coverage:
	GOPATH=$(GOPATH) go tool cover -html voltctl.cp

clean:
	rm -rf voltctl voltctl.cp release
