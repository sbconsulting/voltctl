ifeq ($(GOPATH),)
GOPATH=$(shell pwd)
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
GITCOMMIT=$(shell git log --pretty=format:"%h" -n 1)
GOVERSION=$(shell go version 2>&1 | sed -E  's/.*(go[0-9]+\.[0-9]+\.[0-9]+).*/\1/g')
OSTYPE=$(shell uname -s | tr A-Z a-z)
OSARCH=$(shell uname -p | tr A-Z a-z)
BUILDTIME=$(shell date -u)

LDFLAGS=-ldflags \
	'-X "github.com/ciena/voltctl/cli/version.Version=$(VERSION)"  \
	 -X "github.com/ciena/voltctl/cli/version.GitCommit=$(GITCOMMIT)"  \
	 -X "github.com/ciena/voltctl/cli/version.GoVersion=$(GOVERSION)"  \
	 -X "github.com/ciena/voltctl/cli/version.Os=$(OSTYPE)" \
	 -X "github.com/ciena/voltctl/cli/version.Arch=$(OSARCH)" \
	 -X "github.com/ciena/voltctl/cli/version.BuildTime=$(BUILDTIME)"' \

build:
	GOPATH=$(GOPATH) go build $(LDFLAGS) \
	       cmd/voltctl.go

install:
	GOPATH=$(GOPATH) GOBIN=$(GOPATH)/bin  go install $(LDFLAGS) \
	       cmd/voltctl.go

run:
	GOPATH=$(GOPATH) go run $(LDFLAGS) github.com/ciena/voltctl/cmd $(CMD) 

lint:
	GOPATH=$(GOPATH) find $(GOPATH)/src/github.com/ciena/voltctl -name "*.go" -not -path '$(GOPATH)/src/github.com/ciena/voltctl/vendor/*' | xargs gofmt -l
	GOPATH=$(GOPATH) go vet github.com/ciena/voltctl/...
	dep check

test:
	GOPATH=$(GOPATH) go test $(TEST_ARGS) -cover -coverprofile=voltctl.cp github.com/ciena/voltctl/...

view-coverage:
	GOPATH=$(GOPATH) go tool cover -html voltctl.cp

clean:
	rm -rf voltctl voltctl.cp
