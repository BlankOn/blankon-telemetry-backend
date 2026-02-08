APP_NAME := server
BUILD_DIR := .
CMD_DIR := ./cmd/server

.PHONY: build run test clean

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)

run: build
	./$(APP_NAME)

test:
	go test ./...

clean:
	rm -f $(BUILD_DIR)/$(APP_NAME)
