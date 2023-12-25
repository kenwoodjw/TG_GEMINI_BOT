# Makefile for a Go project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GORUN=$(GOCMD) run

# Binary name
BINARY_NAME=tg_gemini_bot

# Build the project
build:
	$(GOBUILD) -o $(BINARY_NAME) -v

# Test the project
test:
	$(GOTEST) -v ./...

# Clean up
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run the project
run:
	$(GORUN) main.go

# Cross compilation
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-linux -v

build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-windows.exe -v

build-mac:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-mac -v

.PHONY: build test clean run build-linux build-windows build-mac