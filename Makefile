ifeq ($(GOPATH),)
GOPATH=$(shell pwd)
endif

PROTOC_ARCH=$(shell uname -sm | tr '[:upper:]' '[:lower:]' | sed -e 's/darwin/osx/g' -e 's/ /-/g')
PROTOC_VERSION=3.7.1
PROTOC_ZIP=https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(PROTOC_ARCH).zip
PROTO_BASE=$(GOPATH)/src/github.com/ciena/voltctl/voltha-protos@v1
PROTOC=protoc
PROTOC_INCLUDES=-I $(PROTO_BASE)/v1/omci -I $(PROTO_BASE)/v1/voltha -I $(PROTO_BASE)/v1/ietf -I $(PROTO_BASE)/v1/bbf_fiber -I $(PROTO_BASE)/v1/openflow_13 -I /usr/local/include -I $(PROTO_BASE)/common -I $(PROTO_BASE)/v1/common -I $(PROTO_BASE)/v1/ietf -I $(PROTO_BASE)/v1
PROTOC_OUT_DIR=$(GOPATH)/src/github.com/ciena/voltctl/vendor
PROTOC_OUT=--go_out=plugins=grpc:$(PROTOC_OUT_DIR)

help:

protoc:
	@mkdir -p .work_dir
	wget -O .work_dir/protoc-3.7.1.zip $(PROTOC_ZIP)
	unzip -o .work_dir/protoc-3.7.1.zip -d .work_dir
	sudo cp -r .work_dir/include/* /usr/local/include
	sudo cp .work_dir/bin/protoc /usr/local/bin
	GOPATH=$(GOPATH) go get -u github.com/golang/protobuf/protoc-gen-go
	sudo cp bin/protoc-gen-go /usr/local/bin
	@rm -rf .work_dir

protos:
	@rm -rf $(PROTOC_OUT_DIR)/github.com/opencord/voltha-protos/go
	@mkdir -p $(PROTOC_OUT_DIR)
	$(PROTOC) $(PROTOC_INCLUDES) $(PROTOC_OUT) $(PROTO_BASE)/common/google/api/*.proto
	$(PROTOC) $(PROTOC_INCLUDES) $(PROTOC_OUT) $(PROTO_BASE)/v1/common/*.proto
	$(PROTOC) $(PROTOC_INCLUDES) $(PROTOC_OUT) $(PROTO_BASE)/v1/ietf/*.proto
	$(PROTOC) $(PROTOC_INCLUDES) $(PROTOC_OUT) $(PROTO_BASE)/v1/openflow_13/*.proto
	$(PROTOC) $(PROTOC_INCLUDES) $(PROTOC_OUT) $(PROTO_BASE)/v1/schema/*.proto
	$(PROTOC) $(PROTOC_INCLUDES) $(PROTOC_OUT) $(PROTO_BASE)/v1/omci/*.proto
	$(PROTOC) $(PROTOC_INCLUDES) $(PROTOC_OUT) $(PROTO_BASE)/v1/bbf_fiber/*.proto
	$(PROTOC) $(PROTOC_INCLUDES) $(PROTOC_OUT) $(PROTO_BASE)/v1/voltha/*.proto

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
	hexdump -ve '1/1 "0x%02x,"' protosets/voltha_v1.pb | fold -w 60 -s >> $@
	@echo "}" >> $@
	@go fmt $@

encode-protosets: commands/voltha_v1_pb.go commands/voltha_v2_pb.go

build:
	GOPATH=$(GOPATH) go install github.com/ciena/voltctl
