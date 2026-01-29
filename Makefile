APP_NAME := phr-tui
DIST_DIR := dist
SRC_DIR := ./cmd/cli

.PHONY: build run clean all

build:
	@mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/$(APP_NAME) $(SRC_DIR)

run: build
	./$(DIST_DIR)/$(APP_NAME) examples/

clean:
	rm -rf $(APP_NAME) $(DIST_DIR)

# Cross-compilation
all: clean linux mac-intel mac-m1 windows

linux:
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(DIST_DIR)/$(APP_NAME)-linux $(SRC_DIR)

mac-intel:
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(DIST_DIR)/$(APP_NAME)-mac-intel $(SRC_DIR)

mac-m1:
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(DIST_DIR)/$(APP_NAME)-mac-m1 $(SRC_DIR)

windows:
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(DIST_DIR)/$(APP_NAME)-windows.exe $(SRC_DIR)
