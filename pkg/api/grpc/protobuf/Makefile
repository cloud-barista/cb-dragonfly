
regenerate:
	# gogo proto compile
	protoc \
			./gogoproto/gogo.proto \
			-I . \
			-I $(GOPATH)/src/github.com/gogo/protobuf/protobuf \
			-I $(GOPATH)/src \
			--gofast_out=plugins=grpc,paths=source_relative,\
	Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:\
	.

# 1. git clone https://github.com/gogo/protobuf at the same level as the cloud-barista folder
# 2. Run below command
regenerate-df:
	# cbdragonfly proto compile
	protoc \
			./cbdragonfly/cbdragonfly.proto \
			-I . \
			-I /home/snchoi/project/src/cloud-barista/src/github.com/protobuf/protobuf \
			-I /home/snchoi/project/src/cloud-barista/src/github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/protobuf \
			--go_out=plugins=grpc:\
	.
