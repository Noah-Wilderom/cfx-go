BIN_DIR=bin

build-example:
	@GOOS=js GOARCH=wasm go build -o $(BIN_DIR)/examples/command.wasm ./examples/command/
