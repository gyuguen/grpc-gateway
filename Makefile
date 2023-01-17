OUT_DIR = build
OUT_BIN = echo
PROTO_DIR = proto
PROTO_OUT_DIR = pb

.PHONY: build install clean proto-gen proto-clean

build:
	go build -mod=readonly -o $(OUT_DIR)/$(OUT_BIN) .

install:
	go install -mod=readonly .

clean:
	go clean
	rm -rf $(OUT_DIR)

proto-gen:
	mkdir -p $(PROTO_OUT_DIR)
	protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(PROTO_OUT_DIR) --go_opt=paths=source_relative \
        --go-grpc_out=$(PROTO_OUT_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PROTO_OUT_DIR) \
		--grpc-gateway_opt logtostderr=true \
    	--grpc-gateway_opt paths=source_relative \
		--openapiv2_out . \
		--openapiv2_opt logtostderr=true --openapiv2_opt repeated_path_param_separator=ssv \
        $(PROTO_DIR)/echo/v1/echo.proto

proto-clean:
	rm -rf $(PROTO_OUT_DIR)